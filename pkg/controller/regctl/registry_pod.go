package regctl

import (
	"context"
	"hypercloud-operator-go/internal/utils"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"github.com/operator-framework/operator-sdk/pkg/status"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RegistryPod struct {
	pod    *corev1.Pod
	logger *utils.RegistryLogger
}

func (r *RegistryPod) Handle(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {
	if err := r.get(c, reg); err != nil {
		r.logger.Error(err, "Pod error")
		return err
	}

	r.logger.Info("Check if recreating pod is required.")
	if reg.Status.PodRecreateRequired {
		if err := r.delete(c, patchReg); err != nil {
			return err
		}

		r.logger.Info("Recreate pod.")
		patchReg.Status.PodRecreateRequired = false
	}

	return nil
}

func (r *RegistryPod) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	var err error = nil
	podCondition := &status.Condition{
		Type:   regv1.ConditionTypePod,
		Status: corev1.ConditionFalse,
	}
	contCondition := &status.Condition{
		Type:   regv1.ConditionTypeContainer,
		Status: corev1.ConditionFalse,
	}
	defer utils.SetError(err, patchReg, podCondition)
	defer utils.SetError(err, patchReg, contCondition)

	if r.pod == nil || useGet {
		err = r.get(c, reg)
		if err != nil {
			r.logger.Error(err, "Pod error")
			return err
		}
	}

	if r.pod == nil {
		r.logger.Info("Pod is nil")
		podCondition.Status = corev1.ConditionFalse
		contCondition.Status = corev1.ConditionFalse
		err = regv1.MakeRegistryError(regv1.PodNotFound)
		return err
	}

	contStatuses := r.pod.Status.ContainerStatuses
	if len(contStatuses) == 0 {
		r.logger.Info("Container's status is nil")
		podCondition.Status = corev1.ConditionFalse
		contCondition.Status = corev1.ConditionFalse
		err = regv1.MakeRegistryError(regv1.ContainerStatusIsNil)
		return err
	}
	contState := r.pod.Status.ContainerStatuses[0]
	var reason string

	if contState.State.Waiting != nil {
		reason = contState.State.Waiting.Reason
		r.logger.Info(reason)
	} else if contState.State.Running != nil {
		// r.logger.Info(contState.String())
		if contState.Ready {
			reason = "Running"
		} else {
			reason = "NotReady"
		}
	} else if contState.State.Terminated != nil {
		reason = contState.State.Terminated.Reason
		r.logger.Info(reason)
	} else {
		reason = "Unknown"
	}

	r.logger.Info("Get container state", "reason", reason)

	switch reason {
	case "NotReady":
		podCondition.Status = corev1.ConditionTrue
		contCondition.Status = corev1.ConditionFalse
		err = regv1.MakeRegistryError(regv1.PodNotRunning)
		return err

	case "Running":
		podCondition.Status = corev1.ConditionTrue
		contCondition.Status = corev1.ConditionTrue

	default:
		podCondition.Status = corev1.ConditionFalse
		contCondition.Status = corev1.ConditionFalse
		err = regv1.MakeRegistryError(regv1.PodNotRunning)
		return err
	}

	return nil
}

func (r *RegistryPod) create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {

	return nil
}

func (r *RegistryPod) get(c client.Client, reg *regv1.Registry) error {
	r.pod = &corev1.Pod{}
	r.logger = utils.NewRegistryLogger(*r, reg.Namespace, reg.Name+" registry's pod")

	podList := &corev1.PodList{}
	label := map[string]string{}
	label["app"] = "registry"
	label["apps"] = regv1.K8sPrefix + reg.Name

	labelSelector := labels.SelectorFromSet(labels.Set(label))
	listOps := &client.ListOptions{
		Namespace:     reg.Namespace,
		LabelSelector: labelSelector,
	}
	err := c.List(context.TODO(), podList, listOps)
	if err != nil {
		r.logger.Error(err, "Failed to list pods.")
		return err
	}

	if len(podList.Items) == 0 {
		return regv1.MakeRegistryError(regv1.PodNotFound)
	}

	r.pod = &podList.Items[0]

	r.logger = utils.NewRegistryLogger(*r, r.pod.Namespace, r.pod.Name)

	return nil
}

func (r *RegistryPod) patch(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, diff []utils.Diff) error {
	return nil
}

func (r *RegistryPod) delete(c client.Client, patchReg *regv1.Registry) error {
	if err := c.Delete(context.TODO(), r.pod); err != nil {
		r.logger.Error(err, "Unknown error delete pod")
		return err
	}

	podCondition := status.Condition{
		Type:   regv1.ConditionTypePod,
		Status: corev1.ConditionFalse,
	}

	contCondition := status.Condition{
		Type:   regv1.ConditionTypeContainer,
		Status: corev1.ConditionFalse,
	}

	patchReg.Status.Conditions.SetCondition(podCondition)
	patchReg.Status.Conditions.SetCondition(contCondition)

	return nil
}

func (r *RegistryPod) compare(reg *regv1.Registry) []utils.Diff {
	return nil
}
