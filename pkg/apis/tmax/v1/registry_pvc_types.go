package v1

const AccessModeDefault = "ReadWriteMany"

type ExistPvc struct {
	PvcName string `json:"pvcName"`
}

type CreatePvc struct {
	AccessModes      []string `json:"accessModes"`
	StorageSize      string   `json:"storageSize"`
	StorageClassName string   `json:"storageClassName"`
	VolumeMode       string   `json:"volumeMode"`
	DeleteWithPvc    bool     `json:"deleteWithPvc"`
}
