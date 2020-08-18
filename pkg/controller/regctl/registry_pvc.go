package regctl

import (
	"context"
	"hypercloud-operator-go/internal/schemes"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"github.com/operator-framework/operator-sdk/pkg/status"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RegistryPVC struct {
	pvc *corev1.PersistentVolumeClaim
}

func (r *RegistryPVC) Create(c client.Client, reg *regv1.Registry, condition *status.Condition, scheme *runtime.Scheme, useGet bool) error {
	reqLogger := log.Log.WithValues("RegistryPVC.Namespace", reg.Namespace, "RegistryPVC.Name", reg.Name)

	r.pvc = schemes.PersistentVolumeClaim(reg)
	if useGet {
		err := r.get(c, reg, condition)
		if err != nil && !errors.IsNotFound(err) {
			reqLogger.Error(err, "pvc is error")
			return err
		} else if err == nil {
			return err
		}
	}

	if reg.Spec.PersistentVolumeClaim.Exist != nil {
		reqLogger.Info("Use exist registry pvc", "pvc.name", r.pvc.Name, "pvc.namespace", r.pvc.Namespace)
		return nil
	}

	if err := controllerutil.SetControllerReference(reg, r.pvc, scheme); err != nil {
		reqLogger.Error(err, "SetOwnerReference Failed")
		return err
	}

	reqLogger.Info("Create registry pvc", "pvc.name", r.pvc.Name, "pvc.namespace", r.pvc.Namespace)
	err := c.Create(context.TODO(), r.pvc)
	if err != nil {
		if condition == nil {
			condition = &status.Condition{
				Type: status.ConditionType(regv1.ConditionTypePvc),
			}
		}

		condition.Status = corev1.ConditionFalse
		condition.Message = err.Error()

		reqLogger.Error(err, "Creating registry pvc is failed.")
		return err
	}

	return nil
}

func (r *RegistryPVC) get(c client.Client, reg *regv1.Registry, condition *status.Condition) error {
	reqLogger := log.Log.WithValues("RegistryPVC.Namespace", reg.Namespace, "RegistryPVC.Name", reg.Name)

	if r.pvc != nil {
		r.pvc = schemes.PersistentVolumeClaim(reg)
	}

	req := types.NamespacedName{Name: r.pvc.Name, Namespace: r.pvc.Namespace}

	err := c.Get(context.TODO(), req, r.pvc)
	if err != nil {
		reqLogger.Error(err, "Get regsitry pvc is failed")
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

func (r *RegistryPVC) Ready(reg *regv1.Registry, useGet bool) bool {
	if string(r.pvc.Status.Phase) == "pending" {
		return false
	}

	return true
}

func (r *RegistryPVC) StatusPatch(c client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	reqLogger := log.Log.WithValues("RegistryPVC.Namespace", reg.Namespace, "RegistryPVC.Name", reg.Name)

	if useGet {
		r.get(c, reg, condition)
	}

	patch := client.MergeFrom(reg) // Set original obeject
	target := reg.DeepCopy()       // Target to Patch object
	target.Status.Conditions.SetCondition(*condition)

	err := c.Status().Patch(context.TODO(), target, patch)
	if err != nil {
		reqLogger.Error(err, "Unknown error patching status")
		return err
	}
	return nil
}

func (r *RegistryPVC) StatusUpdate(c client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	if useGet {
		r.get(c, reg, condition)
	}

	return nil
}

func (r *RegistryPVC) Update(c client.Client, reg *regv1.Registry, useGet bool) error {

	return nil
}
