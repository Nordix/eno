/*


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

// L2ServiceSpec defines the desired state of L2Service
type L2ServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=4094
	SegmentationID uint16 `json:"SegmentationID"`
	// +kubebuilder:validation:MaxItems:=2
	Subnets     []string `json:"Subnets,omitempty"`
	PhysicalNet []string `json:"PhysicalNet"`
}

// L2ServiceStatus defines the observed state of L2Service
type L2ServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=l2services,scope=Cluster

// L2Service is the Schema for the l2services API
type L2Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   L2ServiceSpec   `json:"spec,omitempty"`
	Status L2ServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// L2ServiceList contains a list of L2Service
type L2ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []L2Service `json:"items"`
}

func init() {
	SchemeBuilder.Register(&L2Service{}, &L2ServiceList{})
}
