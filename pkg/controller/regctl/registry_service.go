package regctl

import (
	"context"
	"hypercloud-operator-go/internal/schemes"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

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

func (r *RegistryService) Create(client client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "Create")
	r.svc = schemes.Service(reg)
	if err := client.Create(context.TODO(), r.svc); err != nil {
		serviceLogger.Error(err, "Create Failed")
		condition.Message = err.Error()
		return err
	}
	return nil
}

func (r *RegistryService) get(client client.Client, reg *regv1.Registry, condition *status.Condition) error {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "Get")
	req := types.NamespacedName{Name: reg.Name, Namespace: reg.Namespace}
	if err := client.Get(context.TODO(), req, r.svc); err != nil {
		serviceLogger.Error(err, "Get Failed")
		condition.Message = err.Error()
	}
	return nil
}

func (r *RegistryService) GetTypeName() string {
	return ServiceTypeName
}

func (r *RegistryService) Patch(client client.Client, reg *regv1.Registry, useGet bool) error {
	return nil
}

func (r *RegistryService) Ready(reg *regv1.Registry, useGet bool) bool {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "Ready")
	if r.svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		loadBalancer := r.svc.Status.LoadBalancer
		lbIP := ""
		// [TODO] Specific Condition is needed
		if len(loadBalancer.Ingress) == 1 {
			if loadBalancer.Ingress[0].Hostname  == "" {
				lbIP = loadBalancer.Ingress[0].IP
			} else {
				lbIP = loadBalancer.Ingress[0].Hostname
			}
		} else {
			// Several Ingress
			return false
		}
		serviceLogger.Info("LoadBalancer IP", lbIP)
	} else if r.svc.Spec.Type == corev1.ServiceTypeClusterIP {
		serviceLogger.Info("Service Type is ClusterIp")
		// [TODO]
	}

	return true
}

func (r *RegistryService) SetOwnerReference(reg *regv1.Registry, scheme *runtime.Scheme, useGet bool) error {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "SetOwnerReference")
	if err := controllerutil.SetControllerReference(reg, r.svc, scheme); err != nil {
		serviceLogger.Error(err, "SetOwnerReference Failed")
		return err
	}
	return nil
}

func (r *RegistryService) StatusPatch(c client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "StatusPatch")
	patch := client.MergeFrom(reg)
	target := reg.DeepCopy()
	target.Status.Conditions.SetCondition(*condition)

	if err := c.Status().Patch(context.TODO(), target, patch); err != nil {
		serviceLogger.Error(err, "StatusPatch Failed")
		return err
	}
	return nil
}

func (r *RegistryService) StatusUpdate(c client.Client, reg *regv1.Registry, condition *status.Condition, useGet bool) error {
	return nil
}

func (r *RegistryService) Update(c client.Client, reg *regv1.Registry, useGet bool) error {
	serviceLogger := log.Log.WithValues("RegistryService.Namespace", r.svc.Namespace,
		"RegistryService.Name", r.svc.Name, "RegistryService.API", "Update")
	if err := c.Status().Update(context.TODO(), reg); err != nil {
		serviceLogger.Error(err, "Update Failed")
		return err
	}
	return nil
}
