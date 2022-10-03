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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	messagesv1alpha1 "github.com/kubbee/ccloud-messaging-topology-operator/api/v1alpha1"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
	"github.com/kubbee/ccloud-messaging-topology-operator/services"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KafkaServiceAccountReconciler reconciles a KafkaServiceAccount object
type KafkaServiceAccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkaserviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkaserviceaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkaserviceaccounts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KafkaServiceAccount object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *KafkaServiceAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	kafkaServiceAccount := &messagesv1alpha1.KafkaServiceAccount{}

	if err := r.Get(ctx, req.NamespacedName, kafkaServiceAccount); err != nil {
		if k8sErrors.IsNotFound(err) {
			logger.Info("KafkaServiceAccount Noit Found.")

			if !kafkaServiceAccount.ObjectMeta.DeletionTimestamp.IsZero() {
				logger.Info("KafkaServiceAccount was marked for deletion.")
				return reconcile.Result{}, nil
			}
		}
	}

	// call the func that creates the service account
	return r.declareServiceAccount(ctx, req, kafkaServiceAccount)
}

func (r *KafkaServiceAccountReconciler) declareServiceAccount(ctx context.Context, req ctrl.Request, sa *messagesv1alpha1.KafkaServiceAccount) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Call func BuildServiceAccnount")

	if serviceAccount, err := services.BuildServiceAccount(sa.Spec.ServiceAccount, sa.Spec.Description, &logger); err != nil {
		return ctrl.Result{}, err
	} else {
		secret, _ := r.declareSecret(serviceAccount, req.Namespace)

		//
		if err := r.Create(ctx, secret); err != nil {
			//TODO implements a rollback action
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *KafkaServiceAccountReconciler) declareSecret(serviceAccount *util.ServiceAccount, namespace string) (*corev1.Secret, error) {

	var labels = make(map[string]string)
	labels["name"] = serviceAccount.Name
	labels["id"] = serviceAccount.Id
	labels["owner"] = "ccloud-messaging-topology-operator"
	labels["controller"] = "kafkaserviceaccount_controller"

	var immutable bool = true

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccount.Name,
			Namespace: namespace,
			Labels:    labels,
		},
		Type:      "kubbee.tech/secret",
		Data:      map[string][]byte{"ServiceAccount": []byte(serviceAccount.Id)},
		Immutable: &immutable,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagesv1alpha1.KafkaServiceAccount{}).
		Complete(r)
}
