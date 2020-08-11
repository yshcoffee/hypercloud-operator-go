package model

import (
	"context"
	"hypercloud-operator-go/internal/schemes"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RegistryPVC struct {
	pvc *corev1.PersistentVolumeClaim
}

func (r *RegistryPVC) Get(client client.Client, reg *regv1.Registry) error {
	req := types.NamespacedName{Name: reg.Name, Namespace: reg.Namespace}
	r.pvc = &corev1.PersistentVolumeClaim{}
	err := client.Get(context.TODO(), req, r.pvc)
	if err != nil {
		return err
	}

	return nil
}

func (r *RegistryPVC) Create(client client.Client, reg *regv1.Registry) error {
	r.pvc = schemes.PersistentVolumeClaim(reg)
	client.Create(context.TODO(), r.pvc)
	return nil
}

func (r *RegistryPVC) Ready(reg *regv1.Registry) bool {
	if string(r.pvc.Status.Phase) == "pending" {
		return false
	}

	return true
}

func (r *RegistryPVC) StatusUpdate(client client.Client, reg *regv1.Registry, status, message, reason string) error {
	reqLogger := log.Log.WithValues("RegistryPVC.Namespace", reg.Namespace, "RegistryPVC.Name", reg.Name)
	conditions := reg.Status.Conditions

	condition := regv1.RegistryCondition{
		LastTransitionTime: metav1.Now(),
		Message:            message,
		Status:             status,
		Reason:             reason,
	}

	conditions[regv1.ConditionOrd[regv1.ConditionTypePvc]] = condition
	reg.Status.Conditions = conditions

	err := client.Status().Update(context.TODO(), reg)
	if err != nil {
		reqLogger.Error(err, "Unknown error updating status")
		return err
	}

	return nil
}
