package schemes

import (
	"crypto/rand"
	"crypto/rsa"
	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	//"crypto/x509"
)

func SecretOpaque(reg *regv1.Registry) *corev1.Secret {
	secretName := regv1.K8sPrefix + reg.Name
	serviceType := LoadBalancer
	port := 0
	registryDomainName := reg.Name
	labels := utils.GetLabel(reg)
	labels["secret"] = "cert"
	ingress := reg.Spec.RegistryService.Ingress
	if ingress != nil {
		serviceType = Ingress
		port = ingress.Port
		if ingress.DomainName != "" {
			registryDomainName = reg.Name + "." + ingress.DomainName
		}
	} else {
		port = reg.Spec.RegistryService.LoadBalancer.Port
	}

	secretType := corev1.SecretTypeOpaque
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Namespace: reg.Namespace,
			Labels: labels,
		},
		Type: secretType,
	}
}

func makeCertificate(reg *regv1.Registry) error {
	return nil
}

func CreateCertDirectory(reg *regv1.Registry) error {
	path := regv1.OpenSslHomeDir + "/" + reg.Namespace + "/" + reg.Name
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	return nil
}
