package regctl

import (
	"github.com/r3labs/diff"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

type RegistrySubresource interface {
	Handle(client.Client, *regv1.Registry, *regv1.Registry, *runtime.Scheme, diff.Changelog, bool) error
	Ready(client.Client, *regv1.Registry, *regv1.Registry, bool) error

	create(client.Client, *regv1.Registry, *regv1.Registry, *runtime.Scheme, bool) error
	get(client.Client, *regv1.Registry) error
	patch(client.Client, *regv1.Registry, diff.Changelog) error
	delete(client.Client, *regv1.Registry, *regv1.Registry, bool) error
}
