package v1

const AccessModeDefault = "ReadWriteMany"

type ExistPvc struct {
	// Use the pvc you have created
	PvcName string `json:"pvcName"`
}

type CreatePvc struct {
	AccessModes []AccessMode `json:"accessModes"`

	// enter the desired storage size (ex: 10Gi)
	StorageSize string `json:"storageSize"`

	StorageClassName string `json:"storageClassName"`

	// +kubebuilder:validation:Enum=Filesystem;Block
	VolumeMode string `json:"volumeMode,omitempty"`

	// Delete the pvc as well when this registry is deleted
	DeleteWithPvc bool `json:"deleteWithPvc,omitempty"`
}

// +kubebuilder:validation:Enum=ReadWriteOnce;ReadWriteMany
type AccessMode string
