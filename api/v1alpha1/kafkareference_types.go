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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KafkaReferenceSpec defines the desired state of KafkaReference
type KafkaReferenceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterName string `json:"clusterName"`
	Environment string `json:"environment"`
	Tenant      string `json:"tenant,omitempty"`
}

type KakfaClusterResources struct {
	Namespace string `json:"namespace,omitempty"`
}

// KafkaReferenceStatus defines the observed state of KafkaReference
type KafkaReferenceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ObservedGeneration int64       `json:"observedGeneration,omitempty"`
	Conditions         []Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KafkaReference is the Schema for the kafkareferences API
type KafkaReference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaReferenceSpec   `json:"spec,omitempty"`
	Status KafkaReferenceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaReferenceList contains a list of KafkaReference
type KafkaReferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaReference `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaReference{}, &KafkaReferenceList{})
}
