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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	messagesv1alpha1 "github.com/kubbee/ccloud-messaging-topology-operator/api/v1alpha1"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
	"github.com/kubbee/ccloud-messaging-topology-operator/services"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KafkaReferenceReconciler reconciles a KafkaReference object
type KafkaReferenceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkareferences,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkareferences/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkareferences/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KafkaReference object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *KafkaReferenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	//
	kafkaReference := &messagesv1alpha1.KafkaReference{}

	if err := r.Get(ctx, req.NamespacedName, kafkaReference); err != nil {
		if k8sErrors.IsNotFound(err) {
			logger.Info("KafkaReference Not Found.")

			if !kafkaReference.ObjectMeta.DeletionTimestamp.IsZero() {
				logger.Info("Was marked for deletion.")
				return reconcile.Result{}, nil // implementing the nil in the future
			}
		}
		return reconcile.Result{}, nil
	}

	return r.declareClusterReference(ctx, kafkaReference)
}

func (r *KafkaReferenceReconciler) declareClusterReference(ctx context.Context, ccloudKafkaReference *messagesv1alpha1.KafkaReference) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Start::declareClusterReference")

	/*
	 * Names where are deploy the CRD of Environment, Kafka and SchemaRegistry
	 */
	var clusterResources string

	if ccloudKafkaReference.Spec.KafkaClusterResource.Namespace == "" {
		clusterResources = "ccloud-cluster-operator-system"
	} else {
		clusterResources = ccloudKafkaReference.Spec.KafkaClusterResource.Namespace
	}

	if connectionCreds, e := r.readCredentials(ctx, clusterResources, ccloudKafkaReference.Spec.Environment); e != nil {
		logger.Error(e, "Error to read environment secret into namespace: "+clusterResources)
	} else {

		environmentId, eIdOk := connectionCreds.Data("environmentId")

		logger.Info("environmentId >>>>>>>> " + string(environmentId))

		if eIdOk {
			secret := &corev1.Secret{}

			err := r.Get(ctx, types.NamespacedName{Name: ccloudKafkaReference.Name, Namespace: ccloudKafkaReference.Namespace}, secret)

			if err != nil && k8sErrors.IsNotFound(err) {

				logger.Info("call the method BuildKafkaReference")

				if clusterId, gcrError := services.BuildKafkaReference(string(environmentId), ccloudKafkaReference.Spec.ClusterName, &logger); gcrError == nil {

					logger.Info("The clusterID is >>>>> " + clusterId)

					kafkaReferenceSecret := &util.KafkaReferenceSecret{
						EnvironmentId: string(environmentId),
						Environment:   ccloudKafkaReference.Spec.Environment,
						Tenant:        ccloudKafkaReference.Spec.Tenant,
						ClusterId:     clusterId,
					}

					err = r.Create(ctx, r.generateSecret(kafkaReferenceSecret, ccloudKafkaReference))

					if err != nil {
						logger.Error(err, "error to create cluster reference secret")
						return reconcile.Result{}, err
					}

					return reconcile.Result{}, nil

				} else {
					logger.Error(gcrError, "Error to get Clueter Reference")
					return reconcile.Result{}, gcrError
				}
			}
		}
	}

	return reconcile.Result{}, nil
}

func (r *KafkaReferenceReconciler) generateSecret(krs *util.KafkaReferenceSecret, kafkaReference *messagesv1alpha1.KafkaReference) *corev1.Secret {

	var labels = make(map[string]string)
	labels["name"] = kafkaReference.Name
	labels["owner"] = "ccloud-messaging-topology-operator"
	labels["controller"] = "kafkareference_controller"

	var immutable bool = true

	// create and return secret object.
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafkaReference.Name,
			Namespace: kafkaReference.Namespace,
			Labels:    labels,
		},
		Type:      "kubbee.tech/cluster-connection-reference",
		Data:      map[string][]byte{"tenant": []byte(krs.Tenant), "clusterId": []byte(krs.ClusterId), "environmentId": []byte(krs.EnvironmentId), "environment": []byte(krs.Environment)},
		Immutable: &immutable,
	}
}

/**
 * This function is responsible to get the Environment Secret on the namespace;
 */
func (r *KafkaReferenceReconciler) readCredentials(ctx context.Context, requestNamespace string, secretName string) (util.ConnectionCredentials, error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Read credentials from cluster")

	secret := &corev1.Secret{}

	if err := r.Get(ctx, types.NamespacedName{Namespace: requestNamespace, Name: secretName}, secret); err != nil {
		return nil, err
	}

	return r.readCredentialsFromKubernetesSecret(secret), nil
}

/**
 * This function is responsible to ready the content of the secret and return for execute operations with the data
 */
func (r *KafkaReferenceReconciler) readCredentialsFromKubernetesSecret(secret *corev1.Secret) *util.ClusterCredentials {
	return &util.ClusterCredentials{
		DataContent: map[string][]byte{
			"environmentName": secret.Data["environmentName"],
			"environmentId":   secret.Data["environmentId"],
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaReferenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagesv1alpha1.KafkaReference{}).
		Complete(r)
}
