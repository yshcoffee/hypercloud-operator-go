package regctl

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

type RegistrySubresource interface {
	Create(client.Client, *regv1.Registry, *regv1.Registry, *runtime.Scheme, bool) error
	//Get(client.Client, *regv1.Registry, *status.Condition) error
	get(client.Client, *regv1.Registry) error
	// GetTypeName() string
	Patch(client.Client, *regv1.Registry, []byte) error
	// RegistryPatch(client.Client, *regv1.Registry, *regv1.Registry, bool) error
	// [TODO] If not Ready
	Ready(client.Client, *regv1.Registry, *regv1.Registry, bool) error
	// StatusPatch(client.Client, *regv1.Registry, *regv1.Registry, bool) error
	// StatusUpdate(client.Client, *regv1.Registry, *regv1.Registry, bool) error
	// Update(client.Client, *regv1.Registry, interface{}, bool) error
}

