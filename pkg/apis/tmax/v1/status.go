package v1

type Status string

const (
	StatusSucceeded = Status("Succeeded")
	StatusReady     = Status("Ready")
	StatusFailed    = Status("Failed")
	StatusRunning   = Status("Running")
	StatusPending   = Status("Pending")
	StatusSkipped   = Status("Skipped")

	ConditionTypeReplicaSet             = "ReplicaSetExist"
	ConditionTypePod                    = "PodRunning"
	ConditionTypeService                = "ServiceExist"
	ConditionTypeSecretOpaque           = "SecretOpaqueExist"
	ConditionTypeSecretDockerConfigJson = "SecretDockerConfigJsonExist"
	ConditionTypeSecretTls              = "SecretTlsExist"
	ConditionTypeIngress                = "IngressExist"
	ConditionTypePvc                    = "PvcExist"
	ConditionTypeConfigMap              = "ConfigMapExist"
)

var ConditionOrd = map[string]int{
	ConditionTypeReplicaSet:             0,
	ConditionTypePod:                    1,
	ConditionTypeService:                2,
	ConditionTypeSecretOpaque:           3,
	ConditionTypeSecretDockerConfigJson: 4,
	ConditionTypeSecretTls:              5,
	ConditionTypeIngress:                6,
	ConditionTypePvc:                    7,
	ConditionTypeConfigMap:              8,
}
