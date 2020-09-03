package regctl

import (
	"context"
	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	"strconv"

	"github.com/operator-framework/operator-sdk/pkg/status"
	corev1 "k8s.io/api/core/v1"
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

func (r *RegistryCertSecret) Handle(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {
	err := r.get(c, reg)
	if err != nil {
		// resource is not exist : have to create
		if  createError := r.create(c, reg, patchReg, scheme); createError != nil {
			r.logger.Error(createError, "Create failed in Handle")
			return createError
		}
	}

	if  isValid := r.compare(reg); isValid == nil {
		if deleteError := r.delete(c, patchReg); deleteError != nil {
			r.logger.Error(deleteError, "Delete failed in Handle")
			return deleteError
		}
	}

	return nil
}

func (r *RegistryCertSecret) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	var opaqueErr error = nil
	var err error = nil

	condition := &status.Condition{
		Status: corev1.ConditionFalse,
		Type:   SecretOpaqueTypeName,
	}

	tlsCondition := &status.Condition{
		Status: corev1.ConditionFalse,
		Type:   SecretTLSTypeName,
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

func (r *RegistryCertSecret) create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {
	condition := &status.Condition{
		Status: corev1.ConditionFalse,
		Type:   SecretOpaqueTypeName,
	}

	tlsCondition := &status.Condition{
		Status: corev1.ConditionFalse,
		Type:   SecretTLSTypeName,
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

func (r *RegistryCertSecret) patch(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, diff []utils.Diff) error {
	// [TODO]
	return nil
}

func (r *RegistryCertSecret) delete(c client.Client, patchReg *regv1.Registry) error {
	condition := &status.Condition{
		Status: corev1.ConditionFalse,
		Type:   SecretOpaqueTypeName,
	}

	tlsCondition := &status.Condition{
		Status: corev1.ConditionFalse,
		Type:   SecretTLSTypeName,
	}

	if err := c.Delete(context.TODO(), r.secretOpaque); err != nil {
		r.logger.Error(err, "Delete failed")
		utils.SetError(err, patchReg, condition)
		return err
	}

	if err := c.Delete(context.TODO(), r.secretTLS); err != nil {
		r.logger.Error(err, "Delete failed")
		utils.SetError(err, patchReg, tlsCondition)
		return err
	}

	return nil
}

func (r *RegistryCertSecret) compare(reg *regv1.Registry) ([]utils.Diff) {
	opaqueData := r.secretOpaque.Data
	if string(opaqueData["ID"]) != reg.Spec.LoginId || string(opaqueData["PASSWD"]) != reg.Spec.LoginPassword {
		return nil
	}

	if string(opaqueData["CLUSTER_IP"]) != reg.Status.ClusterIP {
		return nil
	}

	if reg.Spec.RegistryService.ServiceType == regv1.RegServiceTypeLoadBalancer {
		val, ok := opaqueData["LB_IP"]
		if !ok || string(val) != reg.Status.LoadBalancerIP {
			return nil
		}

		val, ok = opaqueData["REGISTRY_URL"]
		if !ok || string(val) != reg.Status.LoadBalancerIP + ":" + strconv.Itoa(reg.Spec.RegistryService.LoadBalancer.Port) {
			return nil
		}
	} else if reg.Spec.RegistryService.ServiceType == regv1.RegServiceTypeIngress {
		registryDomainName := reg.Name + "." + reg.Spec.RegistryService.Ingress.DomainName
		val, ok := opaqueData["DOMAIN_NAME"]
		if !ok || string(val) != registryDomainName {
			return nil
		}

		val, ok = opaqueData["REGISTRY_URL"]
		if !ok || string(val) != registryDomainName + ":" + string(443) {
			return nil
		}
	} else {
		val, ok := opaqueData["REGISTRY_URL"]
		if !ok || string(val) != reg.Status.ClusterIP + ":" + string(443) {
			return nil
		}
	}

	_, ok := opaqueData[schemes.CertCrtFile]
	if !ok {
		return nil
	}

	_, ok = opaqueData[schemes.CertKeyFile]
	if !ok {
		return nil
	}

	return []utils.Diff{}
}

