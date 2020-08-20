package v1

import "github.com/operator-framework/operator-sdk/pkg/status"

type Status string

const (
	StatusSucceeded = Status("Succeeded")
	StatusFailed    = Status("Failed")
	StatusReady     = Status("Ready")
	StatusNotReady  = Status("NotReady")
	StatusRunning   = Status("Running")
	StatusPending   = Status("Pending")
	StatusSkipped   = Status("Skipped")
	StatusCreating  = Status("Creating")

	ConditionTypeDeployment             = status.ConditionType("ConditionTypeDeploymentExist")
	ConditionTypePod                    = status.ConditionType("PodRunning")
	ConditionTypeContainer              = status.ConditionType("ContainerReady")
	ConditionTypeService                = status.ConditionType("ServiceExist")
	ConditionTypeSecretOpaque           = status.ConditionType("SecretOpaqueExist")
	ConditionTypeSecretDockerConfigJson = status.ConditionType("SecretDockerConfigJsonExist")
	ConditionTypeSecretTls              = status.ConditionType("SecretTlsExist")
	ConditionTypeIngress                = status.ConditionType("IngressExist")
	ConditionTypePvc                    = status.ConditionType("PvcExist")
	ConditionTypeConfigMap              = status.ConditionType("ConfigMapExist")
)
