/*
Copyright 2025.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CarsSpec defines the desired state of Cars
type CarsSpec struct {
	Image            string                         `json:"image,omitempty"`
	StorageClass     string                         `json:"storageClass,omitempty"`
	StorageResources *v1.VolumeResourceRequirements `json:"storageResources,omitempty"`
	StorageVolume    string                         `json:"storageVolume,omitempty"`
	Domain           string                         `json:"domain,omitempty"`
	ClusterIssuer    string                         `json:"clusterIssuer,omitempty"`
}

// CarsStatus defines the observed state of Cars
type CarsStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cars is the Schema for the cars API
type Cars struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CarsSpec   `json:"spec,omitempty"`
	Status CarsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CarsList contains a list of Cars
type CarsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cars `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cars{}, &CarsList{})
}
