package v1

import (
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RegistrySpec defines the desired state of Registry
type RegistrySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Image           string          `json:"image"`
	Description     string          `json:"description"`
	LoginId         string          `json:"loginId"`
	LoginPassword   string          `json:"loginPassword"`
	CustomConfigYml string          `json:"customConfigYml"`
	ReplicaSet      apps.ReplicaSet `json:"replicaSet"`
	Service         v1.Service      `json:"service"`
	// PersistentVolumeClaim object `json:"persistentVolumeClaim"` [TODO]
}

// RegistryStatus defines the observed state of Registry
type RegistryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Conditions     []Condition `json:"conditions"`
	PhaseChangedAt string      `json:"phaseChangedAt"`
	Phase          string      `json:"phase"`
	Message        string      `json:"message"`
	Reason         string      `json:"reason"`
	Capacity       string      `json:"capacity"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Condition struct {
	LastTransitionTime string `json:"lastTransitionTime"`
	Message            string `json:"message"`
	Reason             string `json:"reason"`
	Status             string `json:"status"`
	Type               string `json:"type"`
}

// Registry is the Schema for the registries API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=registries,scope=Namespaced
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistrySpec   `json:"spec,omitempty"`
	Status RegistryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RegistryList contains a list of Registry
type RegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Registry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Registry{}, &RegistryList{})
}
