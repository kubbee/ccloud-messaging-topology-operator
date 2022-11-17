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
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kubbee/ccloud-messaging-topology-operator/controllers/business"
	"github.com/kubbee/ccloud-messaging-topology-operator/cross"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
	"github.com/kubbee/ccloud-messaging-topology-operator/services"

	messagesv1alpha1 "github.com/kubbee/ccloud-messaging-topology-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// KafkaProducerReconciler reconciles a KafkaProducer object
type KafkaProducerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type ccdloudSecretType struct {
	Cluster        string
	Kafka          string
	SchemaRegistry string
}

//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkaproducers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkaproducers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkaproducers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KafkaProducer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *KafkaProducerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	kafkaProducer := &messagesv1alpha1.KafkaProducer{}

	if err := r.Get(ctx, req.NamespacedName, kafkaProducer); err != nil {
		if k8sErrors.IsNotFound(err) {
			logger.Info("KafkaTopic Not Found.")

			if !kafkaProducer.ObjectMeta.DeletionTimestamp.IsZero() {
				logger.Info("Was marked for deletion.")
				return reconcile.Result{}, nil // implementing the nil in the future
			}
		}
		return reconcile.Result{}, nil
	}

	if req.NamespacedName.Namespace != kafkaProducer.Namespace {
		return reconcile.Result{}, errors.New("the namespace declared is different of namespace request")
	}

	return r.produceTopic(ctx, req, kafkaProducer)
}

func (r *KafkaProducerReconciler) produceTopic(ctx context.Context, req ctrl.Request, kafkaProducer *messagesv1alpha1.KafkaProducer) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	logger.Info("Reading Credentials")
	if connCreds := r.readCredentials(ctx, req.NamespacedName.Namespace, kafkaProducer.Spec.KafkaReferenceResource.Name, 1); connCreds != nil {

		logger.Info("Recovered Credentials")
		// Struct configMap
		configMap := &corev1.ConfigMap{}

		if getError := r.Get(ctx, types.NamespacedName{Name: kafkaProducer.Name, Namespace: kafkaProducer.Namespace}, configMap); getError != nil {
			if k8sErrors.IsNotFound(getError) {
				logger.Info("Creating kafka topic")
				// Read secret attributes
				if tenant, x := connCreds.Data("tenant"); x {
					if clusterId, y := connCreds.Data("clusterId"); y {
						if environmentId, z := connCreds.Data("environmentId"); z {

							topic := &util.NewTopic{
								Tenant:     string(tenant),
								Topic:      kafkaProducer.Spec.Topic,
								Partitions: strconv.FormatInt(int64(kafkaProducer.Spec.Partitions), 10),
								Namespace:  req.NamespacedName.Namespace,
							}

							if topic, err := services.BuildTopic(topic, string(environmentId), string(clusterId), &logger); err != nil {
								logger.Error(err, "error to create topic")
								return reconcile.Result{}, err
							} else {
								if connCredsKafka := r.readCredentials(ctx, kafkaProducer.Spec.KafkaClusterResource.Namespace, "kafka-"+string(tenant), 2); connCredsKafka != nil {
									if connCredsSR := r.readCredentials(ctx, kafkaProducer.Spec.KafkaClusterResource.Namespace, "schemaregistry-"+string(tenant), 3); connCredsSR != nil {
										if cfg, err := business.GetConfigMap(connCredsKafka, connCredsSR, kafkaProducer.Name, kafkaProducer.Namespace, *topic); err != nil {
										} else {

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
						}
					}
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// ReadCredentials Get the credentials from namespace
func (r *KafkaProducerReconciler) readCredentials(ctx context.Context, namespace string, secretName string, secretType int) util.ConnectionCredentials {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Read credentials from cluster")

	secret := &corev1.Secret{}

	for a := 1; a <= 3; a++ {

		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: secretName}, secret); err != nil {

			time.Sleep(time.Duration(8*a) * time.Second)

			if a == 3 {
				logger.Error(err, "error to read crentials from cluster")
				return nil
			}

		} else {
			break
		}
	}

	return cross.GetKafkaCredentials(secret, secretType, &logger)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaProducerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagesv1alpha1.KafkaProducer{}).
		Complete(r)
}
