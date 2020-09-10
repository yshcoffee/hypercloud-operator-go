package regctl

import (
	"context"
	"hypercloud-operator-go/internal/schemes"
	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/operator-framework/operator-sdk/pkg/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const ServiceTypeName = regv1.ConditionTypeService

type RegistryService struct {
	svc *corev1.Service
	logger *utils.RegistryLogger
}

func (r *RegistryService) Handle(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {
	err := r.get(c, reg)
	if err != nil && errors.IsNotFound(err) {
		// resource is not exist : have to create
		if  createError := r.create(c, reg, patchReg, scheme); createError != nil {
			r.logger.Error(createError, "Create failed in Handle")
			return createError
		}
	}

	if isValid := r.compare(reg); isValid == nil {
		r.logger.Info("Service is not Valid")
		if deleteError := r.delete(c, reg); deleteError != nil {
			r.logger.Error(deleteError, "Delete failed in Handle")
			return deleteError
		}
	}

	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryService) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	var err error = nil
	condition := &status.Condition {
		Status: corev1.ConditionFalse,
		Type: ServiceTypeName,
	}
	defer utils.SetError(err, patchReg, condition)

	if useGet {
		if err = r.get(c, reg); err != nil {
			r.logger.Error(err, "Getting Service error")
			return err
		}
	}

	if r.svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		loadBalancer := r.svc.Status.LoadBalancer
		lbIP := ""
		// [TODO] Specific Condition is needed
		if len(loadBalancer.Ingress) == 1 {
			if loadBalancer.Ingress[0].Hostname == "" {
				lbIP = loadBalancer.Ingress[0].IP
			} else {
				lbIP = loadBalancer.Ingress[0].Hostname
			}
		} else if len(loadBalancer.Ingress) == 0 {
			// Several Ingress
			return regv1.MakeRegistryError("NotReady")
		}
		reg.Status.LoadBalancerIP = lbIP
		r.logger.Info("LoadBalancer info", "LoadBalancer IP", lbIP)
	} else if r.svc.Spec.Type == corev1.ServiceTypeClusterIP {
		if r.svc.Spec.ClusterIP == "" {
			return regv1.MakeRegistryError("NotReady")
		}
		r.logger.Info("Service Type is ClusterIP(Ingress)")
		// [TODO]
	}
	reg.Status.ClusterIP = r.svc.Spec.ClusterIP
	condition.Status = corev1.ConditionTrue
	err = nil
	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryService) create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {
	condition := &status.Condition {
		Status: corev1.ConditionFalse,
		Type: ServiceTypeName,
	}

	if err := controllerutil.SetControllerReference(reg, r.svc, scheme); err != nil {
		r.logger.Error(err, "Set owner reference failed")
		utils.SetError(err, patchReg, condition)
		return err
	}

	if err := c.Create(context.TODO(), r.svc); err != nil {
		r.logger.Error(err, "Create failed")
		utils.SetError(err, patchReg, condition)
		return err
	}

	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryService) get(c client.Client, reg *regv1.Registry) error {
	if r.svc == nil {
		r.svc = schemes.Service(reg)
		r.logger = utils.NewRegistryLogger(*r, r.svc.Namespace, r.svc.Name)
	}

	req := types.NamespacedName{Name: r.svc.Name, Namespace: r.svc.Namespace}
	if err := c.Get(context.TODO(), req, r.svc); err != nil {
		r.logger.Error(err, "Get Failed")
		return err
	}
	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryService) patch(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, diff []utils.Diff) error {
	// [TODO]
	return nil
}

func (r *RegistryService) delete(c client.Client, patchReg *regv1.Registry) error {
	condition := &status.Condition {
		Status: corev1.ConditionFalse,
		Type: ServiceTypeName,
	}

	if err := c.Delete(context.TODO(), r.svc); err != nil {
		r.logger.Error(err, "Delete failed")
		utils.SetError(err, patchReg, condition)
		return err
	}

	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryService) compare(reg *regv1.Registry) []utils.Diff {
	regServiceSpec := reg.Spec.RegistryService
	if string(regServiceSpec.ServiceType) == "Ingress" && string(r.svc.Spec.Type) != "ClusterIP" {
		r.logger.Error(regv1.MakeRegistryError("Type is different"), "Service Type is different")
		return nil
	}

	isPortValid := false
	for _, port := range r.svc.Spec.Ports {
		if (regServiceSpec.ServiceType == regv1.RegServiceTypeLoadBalancer &&
			regServiceSpec.LoadBalancer.Port == int(port.Port)) ||
			(string(regServiceSpec.ServiceType) == "Ingress" && int(port.Port) == 443) {
			isPortValid = true
		}
	}

	if !isPortValid {
		r.logger.Error(regv1.MakeRegistryError("Port is invalid"), "Port is not valid")
		return nil
	}

	r.logger.Info("Succeed")
	return []utils.Diff{}
}



