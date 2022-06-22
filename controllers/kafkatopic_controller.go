/*
Copyright 2022 Kubbee Tech.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"strconv"
	"strings"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
	"github.com/kubbee/ccloud-messaging-topology-operator/services"

	messagesv1alpha1 "github.com/kubbee/ccloud-messaging-topology-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KafkaTopicReconciler reconciles a KafkaTopic object
type KafkaTopicReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

type ConnectionCredentials interface {
	Data(key string) ([]byte, bool)
}

type ClusterCredentials struct {
	data map[string][]byte
}

func (c ClusterCredentials) Data(key string) ([]byte, bool) {
	result, ok := c.data[key]
	return result, ok
}

//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkatopics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkatopics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkatopics/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KafkaTopic object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *KafkaTopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	KafkaTopic := &messagesv1alpha1.KafkaTopic{}

	if err := r.Get(ctx, req.NamespacedName, KafkaTopic); err != nil {
		if k8sErrors.IsNotFound(err) {
			logger.Info("KafkaTopic Not Found.")

			if !KafkaTopic.ObjectMeta.DeletionTimestamp.IsZero() {
				logger.Info("Was marked for deletion.")
				return reconcile.Result{}, nil // implementing the nil in the future
			}
		}
		return reconcile.Result{}, nil
	}

	return r.declareKafkaTopic(ctx, req, KafkaTopic)
}

func (r *KafkaTopicReconciler) declareKafkaTopic(ctx context.Context, req ctrl.Request, kafkaTopic *messagesv1alpha1.KafkaTopic) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	if connCreds, e := r.readCredentials(ctx, req.NamespacedName.Namespace, kafkaTopic.Spec.KafkaReferenceResource.Name); e != nil {
		logger.Error(e, "Failed to get Secret")
		return reconcile.Result{}, e
	} else {

		cf := &corev1.ConfigMap{}
		getError := r.Get(ctx, types.NamespacedName{Name: kafkaTopic.Name, Namespace: kafkaTopic.Namespace}, cf)

		if getError != nil && k8sErrors.IsNotFound(getError) { //Create ConfigMap
			logger.Info("Creating kafka topic")

			// Read secret attributes
			tenant, isTenantOk := connCreds.Data("tenant")
			clusterId, isClusterIdOk := connCreds.Data("clusterId")
			environmentId, isEnvironmentIdOk := connCreds.Data("environmentId")

			// verify if secrets is ok
			if isTenantOk && isClusterIdOk && isEnvironmentIdOk {

				logger.Info("Success to get secret values")

				Partitions := strconv.FormatInt(int64(kafkaTopic.Spec.Partitions), 10)

				newTopic := &util.NewTopic{
					Tenant:     string(tenant),
					Topic:      kafkaTopic.Spec.TopicName,
					Partitions: Partitions,
					Namespace:  req.NamespacedName.Namespace,
				}

				if topicName, err := services.BuildTopic(newTopic, string(environmentId), string(clusterId), &logger); err != nil {
					logger.Error(err, "error to create topic")
					return reconcile.Result{}, err
				} else {

					kafkaSecretName := "kafka-" + string(tenant)
					connCredsKafka, kErr := r.readCredentialsKafka(ctx, kafkaTopic.Spec.KafkaClusterResource.Namespace, kafkaSecretName)

					schemaRegistryName := "schemaregistry-" + string(tenant)
					connCredsSR, SRErr := r.readCredentialsSchemaRegistry(ctx, kafkaTopic.Spec.KafkaClusterResource.Namespace, schemaRegistryName)

					if kErr != nil && SRErr != nil {
						logger.Error(kErr, "error to read kafka secret")
						logger.Error(SRErr, "error to read schemaregistry secret")
						return reconcile.Result{}, errors.New("error to read kafka or schemaregistry secret from cluster")
					}

					kafkaKey, isOk_KK := connCredsKafka.Data("ApiKey")
					kafkaSecret, isOk_KS := connCredsKafka.Data("ApiSecret")
					kafkaEndpoint, isOK_KE := connCredsKafka.Data("KafkaEndpoint")

					if !isOK_KE && !isOk_KK && !isOk_KS {
						logger.Info("error to read kafka secrets")
						return reconcile.Result{}, errors.New("error to read kafka secrets")
					}

					sRKey, isOk_SRK := connCredsSR.Data("ApiKey")
					sRSecret, isOk_SRS := connCredsSR.Data("ApiSecret")
					sREndpoint, isOk_SRE := connCredsSR.Data("SREndpoint")

					if !isOk_SRK && !isOk_SRS && !isOk_SRE {
						logger.Info("error to read schemaregistry secrets")
						return reconcile.Result{}, errors.New("error to read schemaregistry secrets")
					}

					configMapKafka := &util.ConfigMapKafka{
						TopicName:               *topicName,
						SchemaRegistryURL:       string(sREndpoint),
						SchemaRegistryApiKey:    string(sRKey),
						SchemaRegistryApiSecret: string(sRSecret),
						KafkaURL:                string(kafkaEndpoint),
						KafkaApiKey:             string(kafkaKey),
						KafkaApiSecret:          string(kafkaSecret),
					}

					cfg := createConfigMap(configMapKafka, kafkaTopic)

					if e := r.Create(ctx, cfg); e != nil {
						logger.Error(e, "error to create configmap")
						return reconcile.Result{}, e
					}

					logger.Info("success the topic was configured")
					return reconcile.Result{}, nil
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func createConfigMap(configMapKafka *util.ConfigMapKafka, kafkaTopic *messagesv1alpha1.KafkaTopic) *corev1.ConfigMap {

	var labels = make(map[string]string)
	labels["name"] = kafkaTopic.Name
	labels["owner"] = "ccloud-messaging-topology-operator"
	labels["controller"] = "kafkatopic_controller"

	var immutable bool = true

	/*
	 *reaplce name - to _
	 */
	variable := strings.ReplaceAll(kafkaTopic.Name, "-", "_")

	mapper := make(map[string]string)

	mapper[variable+"_topic_name"] = configMapKafka.TopicName
	mapper[variable+"_schema_registry_url"] = configMapKafka.SchemaRegistryURL
	mapper[variable+"_schema_registry_api"] = configMapKafka.SchemaRegistryApiKey
	mapper[variable+"_schema_registry_secret"] = configMapKafka.SchemaRegistryApiSecret
	mapper[variable+"_kafka_url"] = configMapKafka.KafkaURL
	mapper[variable+"_kafka_api"] = configMapKafka.KafkaApiKey
	mapper[variable+"_kafka_secret"] = configMapKafka.KafkaApiSecret

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafkaTopic.Name,
			Namespace: kafkaTopic.Namespace,
			Labels:    labels,
		},
		Data:      mapper,
		Immutable: &immutable,
	}
}

