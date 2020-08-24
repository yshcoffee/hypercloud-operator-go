package regctl

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"hypercloud-operator-go/internal/schemes"
)

const SecretOpaqueTypeName = regv1.ConditionTypeSecretOpaque

type RegistryCertSecret struct {
	secretOpaque *corev1.Secret
	secretTLS *corev1.Secret
}

func (r *RegistryCertSecret) Create(c client.Client, reg *regv1.Registry, condition *status.Condition, scheme *runtime.Scheme, useGet bool) error {
	logger := utils.GetRegistryLogger(r, reg.Namespace, reg.Name)
	if r.secretOpaque == nil {
		r.secretOpaque = schemes.SecretOpaque(reg)
		logger.Info("Create Secret Done")
	}
	// [TODO] Owner Reference to Opaque Secret
	return nil
}

func (r *RegistryCertSecret) get(c client.Client, reg *regv1.Registry, condition *status.Condition) error {
	return nil
}

func (r *RegistryCertSecret) GetTypeName() string {
	return string(ServiceTypeName)
}

func (r *RegistryCertSecret) Patch(c client.Client, reg *regv1.Registry, useGet bool) error {
	// [TODO]
	return nil
}

func (r *RegistryCertSecret) Ready(reg *regv1.Registry, useGet bool) error {
	return nil
}

func (r *RegistryCertSecret) StatusPatch(c client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	return nil
}

func (r *RegistryCertSecret) StatusUpdate(c client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	return nil
}

func (r *RegistryCertSecret) Update(c client.Client, reg *regv1.Registry, useGet bool) error {
	return nil
}



