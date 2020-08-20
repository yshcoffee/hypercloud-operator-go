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

func (r *RegistryPVC) Create(c client.Client, reg *regv1.Registry, condition *status.Condition, scheme *runtime.Scheme, useGet bool) error {
	if r.pvc == nil || useGet {
		err := r.get(c, reg, condition)
		if err != nil && !errors.IsNotFound(err) {
			r.logger.Error(err, "pvc is error")
			return err
		} else if err == nil {
			return err
		}
	}

	if reg.Spec.PersistentVolumeClaim.Exist != nil {
		r.logger.Info("Use exist registry pvc")
		return nil
	}

	if err := controllerutil.SetControllerReference(reg, r.pvc, scheme); err != nil {
		r.logger.Error(err, "SetOwnerReference Failed")
		return err
	}

	r.logger.Info("Create registry pvc")
	err := c.Create(context.TODO(), r.pvc)
	if err != nil {
		if condition == nil {
			condition = &status.Condition{
				Type: status.ConditionType(regv1.ConditionTypePvc),
			}
		}

		condition.Status = corev1.ConditionFalse
		condition.Message = err.Error()

		r.logger.Error(err, "Creating registry pvc is failed.")
		return err
	}

	return nil
}

func (r *RegistryPVC) get(c client.Client, reg *regv1.Registry, condition *status.Condition) error {
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

func (r *RegistryPVC) GetTypeName() string {
	return string(regv1.ConditionTypePvc)
}

func (r *RegistryPVC) Patch(c client.Client, reg *regv1.Registry, useGet bool) error {
	return nil
}

func (r *RegistryPVC) Ready(reg *regv1.Registry, useGet bool) error {
	if r.pvc == nil || useGet {
		r.get(nil, reg, nil)
	}
	if string(r.pvc.Status.Phase) == "pending" {
		return regv1.MakeRegistryError(regv1.NotReady)
	}

	return nil
}

func (r *RegistryPVC) StatusPatch(c client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	if r.pvc == nil || useGet {
		r.get(c, reg, condition)
	}

	patch := client.MergeFrom(reg) // Set original obeject
	target := reg.DeepCopy()       // Target to Patch object
	target.Status.Conditions.SetCondition(*condition)

	err := c.Status().Patch(context.TODO(), target, patch)
	if err != nil {
		r.logger.Error(err, "Unknown error patching status")
		return err
	}
	return nil
}

func (r *RegistryPVC) StatusUpdate(c client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	return nil
}

func (r *RegistryPVC) Update(c client.Client, reg *regv1.Registry, useGet bool) error {
	return nil
}
