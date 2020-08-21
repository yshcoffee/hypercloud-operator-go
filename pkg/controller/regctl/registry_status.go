package regctl

import (
	"context"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-sdk/pkg/status"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// UpdateRegistryStatus ...
// If registry status is patched, return true.
func UpdateRegistryStatus(c client.Client, reg *regv1.Registry) bool {
	reqLogger := logf.Log.WithName("controller_registry").WithValues("Request.Namespace", reg.Namespace, "Request.Name", reg.Name)
	falseTypes := []status.ConditionType{}
	checkTypes := getCheckTypes(reg)

	var desiredStatus regv1.Status

	// Check if all subresources are true
	reqLogger.Info("Check if status fields are normal.")
	for _, t := range checkTypes {
		if reg.Status.Conditions.IsUnknownFor(t) {
			reqLogger.Info("Initialize status fields")
			InitRegistryStatus(c, reg)

			if reg.Status.Phase == string(regv1.StatusCreating) {
				reqLogger.Info("status fields are abnormal. Initialize the registry status.")
				return false
			}

			desiredStatus = regv1.StatusCreating
			patch := client.MergeFrom(reg)
			target := reg.DeepCopy()
			message := "Registry is creating. All resources in registry has not yet been created."
			reason := "RegistryNotCreated"

			target.Status.Message = message
			target.Status.Reason = reason
			target.Status.Phase = string(desiredStatus)
			target.Status.PhaseChangedAt = metav1.Now()

			reqLogger.Info("Current Status(" + reg.Status.Phase + ") -> Desired Status(" + string(desiredStatus) + ")")
			// Patch the status to desired status.
			c.Status().Patch(context.TODO(), target, patch)
			return true

		} else if reg.Status.Conditions.IsFalseFor(t) {
			falseTypes = append(falseTypes, t)
		}
	}

	reqLogger.Info("Get desired status.")
	for _, t := range falseTypes {
		if len(falseTypes) > 1 {
			desiredStatus = regv1.StatusCreating
		} else if len(falseTypes) == 1 && t == regv1.ConditionTypePod {
			desiredStatus = regv1.StatusNotReady
		} else {
			desiredStatus = regv1.StatusRunning
		}
	}

	// Chcck if current status is desired status. If does not same, patch the status.
	reqLogger.Info("Check if current status is desired status.")
	if reg.Status.Phase == string(desiredStatus) {
		return false
	}
	reqLogger.Info("Current Status(" + reg.Status.Phase + ") -> Desired Status(" + string(desiredStatus) + ")")

	var message, reason string
	patch := client.MergeFrom(reg)
	target := reg.DeepCopy()

	switch desiredStatus {
	case regv1.StatusCreating:
		message = "Registry is creating. All resources in registry has not yet been created."
		reason = "RegistryNotCreated"
	case regv1.StatusNotReady:
		message = "Registry is not ready."
		reason = "NotReady"
	case regv1.StatusRunning:
		message = "Registry is running. All registry resources are operating normally."
		reason = "Running"
	}

	target.Status.Message = message
	target.Status.Reason = reason
	target.Status.Phase = string(desiredStatus)
	target.Status.PhaseChangedAt = metav1.Now()

	// Patch the status to desired status.
	reqLogger.Info("Status patch.")
	c.Status().Patch(context.TODO(), target, patch)

	return true
}

func InitRegistryStatus(c client.Client, reg *regv1.Registry) {
	reqLogger := logf.Log.WithName("controller_registry").WithValues("Request.Namespace", reg.Namespace, "Request.Name", reg.Name)

	if reg.Status.Conditions == nil {
		reg.Status.Conditions = status.NewConditions()
	}

	// Set Conditions
	checkTypes := getCheckTypes(reg)
	for _, t := range checkTypes {
		reqLogger.Info("Check Type: " + string(t))
		if reg.Status.Conditions.GetCondition(t) == nil {
			newCondition := status.Condition{Type: t, Status: corev1.ConditionFalse}
			reg.Status.Conditions.SetCondition(newCondition)
		}
	}

	reg.Status.Message = "registry is creating."
	reg.Status.Reason = "Creating"
	reg.Status.Phase = string(regv1.StatusCreating)
	reg.Status.PhaseChangedAt = metav1.Now()

	err := c.Status().Update(context.TODO(), reg)
	if err != nil {
		reqLogger.Error(err, "cannot update status")
	}
}

func getCheckTypes(reg *regv1.Registry) []status.ConditionType {
	checkTypes := []status.ConditionType{
		regv1.ConditionTypeDeployment,
		regv1.ConditionTypePod,
		regv1.ConditionTypeContainer,
		regv1.ConditionTypeService,
		regv1.ConditionTypeSecretOpaque,
		regv1.ConditionTypeSecretDockerConfigJson,
		regv1.ConditionTypePvc,
		regv1.ConditionTypeConfigMap,
	}

	if reg.Spec.RegistryService.Ingress != nil {
		checkTypes = append(checkTypes, regv1.ConditionTypeSecretTls, regv1.ConditionTypeIngress)
	}

	return checkTypes
}
