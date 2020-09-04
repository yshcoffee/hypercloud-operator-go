package schemes

import (
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Ingress(reg *regv1.Registry) (*v1beta1.Ingress) {
	if (!regBodyCheckForIngress(reg)) {
		return nil
	}
	registryDomain := reg.Name + "." + reg.Spec.RegistryService.Ingress.DomainName

	ingressTLS := v1beta1.IngressTLS{
		Hosts: []string{registryDomain},
		SecretName: regv1.K8sPrefix + regv1.TLSPrefix + reg.Name,
	}
	httpIngressPath := v1beta1.HTTPIngressPath{
		Path: "/",
		Backend: v1beta1.IngressBackend{
			ServiceName: regv1.K8sPrefix + reg.Name,
			ServicePort: intstr.FromInt(443),
		},
	}

	rule := v1beta1.IngressRule{
		Host: registryDomain,
		IngressRuleValue: v1beta1.IngressRuleValue{
			HTTP: &v1beta1.HTTPIngressRuleValue{
				Paths: []v1beta1.HTTPIngressPath {
					httpIngressPath,
				},
			},
		},
	}

	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: regv1.K8sPrefix + reg.Name,
			Namespace: reg.Namespace,
			Labels: map[string]string {
				"app" : "registry",
				"apps" : regv1.K8sPrefix + reg.Name,
			},
			Annotations: map[string]string {
				"kubernetes.io/ingress.class" : "nginx-shd",
				"nginx.ingress.kubernetes.io/proxy-connect-timeout" : "3600",
				"nginx.ingress.kubernetes.io/proxy-read-timeout" : "3600",
				"nginx.ingress.kubernetes.io/ssl-redirect" : "true",
				"nginx.ingress.kubernetes.io/backend-protocol" : "HTTPS",
				"nginx.ingress.kubernetes.io/proxy-body-size" : "0",
			},
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS {
				ingressTLS,
			},
			Rules: []v1beta1.IngressRule {
				rule,
			},
		},
	}
}

func regBodyCheckForIngress(reg *regv1.Registry) bool {
	regService := reg.Spec.RegistryService
	if (regService.ServiceType != "Ingress") {
		return false
	}
	if (regService.Ingress.DomainName == "") {
		return false
	}
	if (reg.Status.ClusterIP == "") {
		return false
	}
	return true
}