/*
 *
 */
func (r *KafkaTopicReconciler) readCredentials(ctx context.Context, requestNamespace string, secretName string) (util.ConnectionCredentials, error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Read credentials from cluster")

	secret := &corev1.Secret{}

	if err := r.Get(ctx, types.NamespacedName{Namespace: requestNamespace, Name: secretName}, secret); err != nil {
		return nil, err
	}

	return readCredentialsFromKubernetesSecret(secret), nil
}

/*
 *
 */
func readCredentialsFromKubernetesSecret(secret *corev1.Secret) *util.ClusterCredentials {
	return &util.ClusterCredentials{
		DataContent: map[string][]byte{
			"tenant":        secret.Data["tenant"],
			"clusterId":     secret.Data["clusterId"],
			"environmentId": secret.Data["environmentId"],
		},
	}
}

/**
 * This function is responsible to get the Environment Secret on the namespace;
 */
func (r *KafkaTopicReconciler) readCredentialsKafka(ctx context.Context, requestNamespace string, secretName string) (util.ConnectionCredentials, error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Read credentials kafka from cluster")

	secret := &corev1.Secret{}

	if err := r.Get(ctx, types.NamespacedName{Namespace: requestNamespace, Name: secretName}, secret); err != nil {
		return nil, err
	}

	return readCredentialsKafkaFromKubernetesSecret(secret), nil
}

/**
 * This function is responsible to ready the content of the secret and return for execute operations with the data
 */
func readCredentialsKafkaFromKubernetesSecret(secret *corev1.Secret) *util.ClusterCredentials {
	return &util.ClusterCredentials{
		DataContent: map[string][]byte{
			"ApiKey":        secret.Data["ApiKey"],
			"ApiSecret":     secret.Data["ApiSecret"],
			"KafkaEndpoint": secret.Data["KafkaEndpoint"],
		},
	}
}

/**
 * This function is responsible to get the Environment Secret on the namespace;
 */
func (r *KafkaTopicReconciler) readCredentialsSchemaRegistry(ctx context.Context, requestNamespace string, secretName string) (util.ConnectionCredentials, error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Read credentials schemaregistry from cluster")

	secret := &corev1.Secret{}

	if err := r.Get(ctx, types.NamespacedName{Namespace: requestNamespace, Name: secretName}, secret); err != nil {
		return nil, err
	}

	return readCredentialsSchemaRegistryFromKubernetesSecret(secret), nil
}

/**
 * This function is responsible to ready the content of the secret and return for execute operations with the data
 */
func readCredentialsSchemaRegistryFromKubernetesSecret(secret *corev1.Secret) *util.ClusterCredentials {
	return &util.ClusterCredentials{
		DataContent: map[string][]byte{
			"ApiKey":     secret.Data["ApiKey"],
			"ApiSecret":  secret.Data["ApiSecret"],
			"SREndpoint": secret.Data["SREndpoint"],
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagesv1alpha1.KafkaTopic{}).
		Complete(r)
}
