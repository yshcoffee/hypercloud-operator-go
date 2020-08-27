package regctl

import (
	"context"
	"hypercloud-operator-go/internal/utils"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"github.com/operator-framework/operator-sdk/pkg/status"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RegistryPod struct {
	pod    *corev1.Pod
	logger *utils.RegistryLogger
}

func (r *RegistryPod) Create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme, useGet bool) error {
	if r.pod == nil || useGet {
		err := r.get(c, reg)
		if err != nil && !errors.IsNotFound(err) {
			r.logger.Error(err, "pod is error")
			return err
		} else if err == nil {
			return err
		}
	}

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

func (r *RegistryPod) Patch(c client.Client, reg *regv1.Registry, patchJson []byte) error {
	return nil
}

func (r *RegistryPod) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	podCondition := status.Condition{
		Type: regv1.ConditionTypePod,
	}
	contCondition := status.Condition{
		Type: regv1.ConditionTypeContainer,
	}

	if r.pod == nil || useGet {
		err := r.get(c, reg)
		if err != nil {
			r.logger.Error(err, "Pod error")
			return err
		}
	}

	if r.pod == nil {
		r.logger.Info("pod is nil")
		podCondition.Status = corev1.ConditionFalse
		contCondition.Status = corev1.ConditionFalse

		patchReg.Status.Conditions.SetCondition(podCondition)
		patchReg.Status.Conditions.SetCondition(contCondition)
		return regv1.MakeRegistryError(regv1.PodNotFound)
	}

	contStatuses := r.pod.Status.ContainerStatuses
	if len(contStatuses) == 0 {
		r.logger.Info("Container's status is nil")
		podCondition.Status = corev1.ConditionFalse
		contCondition.Status = corev1.ConditionFalse

		patchReg.Status.Conditions.SetCondition(podCondition)
		patchReg.Status.Conditions.SetCondition(contCondition)
		return regv1.MakeRegistryError(regv1.ContainerStatusIsNil)
	}
	contState := r.pod.Status.ContainerStatuses[0]
	var reason string

	if contState.State.Waiting != nil {
		reason = contState.State.Waiting.Reason
		r.logger.Info(reason)
	} else if contState.State.Running != nil {
		r.logger.Info(contState.String())
		if contState.Ready {
			reason = "Running"
		} else {
			reason = "NotReady"
		}
	} else if contState.State.Terminated != nil {
		reason = contState.State.Waiting.Reason
		r.logger.Info(reason)
	} else {
		reason = "Unknown"
	}

	r.logger.Info("Get container state", "reason", reason)

	switch reason {
	case "NotReady":
		podCondition.Status = corev1.ConditionTrue
		contCondition.Status = corev1.ConditionFalse

	case "Running":
		podCondition.Status = corev1.ConditionTrue
		contCondition.Status = corev1.ConditionTrue

	default:
		podCondition.Status = corev1.ConditionFalse
		contCondition.Status = corev1.ConditionFalse
	}

	patchReg.Status.Conditions.SetCondition(podCondition)
	patchReg.Status.Conditions.SetCondition(contCondition)

	return nil
}