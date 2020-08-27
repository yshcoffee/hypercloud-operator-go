package regctl

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-sdk/pkg/status"
	"hypercloud-operator-go/internal/schemes"
	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	SecretDCJTypeName = regv1.ConditionTypeSecretDockerConfigJson
)

type RegistryDCJSecret struct {
	secretDCJ *corev1.Secret
	logger    logr.Logger
}

func (r *RegistryDCJSecret) Create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme, useGet bool) error {
	condition := status.Condition{
		Status: corev1.ConditionFalse,
		Type: SecretDCJTypeName,
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

	if err := controllerutil.SetControllerReference(reg, r.secretDCJ, scheme); err != nil {
		utils.SetError(err, patchReg, condition)
		return err
	}

	if err := c.Create(context.TODO(), r.secretDCJ); err != nil {
		r.logger.Error(err, "Create failed")
		utils.SetError(err, patchReg, condition)
		return err
	}

	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryDCJSecret) get(c client.Client, reg *regv1.Registry) error {
	r.secretDCJ = schemes.DCJSecret(reg)
	if r.secretDCJ == nil {
		return regv1.MakeRegistryError("Registry has no fields DCJ Secret required")
	}
	r.logger = utils.GetRegistryLogger(*r, r.secretDCJ.Namespace, r.secretDCJ.Name)

	req := types.NamespacedName{Name: r.secretDCJ.Name, Namespace: r.secretDCJ.Namespace}
	if err := c.Get(context.TODO(), req, r.secretDCJ); err != nil {
		r.logger.Error(err, "Get failed")
		return err
	}

	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryDCJSecret) Patch(c client.Client, reg *regv1.Registry, json []byte) error {
	return nil
}

func (r *RegistryDCJSecret) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	if useGet {
		if err := r.get(c, reg); err != nil {
			r.logger.Error(err, "Get failed")
			return err
		}
	}

	condition := status.Condition {
		Status: corev1.ConditionFalse,
		Type: SecretDCJTypeName,
	}

	err := regv1.MakeRegistryError("Secret DCJ Error")
	if _, ok := r.secretDCJ.Data[schemes.DockerConfigJson]; !ok {
		r.logger.Error(err, "No certificate in data")
		utils.SetError(err, patchReg, condition)
		return nil
	}

	r.logger.Info("Succeed")
	return nil
}
