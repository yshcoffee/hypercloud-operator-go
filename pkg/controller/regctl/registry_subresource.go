package regctl

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmaxv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

type RegistrySubresource interface {
	Get(client.Client, *tmaxv1.Registry) error
	Create(client.Client, *tmaxv1.Registry) error
	Ready(*tmaxv1.Registry) bool
	Patch(client.Client, *tmaxv1.Registry) error
	Update(client.Client, *tmaxv1.Registry) error
	StatusPatch(client.Client, *tmaxv1.Registry) error
	StatusUpdate(client.Client, *tmaxv1.Registry) error
}

