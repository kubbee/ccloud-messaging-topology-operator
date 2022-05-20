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

	"github.com/go-logr/logr"
	messagesv1alpha1 "github.com/kubbee/ccloud-messaging-topology-operator/api/v1alpha1"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
	"github.com/kubbee/ccloud-messaging-topology-operator/services"
	"k8s.io/client-go/tools/record"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KafkaReferenceReconciler reconciles a KafkaReference object
type KafkaReferenceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
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

	kafkaReference := &messagesv1alpha1.KafkaReference{}

	if err := r.Get(ctx, req.NamespacedName, kafkaReference); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if !kafkaReference.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Deleting")
		return ctrl.Result{}, nil // implementing the nil in the future
	}

	logger.Info(">Calls declareClusterReference<")

	return r.declareClusterReference(ctx, kafkaReference)
}

func (r *KafkaReferenceReconciler) declareClusterReference(ctx context.Context, kr *messagesv1alpha1.KafkaReference) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	logger.Info(">start declareClusterReference<")

	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: kr.Name, Namespace: kr.Namespace}, secret)

	if err != nil && k8sErrors.IsNotFound(err) {

		if kReference, gcrError := services.GetClusterReference(kr, &logger); gcrError != nil {
			logger.Error(gcrError, "Error to get Clueter Reference")
			// build and return the error
			return reconcile.Result{}, gcrError
		} else {

			logger.Info(">runs declareClusterReference<")

			kafkaClusterSecret := &util.KafkaReferenceSecret{
				ClusterId:     kReference.ClusterId,
				EnvironmentId: kReference.EnvironmentId,
				Tenant:        kr.Spec.Tenant,
			}

			err = r.Create(ctx, r.generateSecret(kafkaClusterSecret, kr))

			if err != nil {
				logger.Error(err, "error to create cluster reference secret")
				return reconcile.Result{}, err
			}

			return reconcile.Result{}, nil
		}
	}

	return reconcile.Result{}, nil
}

func (r *KafkaReferenceReconciler) generateSecret(krs *util.KafkaReferenceSecret, kr *messagesv1alpha1.KafkaReference) *corev1.Secret {

	var labels = make(map[string]string)
	labels["name"] = kr.Name
	labels["owner"] = "ccloud-messaging-topology-operator"
	labels["controller"] = "kafkareference_controller"

	var immutable bool = true

	// create and return secret object.
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kr.Name,
			Namespace: kr.Namespace,
			Labels:    labels,
		},
		Type:      "kubbee.tech/cluster-connection-reference",
		Data:      map[string][]byte{"tenant": []byte(krs.Tenant), "clusterId": []byte(krs.ClusterId), "environmentId": []byte(krs.EnvironmentId)},
		Immutable: &immutable,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaReferenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagesv1alpha1.KafkaReference{}).
		Complete(r)
}
