package regctl

import (
	"context"
	"hypercloud-operator-go/internal/schemes"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RegistryPVC struct {
	pvc *corev1.PersistentVolumeClaim
}

func (r *RegistryPVC) GetTypeName() string {
	return regv1.ConditionTypePvc
}

func (r *RegistryPVC) Patch(c client.Client, registry *regv1.Registry) error {
	return nil
}

func (r *RegistryPVC) Update(c client.Client, registry *regv1.Registry) error {
	return nil
}

func (r *RegistryPVC) StatusPatch(c client.Client, registry *regv1.Registry, condition *regv1.RegistryCondition) error {
	return nil
}


func (r *RegistryPVC) Get(client client.Client, reg *regv1.Registry, condition *regv1.RegistryCondition) error {
	return nil
}

func (r *RegistryPVC) Create(client client.Client, reg *regv1.Registry, condition *regv1.RegistryCondition) error {
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

func (r *RegistryPVC) StatusUpdate(client client.Client, reg *regv1.Registry, condition *regv1.RegistryCondition) error {
	reqLogger := log.Log.WithValues("RegistryPVC.Namespace", reg.Namespace, "RegistryPVC.Name", reg.Name)
	conditions := reg.Status.Conditions

	/*
	condition := regv1.RegistryCondition{
		LastTransitionTime: metav1.Now(),
	}

	conditions[regv1.ConditionOrd[regv1.ConditionTypePvc]] = condition
	*/
	reg.Status.Conditions = conditions

	err := client.Status().Update(context.TODO(), reg)
	if err != nil {
		reqLogger.Error(err, "Unknown error updating status")
		return err
	}

	return nil
}

