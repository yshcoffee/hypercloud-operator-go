package regctl

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-sdk/pkg/status"
	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"hypercloud-operator-go/internal/schemes"
)

const SecretOpaqueTypeName = regv1.ConditionTypeSecretOpaque

type RegistryCertSecret struct {
	secretOpaque *corev1.Secret
	secretTLS *corev1.Secret
	logger logr.Logger
}

func (r *RegistryCertSecret) Create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme, useGet bool) error {
	condition := status.Condition {
		Status: corev1.ConditionFalse,
		Type: ServiceTypeName,
	}

	if useGet {
		err := r.get(c, reg)
		if err != nil && !errors.IsNotFound(err) {
			r.logger.Error(err, "Getting Secret failed")
			utils.SetError(err, patchReg, condition)
			return err
		} else if err == nil {
			r.logger.Info("Secret already exist")
			return err
		}
	}

	if err := controllerutil.SetControllerReference(reg, r.secretOpaque, scheme); err != nil {
		utils.SetError(err, patchReg, condition)
		return err
	}

	if err := c.Create(context.TODO(), r.secretOpaque); err != nil {
		r.logger.Error(err, "Create failed")
		utils.SetError(err, patchReg, condition)
		return err
	}
	// [TODO] Create tls

	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryCertSecret) get(c client.Client, reg *regv1.Registry) error {
	if r.secretOpaque == nil {
		r.secretOpaque = schemes.SecretOpaque(reg)
		r.logger = utils.GetRegistryLogger(*r, r.secretOpaque.Namespace, r.secretOpaque.Name)
	}
	req := types.NamespacedName{Name: r.secretOpaque.Name, Namespace: r.secretOpaque.Namespace}

	if err := c.Get(context.TODO(), req, r.secretOpaque); err != nil {
		r.logger.Error(err, "Get failed")
		return err
	}
	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryCertSecret) Patch(c client.Client, reg *regv1.Registry, json []byte) error {
	// [TODO]
	return nil
}

func (r *RegistryCertSecret) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	return nil
}




