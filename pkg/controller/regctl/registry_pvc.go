package regctl

import (
	"context"
	"hypercloud-operator-go/internal/schemes"
	"hypercloud-operator-go/internal/utils"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-sdk/pkg/status"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type RegistryPVC struct {
	pvc    *corev1.PersistentVolumeClaim
	logger logr.Logger
}

func (r *RegistryPVC) Create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme, useGet bool) error {
	if r.pvc == nil || useGet {
		err := r.get(c, reg)
		if err != nil && !errors.IsNotFound(err) {
			r.logger.Error(err, "pvc is error")
			return err
		} else if err == nil {
			return err
		}
	}

	if reg.Spec.PersistentVolumeClaim.Exist != nil {
		r.logger.Info("Use exist registry pvc. Need not to create pvc.")
		return nil
	}

	if err := controllerutil.SetControllerReference(reg, r.pvc, scheme); err != nil {
		r.logger.Error(err, "SetOwnerReference Failed")
		condition := status.Condition{
			Status:  corev1.ConditionFalse,
			Type:    regv1.ConditionTypePvc,
			Message: err.Error(),
		}

		patchReg.Status.Conditions.SetCondition(condition)
		return nil
	}

	r.logger.Info("Create registry pvc")
	err := c.Create(context.TODO(), r.pvc)
	if err != nil {
		condition := status.Condition{
			Status:  corev1.ConditionFalse,
			Type:    regv1.ConditionTypePvc,
			Message: err.Error(),
		}

		patchReg.Status.Conditions.SetCondition(condition)
		r.logger.Error(err, "Creating registry pvc is failed.")
		return nil
	}

	return nil
}

func (r *RegistryPVC) get(c client.Client, reg *regv1.Registry) error {
	r.pvc = schemes.PersistentVolumeClaim(reg)
	r.logger = utils.GetRegistryLogger(*r, r.pvc.Namespace, r.pvc.Name)

	req := types.NamespacedName{Name: r.pvc.Name, Namespace: r.pvc.Namespace}
	err := c.Get(context.TODO(), req, r.pvc)
	if err != nil {
		r.logger.Error(err, "Get regsitry pvc is failed")
		return err
	}

	return nil
}

func (r *RegistryPVC) Patch(c client.Client, reg *regv1.Registry, patchJson []byte) error {
	return nil
}

func (r *RegistryPVC) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	if useGet {
		err := r.get(c, reg)
		if err != nil {
			r.logger.Error(err, "PersistentVolumeClaim error")
			return err
		}
	}

	if string(r.pvc.Status.Phase) == "pending" {
		r.logger.Info("NotReady")
		condition := status.Condition{
			Status: corev1.ConditionFalse,
			Type:   regv1.ConditionTypePvc,
		}

		patchReg.Status.Conditions.SetCondition(condition)
		return nil
	}

	r.logger.Info("Ready")
	condition := status.Condition{
		Status: corev1.ConditionTrue,
		Type:   regv1.ConditionTypePvc,
	}

	patchReg.Status.Conditions.SetCondition(condition)
	return nil
}

func (r *RegistryPVC) Update(c client.Client, reg *regv1.Registry, useGet bool) error {
	return nil
}
