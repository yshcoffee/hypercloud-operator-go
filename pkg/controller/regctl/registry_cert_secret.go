package regctl

import (
	"context"
	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"github.com/operator-framework/operator-sdk/pkg/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"hypercloud-operator-go/internal/schemes"
)

const SecretOpaqueTypeName = regv1.ConditionTypeSecretOpaque
const SecretTLSTypeName = regv1.ConditionTypeSecretTls

type RegistryCertSecret struct {
	secretOpaque *corev1.Secret
	secretTLS    *corev1.Secret
	logger       *utils.RegistryLogger
}

func (r *RegistryCertSecret) Create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme, useGet bool) error {
	condition := status.Condition{
		Status: corev1.ConditionFalse,
		Type:   SecretOpaqueTypeName,
	}

	tlsCondition := status.Condition{
		Status: corev1.ConditionFalse,
		Type:   SecretTLSTypeName,
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

	if err := controllerutil.SetControllerReference(reg, r.secretTLS, scheme); err != nil {
		utils.SetError(err, patchReg, tlsCondition)
		return err
	}

	if err := c.Create(context.TODO(), r.secretOpaque); err != nil {
		r.logger.Error(err, "Create failed")
		utils.SetError(err, patchReg, condition)
		return err
	}

	if err := c.Create(context.TODO(), r.secretTLS); err != nil {
		r.logger.Error(err, "Create failed")
		utils.SetError(err, patchReg, tlsCondition)
		return err
	}

	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryCertSecret) get(c client.Client, reg *regv1.Registry) error {
	r.secretOpaque, r.secretTLS = schemes.Secrets(reg)
	if r.secretOpaque == nil && r.secretTLS == nil {
		return regv1.MakeRegistryError("Registry has no fields Secrets required")
	}

	r.logger = utils.NewRegistryLogger(*r, r.secretOpaque.Namespace, r.secretOpaque.Name)

	req := types.NamespacedName{Name: r.secretOpaque.Name, Namespace: r.secretOpaque.Namespace}
	if err := c.Get(context.TODO(), req, r.secretOpaque); err != nil {
		r.logger.Error(err, "Get failed")
		return err
	}

	req = types.NamespacedName{Name: r.secretTLS.Name, Namespace: r.secretTLS.Namespace}
	if err := c.Get(context.TODO(), req, r.secretTLS); err != nil {
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
	var opaqueErr error = nil
	var err error = nil

	condition := status.Condition{
		Status: corev1.ConditionFalse,
		Type:   regv1.ConditionTypeSecretOpaque,
	}

	tlsCondition := status.Condition{
		Status: corev1.ConditionFalse,
		Type:   regv1.ConditionTypeSecretTls,
	}

	defer utils.SetError(opaqueErr, patchReg, condition)

	if useGet {
		if opaqueErr = r.get(c, reg); opaqueErr != nil {
			r.logger.Error(opaqueErr, "Get failed")
			return opaqueErr
		}
	}

	// DATA Check
	opaqueErr = regv1.MakeRegistryError("Secret Opaque Error")
	if _, ok := r.secretOpaque.Data[schemes.CertCrtFile]; !ok {
		r.logger.Error(opaqueErr, "No certificate in data")
		return nil
	}

	if _, ok := r.secretOpaque.Data[schemes.CertKeyFile]; !ok {
		r.logger.Error(opaqueErr, "No private key in data")
		return nil
	}
	condition.Status = corev1.ConditionTrue

	defer utils.SetError(err, patchReg, tlsCondition)
	err = regv1.MakeRegistryError("Secret TLS Error")
	if _, ok := r.secretTLS.Data[schemes.TLSCert]; !ok {
		r.logger.Error(err, "No certificate in data")
		return nil
	}

	if _, ok := r.secretTLS.Data[schemes.TLSKey]; !ok {
		r.logger.Error(err, "No private key in data")
		return nil
	}

	tlsCondition.Status = corev1.ConditionTrue
	err = nil
	r.logger.Info("Succeed")
	return nil
}
