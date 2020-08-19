package schemes

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

const (
	RegistryTargetPort   = 443
	RegistryPortProtocol = "TCP"
	RegistryPortName     = "tls"

	Ingress = "Ingress"
	LoadBalancer = "LoadBalancer"
)

func Service(reg *regv1.Registry) *corev1.Service {
	regServiceName := regv1.K8sPrefix + reg.Name
	label := utils.GetLabel(reg)
	label["app"] = "registry"
	label["apps"] = regv1.K8sPrefix + reg.Name
	port := RegistryTargetPort
	serviceName := Ingress
	if reg.Spec.RegistryService.Ingress == nil {
		serviceName = LoadBalancer
	}

	if serviceName == Ingress {
		if reg.Spec.RegistryService.Ingress.Port != 0 {
			port = reg.Spec.RegistryService.Ingress.Port
		}
	} else {
		if reg.Spec.RegistryService.LoadBalancer.Port != 0 {
			port = reg.Spec.RegistryService.LoadBalancer.Port
		}
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      regServiceName,
			Namespace: reg.Namespace,
			Labels:    label,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceType(serviceName),
			Selector: map[string]string{
				regv1.K8sPrefix + reg.Name: "lb",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     RegistryPortName,
					Protocol: RegistryPortProtocol,
					Port:     int32(port),
				},
			},
		},
	}
}
