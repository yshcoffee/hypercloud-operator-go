package regctl

import (
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/operator-framework/operator-sdk/pkg/status"
	"sigs.k8s.io/controller-runtime/pkg/client"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

type RegistrySubresource interface {
	Create(client.Client, *regv1.Registry, *regv1.RegistryCondition) error
	Get(client.Client, *regv1.Registry, *regv1.RegistryCondition) error
	GetTypeName() string
	Patch(client.Client, *regv1.Registry) error
	Ready(*regv1.Registry) bool
	SetOwnerReference(*regv1.Registry, *runtime.Scheme) error
	StatusPatch(client.Client, *regv1.Registry, *status.Condition) error
	StatusUpdate(client.Client, *regv1.Registry, *status.Condition) error
	Update(client.Client, *regv1.Registry) error
}
