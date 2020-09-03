package regctl

import (
	"context"
	"hypercloud-operator-go/internal/schemes"
	"hypercloud-operator-go/internal/utils"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"github.com/operator-framework/operator-sdk/pkg/status"
	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type RegistryDeployment struct {
	deploy *appsv1.Deployment
	logger *utils.RegistryLogger
}

func (r *RegistryDeployment) Handle(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {
	if err := r.get(c, reg); err != nil {
		if errors.IsNotFound(err) {
			if err := r.create(c, reg, patchReg, scheme, false); err != nil {
				r.logger.Error(err, "create Deployment error")
				return err
			}
		} else {
			r.logger.Error(err, "Deployment error")
			return err
		}
	}

	r.logger.Info("Check if patch exists.")
	diff, _ := r.compare(c, reg, false)
	if len(diff) > 0 {
		r.patch(c, reg, patchReg, diff)
	}

	return nil
}

func (r *RegistryDeployment) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	if useGet {
		err := r.get(c, reg)
		if err != nil {
			r.logger.Error(err, "Deployment error")
			return err
		}
	}

	if r.deploy == nil {
		r.logger.Info("NotReady")
		condition := status.Condition{
			Status: corev1.ConditionFalse,
			Type:   regv1.ConditionTypeDeployment,
		}

		patchReg.Status.Conditions.SetCondition(condition)
		return nil
	}

	r.logger.Info("Ready")
	condition := status.Condition{
		Status: corev1.ConditionTrue,
		Type:   regv1.ConditionTypeDeployment,
	}

	patchReg.Status.Conditions.SetCondition(condition)
	return nil
}

func (r *RegistryDeployment) create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme, useGet bool) error {
	if r.deploy == nil || useGet {
		err := r.get(c, reg)
		if err != nil && !errors.IsNotFound(err) {
			r.logger.Error(err, "Deployment error")
			return err
		} else if err == nil {
			return err
		}
	}

	if err := controllerutil.SetControllerReference(reg, r.deploy, scheme); err != nil {
		r.logger.Error(err, "SetOwnerReference Failed")
		condition := status.Condition{
			Status:  corev1.ConditionFalse,
			Type:    regv1.ConditionTypeDeployment,
			Message: err.Error(),
		}

		patchReg.Status.Conditions.SetCondition(condition)
		return nil
	}

	r.logger.Info("Create registry deployment")
	err := c.Create(context.TODO(), r.deploy)
	if err != nil {
		condition := status.Condition{
			Status:  corev1.ConditionFalse,
			Type:    regv1.ConditionTypeDeployment,
			Message: err.Error(),
		}

		patchReg.Status.Conditions.SetCondition(condition)
		r.logger.Error(err, "Creating registry deployment is failed.")
		return nil
	}

	return nil
}

func (r *RegistryDeployment) get(c client.Client, reg *regv1.Registry) error {
	r.deploy = schemes.Deployment(reg)
	r.logger = utils.NewRegistryLogger(*r, r.deploy.Namespace, r.deploy.Name)

	req := types.NamespacedName{Name: r.deploy.Name, Namespace: r.deploy.Namespace}

	err := c.Get(context.TODO(), req, r.deploy)
	if err != nil {
		r.logger.Error(err, "Get regsitry deployment is failed")
		return err
	}

	return nil
}

func (r *RegistryDeployment) patch(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, diff []utils.Diff) error {
	target := r.deploy.DeepCopy()
	originObject := client.MergeFrom(r.deploy)

	for _, d := range diff {
		switch d.Key {
		case "DeleteWithPvc":

		}
	}

	// Patch
	if err := c.Patch(context.TODO(), target, originObject); err != nil {
		r.logger.Error(err, "Unknown error patching status")
		return err
	}
	return nil
}

func (r *RegistryDeployment) delete(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	if r.deploy == nil || useGet {
		err := r.get(c, reg)
		if err != nil {
			r.logger.Error(err, "deploy error")
			return err
		}
	}

	condition := status.Condition{
		Type:   regv1.ConditionTypeDeployment,
		Status: corev1.ConditionFalse,
	}

	patchReg.Status.Conditions.SetCondition(condition)

	c.Delete(context.TODO(), r.deploy)
	return nil
}

func (r *RegistryDeployment) compare(c client.Client, reg *regv1.Registry, useGet bool) ([]utils.Diff, bool) {
	diff := []utils.Diff{}

	return diff, len(diff) > 0
}
