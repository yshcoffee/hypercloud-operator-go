package regctl

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
	"sigs.k8s.io/controller-runtime/pkg/client"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

type RegistrySubresource interface {
	GetTypeName() string
	Get(client.Client, *regv1.Registry, *status.Condition) error
	Create(client.Client, *regv1.Registry, *status.Condition) error
	Ready(*regv1.Registry) bool
	Patch(client.Client, *regv1.Registry) error
	Update(client.Client, *regv1.Registry) error
	StatusPatch(client.Client, *regv1.Registry, *status.Condition) error
	StatusUpdate(client.Client, *regv1.Registry, *status.Condition) error
}
