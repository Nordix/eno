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

// L2ServiceAttachmentSpec defines the desired state of L2ServiceAttachment
type L2ServiceAttachmentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	L2Services      []string `json:"L2Services"`
	ConnectionPoint string   `json:"ConnectionPoint"`
	// +kubebuilder:validation:Enum=kernel;dpdk
	PodInterfaceType string `json:"PodInterfaceType"`
	// +kubebuilder:validation:Enum=trunk;access;selectivetrunk
	VlanType string `json:"VlanType"`
	// +kubebuilder:validation:Enum=ovs;host-device
	Implementation string `json:"Implementation,omitempty"`
}

// L2ServiceAttachmentStatus defines the observed state of L2ServiceAttachment
type L2ServiceAttachmentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Enum=pending;ready;error;terminating;deleted
	Phase   string `json:"Phase,omitempty"`
	Message string `json:"Message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// L2ServiceAttachment is the Schema for the l2serviceattachments API
type L2ServiceAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   L2ServiceAttachmentSpec   `json:"spec,omitempty"`
	Status L2ServiceAttachmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// L2ServiceAttachmentList contains a list of L2ServiceAttachment
type L2ServiceAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []L2ServiceAttachment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&L2ServiceAttachment{}, &L2ServiceAttachmentList{})
}
