package regctl

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

type RegistrySubresource interface {
	Handle(client.Client, *regv1.Registry, *regv1.Registry, *runtime.Scheme) error
	Ready(client.Client, *regv1.Registry, *regv1.Registry, bool) error

	create(client.Client, *regv1.Registry, *regv1.Registry, *runtime.Scheme, bool) error
	get(client.Client, *regv1.Registry) error
	patch(client.Client, *regv1.Registry, *regv1.Registry, []utils.Diff) error
	delete(client.Client, *regv1.Registry, *regv1.Registry, bool) error
	compare(client.Client, *regv1.Registry, bool) ([]utils.Diff, bool)
}
