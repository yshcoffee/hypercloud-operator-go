package regctl

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

type RegistrySubresource interface {
	Create(client.Client, *regv1.Registry, *status.Condition, *runtime.Scheme, bool) error
	//Get(client.Client, *regv1.Registry, *status.Condition) error
	get(client.Client, *regv1.Registry, *status.Condition) error
	GetTypeName() string
	Patch(client.Client, *regv1.Registry, bool) error
	// [TODO] If not Ready
	Ready(*regv1.Registry, bool) error
	StatusPatch(client.Client, *regv1.Registry, *status.Condition, bool) error
	StatusUpdate(client.Client, *regv1.Registry, *status.Condition, bool) error
	Update(client.Client, *regv1.Registry, bool) error
}
