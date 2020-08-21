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

func (r *RegistryPod) Create(c client.Client, reg *regv1.Registry, conditions *status.Conditions, scheme *runtime.Scheme, useGet bool) error {
	return nil
}

func (r *RegistryPod) get(c client.Client, reg *regv1.Registry, conditions *status.Conditions) error {
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

func (r *RegistryPod) GetTypeName() string {
	return string(regv1.ConditionTypePvc)
}

func (r *RegistryPod) Patch(c client.Client, reg *regv1.Registry, useGet bool) error {
	return nil
}

func (r *RegistryPod) Ready(c client.Client, reg *regv1.Registry, conditions *status.Conditions, useGet bool) error {
	if r.pod == nil || useGet {
		err := r.get(c, reg, conditions)
		if err != nil {
			r.logger.Error(err, "Pod error")
			return err
		}
	}

	condition1 := status.Condition{
		Status: corev1.ConditionTrue,
		Type:   regv1.ConditionTypePod,
	}
	conditions.SetCondition(condition1)

	condition2 := status.Condition{
		Status: corev1.ConditionTrue,
		Type:   regv1.ConditionTypeContainer,
	}
	conditions.SetCondition(condition2)

	return nil
}

func (r *RegistryPod) StatusPatch(c client.Client, reg *regv1.Registry, conditions *status.Conditions, useGet bool) error {
	if r.pod == nil || useGet {
		err := r.get(c, reg, conditions)
		if err != nil {
			r.logger.Error(err, "Pod error")
			return err
		}
	}

	patch := client.MergeFrom(reg) // Set original obeject
	target := reg.DeepCopy()       // Target to Patch object

	for _, condition := range *conditions {
		r.logger.Info("patch condition", "type", string(condition.Type))
		target.Status.Conditions.SetCondition(condition)
	}

	err := c.Status().Patch(context.TODO(), target, patch)
	if err != nil {
		r.logger.Error(err, "Unknown error patching status")
		return err
	}
	return nil
}

func (r *RegistryPod) StatusUpdate(c client.Client, reg *regv1.Registry, conditions *status.Conditions, useGet bool) error {
	return nil
}

func (r *RegistryPod) Update(c client.Client, reg *regv1.Registry, useGet bool) error {
	return nil
}
