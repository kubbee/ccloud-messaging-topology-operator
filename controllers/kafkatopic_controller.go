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
	"fmt"
	"strconv"

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
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if !KafkaTopic.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Deleting")
		return ctrl.Result{}, nil // implementing the nil in the future
	}

	logger.Info(">Calls declareKafkaTopic<")

	return r.declareKafkaTopic(ctx, req, KafkaTopic)
}

func (r *KafkaTopicReconciler) declareKafkaTopic(ctx context.Context, req ctrl.Request, kafkaTopic *messagesv1alpha1.KafkaTopic) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	if connectionCreds, cErr := r.getSecret(ctx, req.NamespacedName.Namespace, "kafka-reference-connection"); cErr != nil {
		logger.Error(cErr, "Failed to get Secret")
		return reconcile.Result{}, cErr
	} else {

		cf := &corev1.ConfigMap{}
		getError := r.Get(ctx, types.NamespacedName{Name: kafkaTopic.Name, Namespace: kafkaTopic.Namespace}, cf)

		if getError != nil && k8sErrors.IsNotFound(getError) { //Create ConfigMap

			logger.Info("Trying to create kafka topic")

			tenant, isTenantOk := connectionCreds.Data("tenant")
			clusterId, isClusterIdOk := connectionCreds.Data("clusterId")
			environmentId, isEnvironmentIdOk := connectionCreds.Data("environmentId")

			if isTenantOk && isClusterIdOk && isEnvironmentIdOk {

				logger.Info("Success to get secret values")

				clusterReference := util.ClusterReference{
					ClusterId:     string(clusterId),
					EnvironmentId: string(environmentId),
				}

				Partitions := strconv.FormatInt(int64(kafkaTopic.Spec.Partitions), 10)
				logger.Info("Converting the partitions was converted to string? " + Partitions)

				newTopic := &util.NewTopic{
					Tenant:           string(tenant),
					Topic:            kafkaTopic.Spec.TopicName,
					Partitions:       Partitions,
					Namespace:        req.NamespacedName.Namespace,
					ClusterReference: clusterReference,
				}

				if topicRefence, topicError := services.GetTopic(newTopic, &logger); topicError != nil {
					return reconcile.Result{}, topicError
				} else {
					cfg := r.createConfigMap(topicRefence, kafkaTopic)

					ccmError := r.Create(ctx, cfg)

					if ccmError != nil {
						logger.Error(ccmError, "error to create configmap")
						return reconcile.Result{}, ccmError
					}

					logger.Info("ConfigMap created", "Name", "Namespace", kafkaTopic.Name, kafkaTopic.Namespace)
					return reconcile.Result{}, nil
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *KafkaTopicReconciler) getSecret(ctx context.Context, requestNamespace string, secretName string) (ConnectionCredentials, error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("getting secret")
	secret := &corev1.Secret{}

	if err := r.Get(ctx, types.NamespacedName{Namespace: requestNamespace, Name: secretName}, secret); err != nil {
		return nil, err
	}

	return readCredentialsFromKubernetesSecret(secret)
}

func (r *KafkaTopicReconciler) createConfigMap(topicReference *util.TopicReference, kt *messagesv1alpha1.KafkaTopic) *corev1.ConfigMap {

	var labels = make(map[string]string)
	labels["name"] = kt.Name
	labels["owner"] = "ccloud-messaging-topology-operator"
	labels["controller"] = "kafkatopic_controller"

	var immutable bool = true

	mapper := make(map[string]string)

	mapper["topic.name"] = topicReference.Topic
	mapper["schema_registry.url"] = topicReference.SchemaRegistry.EndpointUrl
	mapper["schema_registry.api"] = topicReference.SchemaRegistryApiKey.Api
	mapper["schema_registry.secret"] = topicReference.SchemaRegistryApiKey.Secret
	mapper["kafka.url"] = topicReference.ClusterKafka.Endpoint
	mapper["kafka.api"] = topicReference.KafkaApiKey.Api
	mapper["kafka.secret"] = topicReference.KafkaApiKey.Secret

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kt.Name,
			Namespace: kt.Namespace,
			Labels:    labels,
		},
		Data:      mapper,
		Immutable: &immutable,
	}
}

func readCredentialsFromKubernetesSecret(secret *corev1.Secret) (ConnectionCredentials, error) {
	if secret == nil {
		return nil, fmt.Errorf("unable to retrieve information from Kubernetes secret %s: %w", secret.Name, errors.New("nil secret"))
	}

	return ClusterCredentials{
		data: map[string][]byte{
			"tenant":        secret.Data["tenant"],
			"clusterId":     secret.Data["clusterId"],
			"environmentId": secret.Data["environmentId"],
		},
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagesv1alpha1.KafkaTopic{}).
		Complete(r)
}
