package utils

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func CheckAndCreateObject(client client.Client, namespacedName types.NamespacedName, obj runtime.Object) error {
	resourceType := reflect.TypeOf(obj).String()
	reqLogger := log.Log.WithValues(resourceType+".Namespace", namespacedName.Namespace, resourceType+".Name", namespacedName.Name)

	err := client.Get(context.TODO(), namespacedName, obj)
	if err != nil && k8serrors.IsNotFound(err) {
		reqLogger.Info("Creating")
		if err = client.Create(context.TODO(), obj); err != nil {
			reqLogger.Error(err, "Error creating")
			return err
		}
	} else if err != nil {
		reqLogger.Error(err, "Error getting status")
		return err
	} else {
		reqLogger.Info("Already Exist")
	}
	return nil
}

type Patcher struct {
	PatchType types.PatchType
	DataBytes []byte
}

func (p *Patcher) Type() types.PatchType {
	return p.PatchType
}

func (p *Patcher) Data(obj runtime.Object) ([]byte, error) {
	return p.DataBytes, nil
}
