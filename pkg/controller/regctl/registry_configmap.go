package regctl

import (
	"context"
	"hypercloud-operator-go/internal/schemes"
	"hypercloud-operator-go/internal/utils"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"github.com/operator-framework/operator-sdk/pkg/status"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type RegistryConfigMap struct {
	cm     *corev1.ConfigMap
	logger *utils.RegistryLogger
}

func (r *RegistryConfigMap) Handle(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {
	if err := r.get(c, reg); err != nil {
		if errors.IsNotFound(err) {
			if err := r.create(c, reg, patchReg, scheme); err != nil {
				r.logger.Error(err, "create configmap error")
				return err
			}
		} else {
			r.logger.Error(err, "configmap error")
			return err
		}
	}

	return nil
}

func (r *RegistryConfigMap) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	var err error = nil
	condition := &status.Condition{
		Status: corev1.ConditionFalse,
		Type:   regv1.ConditionTypeConfigMap,
	}
	defer utils.SetError(err, patchReg, condition)

	if useGet {
		err = r.get(c, reg)
		if err != nil {
			r.logger.Error(err, "PersistentVolumeClaim error")
			return err
		}
	}

	_, exist := r.cm.Data["config.yml"]
	if !exist {
		r.logger.Info("NotReady")
		err = regv1.MakeRegistryError("NotReady")
		return err
	}

	r.logger.Info("Ready")
	condition.Status = corev1.ConditionTrue
	return nil
}

func (r *RegistryConfigMap) create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {
	if len(reg.Spec.CustomConfigYml) > 0 {
		r.logger.Info("Use exist registry configmap. Need not to create configmap. (Configmap: " + reg.Spec.CustomConfigYml + ")")
		return nil
	}

	defaultCm := &corev1.ConfigMap{}
	defaultCmType := schemes.DefaultConfigMapType()

	// Read Default ConfigMap
	if err := c.Get(context.TODO(), *defaultCmType, defaultCm); err != nil {
		r.logger.Error(err, "get default configmap error")
		return nil
	}

	r.cm = schemes.ConfigMap(reg, defaultCm.Data)

	if err := controllerutil.SetControllerReference(reg, r.cm, scheme); err != nil {
		r.logger.Error(err, "SetOwnerReference Failed")
		condition := status.Condition{
			Status:  corev1.ConditionFalse,
			Type:    regv1.ConditionTypeConfigMap,
			Message: err.Error(),
		}

		patchReg.Status.Conditions.SetCondition(condition)
		return nil
	}

	r.logger.Info("Create registry configmap")
	err := c.Create(context.TODO(), r.cm)
	if err != nil {
		condition := status.Condition{
			Status:  corev1.ConditionFalse,
			Type:    regv1.ConditionTypeConfigMap,
			Message: err.Error(),
		}

		patchReg.Status.Conditions.SetCondition(condition)
		r.logger.Error(err, "Creating registry configmap is failed.")
		return nil
	}

	return nil
}

func (r *RegistryConfigMap) get(c client.Client, reg *regv1.Registry) error {
	r.cm = schemes.ConfigMap(reg, map[string]string{})
	r.logger = utils.NewRegistryLogger(*r, r.cm.Namespace, r.cm.Name)

	req := types.NamespacedName{Name: r.cm.Name, Namespace: r.cm.Namespace}
	err := c.Get(context.TODO(), req, r.cm)
	if err != nil {
		r.logger.Error(err, "Get regsitry configmap is failed")
		return err
	}

	return nil
}

func (r *RegistryConfigMap) patch(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, diff []utils.Diff) error {
	return nil
}

func (r *RegistryConfigMap) delete(c client.Client, patchReg *regv1.Registry) error {
	if err := c.Delete(context.TODO(), r.cm); err != nil {
		r.logger.Error(err, "Unknown error delete configmap")
		return err
	}
	condition := status.Condition{
		Type:   regv1.ConditionTypeConfigMap,
		Status: corev1.ConditionFalse,
	}

	patchReg.Status.Conditions.SetCondition(condition)
	return nil
}

func (r *RegistryConfigMap) compare(reg *regv1.Registry) []utils.Diff {
	return nil
}
