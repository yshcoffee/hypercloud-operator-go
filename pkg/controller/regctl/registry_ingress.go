package regctl

import (
	"context"
	"github.com/operator-framework/operator-sdk/pkg/status"
	"hypercloud-operator-go/internal/schemes"
	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const IngressTypeName = regv1.ConditionTypeIngress

type RegistryIngress struct {
	ingress *v1beta1.Ingress
	logger *utils.RegistryLogger
}

func (r *RegistryIngress) Create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme, useGet bool) error {
	condition := status.Condition {
		Status: corev1.ConditionFalse,
		Type: IngressTypeName,
	}

	if useGet {
		err := r.get(c, reg)
		if err != nil && !errors.IsNotFound(err) {
			r.logger.Error(err, "Getting Ingress failed")
			utils.SetError(err, patchReg, &condition)
			return err
		} else if err == nil {
			r.logger.Info("Ingress already exist")
			return err
		}
	}

	if err := controllerutil.SetControllerReference(reg, r.ingress, scheme); err != nil {
		r.logger.Error(err, "Controller reference failed")
		utils.SetError(err, patchReg, &condition)
		return err
	}

	if err := c.Create(context.TODO(), r.ingress); err != nil {
		r.logger.Error(err, "Create failed")
		utils.SetError(err, patchReg, &condition)
		return err
	}

	r.logger.Info("Succeed")
	return nil
}


func (r *RegistryIngress) get(c client.Client, reg *regv1.Registry) error {
	r.ingress = schemes.Ingress(reg)
	if r.ingress == nil {
		return regv1.MakeRegistryError("Registry has no fields Ingress required")
	}
	r.logger = utils.NewRegistryLogger(*r, r.ingress.Namespace, r.ingress.Name)


	req := types.NamespacedName{Name: r.ingress.Name, Namespace: r.ingress.Namespace}
	if err := c.Get(context.TODO(), req, r.ingress); err != nil {
		r.logger.Error(err, "Get failed")
		return err
	}

	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryIngress) Patch(c client.Client, reg *regv1.Registry, json []byte) error {
	// [TODO}
	return nil
}

func (r *RegistryIngress) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	var err error = nil
	condition := status.Condition {
		Status: corev1.ConditionFalse,
		Type: IngressTypeName,
	}

	defer utils.SetError(err, patchReg, &condition)

	if useGet {
		if err = r.get(c, reg); err != nil {
			r.logger.Error(err, "Get failed")
			return err
		}
	}

	err = regv1.MakeRegistryError("Ingress Error")
	if _, ok := r.ingress.Annotations["kubernetes.io/ingress.class"]; !ok {
		return err
	}
	if _, ok := r.ingress.Annotations["nginx.ingress.kubernetes.io/proxy-connect-timeout"]; !ok {
		return err
	}
	if _, ok := r.ingress.Annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"]; !ok {
		return err
	}
	if _, ok := r.ingress.Annotations["nginx.ingress.kubernetes.io/ssl-redirect"]; !ok {
		return err
	}
	if val, ok := r.ingress.Annotations["nginx.ingress.kubernetes.io/backend-protocol"]; ok {
		if val != "HTTPS" {
			return err
		}
	} else {
		return err
	}
	if _, ok := r.ingress.Annotations["nginx.ingress.kubernetes.io/proxy-body-size"]; !ok {
		return err
	}

	err = nil
	condition.Status = corev1.ConditionTrue
	r.logger.Info("Succeed")
	return nil
}


