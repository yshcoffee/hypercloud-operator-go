package regctl

import (
	"context"
	"github.com/go-logr/logr"
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
	logger logr.Logger
}

func (r *RegistryService) Create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme, useGet bool) error {
	condition := status.Condition {
		Status: corev1.ConditionFalse,
		Type: ServiceTypeName,
	}

	if useGet {
		err := r.get(c, reg)
		if err != nil && !errors.IsNotFound(err) {
			r.logger.Error(err, "Getting Service failed")
			utils.SetError(err, patchReg, condition)
			return err
		} else if err == nil {
			r.logger.Info("Service already exist")
			return err
		}
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
		r.logger = utils.GetRegistryLogger(*r, r.svc.Namespace, r.svc.Name)
	}

	req := types.NamespacedName{Name: r.svc.Name, Namespace: r.svc.Namespace}
	if err := c.Get(context.TODO(), req, r.svc); err != nil {
		r.logger.Error(err, "Get Failed")
		return err
	}
	r.logger.Info("Succeed")
	return nil
}

func (r *RegistryService) Patch(c client.Client, reg *regv1.Registry, json []byte) error {
	// [TODO]
	return nil
}

func (r *RegistryService) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	condition := status.Condition {
		Status: corev1.ConditionFalse,
		Type: ServiceTypeName,
	}

	if useGet {
		if err := r.get(c, reg); err != nil {
			r.logger.Error(err, "Getting Service error")
			utils.SetError(err, patchReg, condition)
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
			// [TODO] Is this error?
			utils.SetError(nil, patchReg, condition)
			return regv1.MakeRegistryError("NotReady")
		}
		reg.Spec.RegistryService.LoadBalancer.IP = lbIP
		r.logger.Info("LoadBalancer info", "LoadBalancer IP", lbIP)
	} else if r.svc.Spec.Type == corev1.ServiceTypeClusterIP {
		r.logger.Info("Service Type is ClusterIp")
		// [TODO]
	}
	reg.Spec.RegistryService.ClusterIP = r.svc.Spec.ClusterIP
	r.logger.Info("Succeed Info", "LoadBalancerIP", reg.Spec.RegistryService.LoadBalancer.IP, "ClusterIP", reg.Spec.RegistryService.ClusterIP)
	return nil
}

