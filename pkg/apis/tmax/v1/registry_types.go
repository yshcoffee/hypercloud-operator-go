package v1

import (
	"google.golang.org/genproto/googleapis/type/datetime"
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
	Image                 string             `json:"image"`
	Description           string             `json:"description"`
	LoginId               string             `json:"loginId"`
	LoginPassword         string             `json:"loginPassword"`
	CustomConfigYml       string             `json:"customConfigYml"`
	DomainName            string             `json:"domainName"`
	RegistryReplicaSet    RegistryReplicaSet `json:"registryReplicaSet"`
	RegistryService       RegistryService    `json:"service"`
	PersistentVolumeClaim RegistryPVC        `json:"persistentVolumeClaim"`
}

type RegistryReplicaSet struct {
	Labels       map[string]string    `json:"labels"`
	NodeSelector map[string]string    `json:"nodeSelector"`
	Selector     metav1.LabelSelector `json:"selector"`
	Tolerations  []v1.Toleration      `json:"tolerations"`
}

type RegistryService struct {
	Ingress      Ingress      `json:"ingress"`
	LoadBalancer LoadBalancer `json:"loadBalancer"`
}

type RegistryPVC struct {
	MountPath string    `json:"mountPath"`
	Exist     ExistPvc  `json:"exist"`
	Create    CreatePvc `json:"create"`
}

// RegistryStatus defines the observed state of Registry
type RegistryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Conditions     []RegistryCondition `json:"conditions"`
	Phase          string              `json:"phase"`
	Message        string              `json:"message"`
	Reason         string              `json:"reason"`
	PhaseChangedAt datetime.DateTime   `json:"phaseChangedAt"`
	Capacity       string              `json:"capacity"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RegistryCondition struct {
	LastProbeTime      datetime.DateTime `json:"lastProbeTime"`
	LastTransitionTime datetime.DateTime `json:"lastTransitionTime"`
	Message            string            `json:"message"`
	Reason             string            `json:"reason"`
	Status             string            `json:"status"`
	Type               string            `json:"type"`
}

// Registry is the Schema for the registries API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=registries,scope=Namespaced
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistrySpec   `json:"spec,omitempty"`
	Status RegistryStatus `json:"status,omitempty"`

	OperatorStartTime string `json:"operatorStartTime"`
}

const (
	RegistryLoginUrl = CustomObjectGroup + "/registry-login-url"
	RegistryKind     = "Registry"
)

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
