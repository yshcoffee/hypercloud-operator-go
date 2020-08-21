package regctl

import (
	"context"
	"hypercloud-operator-go/internal/utils"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-sdk/pkg/status"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RegistryPod struct {
	pod    *corev1.Pod
	logger logr.Logger
}

func (r *RegistryPod) Create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme, useGet bool) error {
	return nil
}

func (r *RegistryPod) get(c client.Client, reg *regv1.Registry) error {
	r.pod = &corev1.Pod{}
	r.logger = utils.GetRegistryLogger(*r, reg.Namespace, reg.Name+" registry's pod")

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

	r.logger = utils.GetRegistryLogger(*r, r.pod.Namespace, r.pod.Name)

	return nil
}

func (r *RegistryPod) Patch(c client.Client, reg *regv1.Registry, useGet bool) error {
	return nil
}

func (r *RegistryPod) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	if r.pod == nil || useGet {
		err := r.get(c, reg)
		if err != nil {
			r.logger.Error(err, "Pod error")
			return err
		}
	}

	condition1 := status.Condition{
		Status: corev1.ConditionTrue,
		Type:   regv1.ConditionTypePod,
	}
	patchReg.Status.Conditions.SetCondition(condition1)

	condition2 := status.Condition{
		Status: corev1.ConditionTrue,
		Type:   regv1.ConditionTypeContainer,
	}
	patchReg.Status.Conditions.SetCondition(condition2)

	return nil
}
