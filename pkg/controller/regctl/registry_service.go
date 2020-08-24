package regctl

import (
	"context"
	"hypercloud-operator-go/internal/schemes"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/operator-framework/operator-sdk/pkg/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const ServiceTypeName = regv1.ConditionTypeService

type RegistryService struct {
	svc *corev1.Service
	//[TODO] Logging
}

func (r *RegistryService) Create(c client.Client, reg *regv1.Registry, condition *status.Condition, scheme *runtime.Scheme, useGet bool) error {
	if r.svc == nil {
		r.svc = schemes.Service(reg)
		if err := controllerutil.SetControllerReference(reg, r.svc, scheme); err != nil {
			return err
		}
	}
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.Api", "Create")

	if useGet {
		err := r.get(c, reg, condition)
		if err != nil && !errors.IsNotFound(err) {
			serviceLogger.Error(err, "Getting Service failed")
			return err
		} else if err == nil {
			serviceLogger.Info("Service already exist")
			return err
		}
	}

	if err := c.Create(context.TODO(), r.svc); err != nil {
		serviceLogger.Error(err, "Create Failed")
		condition.Status = corev1.ConditionFalse
		condition.Message = err.Error()
		return err
	}

	serviceLogger.Info("Succeed")
	return nil
}

func (r *RegistryService) get(c client.Client, reg *regv1.Registry, condition *status.Condition) error {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "get")
	req := types.NamespacedName{Name: r.svc.Name, Namespace: r.svc.Namespace}
	if err := c.Get(context.TODO(), req, r.svc); err != nil {
		serviceLogger.Error(err, "Get Failed")
		condition.Message = err.Error()
		return err
	}
	serviceLogger.Info("Succeed")
	return nil
}

func (r *RegistryService) GetTypeName() string {
	return string(ServiceTypeName)
}

func (r *RegistryService) Patch(c client.Client, reg *regv1.Registry, useGet bool) error {
	// [TODO]
	return nil
}

func (r *RegistryService) Ready(reg *regv1.Registry, useGet bool) error {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "Ready")
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
		} else {
			// Several Ingress
			return regv1.MakeRegistryError(regv1.NotReady)
		}
		reg.Spec.RegistryService.LoadBalancer.IP = lbIP
		serviceLogger.Info("LoadBalancer info", "LoadBalancer IP", lbIP)
	} else if r.svc.Spec.Type == corev1.ServiceTypeClusterIP {
		serviceLogger.Info("Service Type is ClusterIp")
		// [TODO]
	}
	reg.Spec.RegistryService.ClusterIP = r.svc.Spec.ClusterIP
	serviceLogger.Info("Succeed")
	return regv1.MakeRegistryError(regv1.Running)
}

func (r *RegistryService) StatusPatch(c client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "StatusPatch")

	if useGet {
		if err := r.get(c, reg, condition); (err != nil && !errors.IsNotFound(err)) || err == nil {
			serviceLogger.Error(err, "Getting Service failed")
			return err
		}
	}

	patch := client.MergeFrom(reg)
	target := reg.DeepCopy()
	target.Status.Conditions.SetCondition(*condition)

	if err := c.Status().Patch(context.TODO(), target, patch); err != nil {
		serviceLogger.Error(err, "StatusPatch Failed")
		return err
	}

	serviceLogger.Info("Succeed")
	return nil
}

func (r *RegistryService) StatusUpdate(c client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "StatusUpdate")

	// [TODO] Why this condition needed?
	if useGet {
		if err := r.get(c, reg, condition); (err != nil && !errors.IsNotFound(err)) || err == nil {
			serviceLogger.Error(err, "Getting Service failed")
			return err
		}
	}

	registryCondition := reg.Status.Conditions.GetCondition(ServiceTypeName)
	condition.DeepCopyInto(registryCondition)
	if err := c.Status().Update(context.TODO(), reg); err != nil {
		serviceLogger.Error(err, "Update Failed")
		return err
	}

	return nil
}

func (r *RegistryService) Update(c client.Client, reg *regv1.Registry, useGet bool) error {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "Update")

	// [TODO] Is this Spec update?
	if err := c.Status().Update(context.TODO(), reg); err != nil {
		serviceLogger.Error(err, "Update Failed")
		return err
	}
	serviceLogger.Info("Succeed")
	return nil
}
