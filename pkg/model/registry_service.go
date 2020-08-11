package model

import (
	"context"
	tmaxv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RegistryService struct {
	svc corev1.Service
}

func (r *RegistryService) Get(client client.Client, reg *tmaxv1.Registry) error {
	req := types.NamespacedName{Name: reg.Name, Namespace: reg.Namespace}
	svc := &corev1.Service{}
	err := client.Get(context.TODO(), req, svc)
	if err != nil {
		return err
	}

	return nil
}

func (r *RegistryService) Create(client client.Client, reg *tmaxv1.Registry) error {

	return nil
}

func (r *RegistryService) Ready(reg *tmaxv1.Registry) bool {
	// if reg.Spec.RegistryService.LoadBalancer == nil {

	// }
	if len(r.svc.Status.LoadBalancer.Ingress) == 0 {
		return false
	}

	return true
}

func (r *RegistryService) Update(client client.Client, reg *tmaxv1.Registry) error {
	reqLogger := log.Log.WithValues("RegistryService.Namespace", reg.Namespace, "RegistryService.Name", reg.Name)
	err := client.Status().Update(context.TODO(), reg)
	if err != nil {
		reqLogger.Error(err, "Unknown error updating status")
		return err
	}

	return nil
}
