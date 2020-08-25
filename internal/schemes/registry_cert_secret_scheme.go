package schemes

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/big"
	"net"
	"time"
)

const (
	CertKeyFile = "localhub.key"
	CertCrtFile = "localhub.crt"
	TLSCert = "tls.crt"
	TLSKey = "tls.key"
)

func Secrets(reg *regv1.Registry) (*corev1.Secret, *corev1.Secret) {
	logger := utils.GetRegistryLogger(corev1.Secret{}, reg.Namespace, reg.Name + "secret")
	secretType := corev1.SecretTypeOpaque
	secretName := regv1.K8sPrefix + reg.Name
	serviceType := reg.Spec.RegistryService.ServiceType
	port := reg.Spec.RegistryService.Port
	data := map[string][]byte{}
	data["ID"] = []byte(reg.Spec.LoginId)
	data["PASSWD"] = []byte(reg.Spec.LoginPassword)
	data["CLUSTER_IP"] = []byte(reg.Spec.RegistryService.ClusterIP)

	if serviceType == regv1.RegServiceTypeIngress {
		registryDomainName := reg.Name +  "." + reg.Spec.RegistryService.Ingress.DomainName
		data["DOMAIN_NAME"] = []byte(registryDomainName)
		data["REGISTRY_URL"] = []byte(registryDomainName + ":" + string(port))
	} else if serviceType == regv1.RegServiceTypeLoadBalancer {
		data["LB_IP"] = []byte(reg.Spec.RegistryService.LoadBalancer.IP)
		data["REGISTRY_URL"] = []byte(reg.Spec.RegistryService.LoadBalancer.IP + ":" + string(port))
	} else {
		data["REGISTRY_URL"] = []byte(reg.Spec.RegistryService.ClusterIP + ":" + string(port))
	}

	// parentCert, parentPrivateKey == nil ==> Self Signed Certificate
	certificateBytes, privateKey, err := makeCertificate(reg, nil, nil)
	if err != nil {
		// ERROR
		logger.Error(err, "Create certificate failed")
		return nil, nil
	}
	logger.Info("Create Certificate Succeed")
	data[CertCrtFile] = certificateBytes // have to do parse
	data[CertKeyFile] = x509.MarshalPKCS1PrivateKey(privateKey) // have to do unmarshal

	logger.Info("Create Secret Opaque Succeed")

	tlsSecretType := corev1.SecretTypeTLS
	tlsData := map[string][]byte{}
	tlsData[TLSCert] = data[CertCrtFile]
	tlsData[TLSKey] = data[CertKeyFile]

	logger.Info("Create Secret TLS Succeed")



	return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: secretName,
				Namespace: reg.Namespace,
				Labels: map[string]string {
					"secret": "cert",
				},
			},
			Type: secretType,
			Data: data,
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta {
				Name: reg.Name,
				Namespace: reg.Namespace,
				Labels: map[string]string {
					"secret": "tls",
				},
			},
			Type: tlsSecretType,
			Data: tlsData,
		}
}

// [TODO] Logging
func makeCertificate(reg *regv1.Registry, parentCert *x509.Certificate,
	parentPrivateKey *rsa.PrivateKey) ([]byte, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country: []string{"KR"},
			Organization: []string{"tmax"},
			StreetAddress: []string{"Seoul"},
			CommonName: reg.Spec.RegistryService.ClusterIP,
		},
		NotBefore: time.Now(),
		NotAfter: time.Now().Add(time.Hour * 24 * 1000),

		KeyUsage: x509.KeyUsageCRLSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA: false,
		BasicConstraintsValid: true,
	}

	template.IPAddresses = []net.IP{net.ParseIP(reg.Spec.RegistryService.ClusterIP)}
	if reg.Spec.RegistryService.ServiceType == regv1.RegServiceTypeLoadBalancer  {
		_ = append(template.IPAddresses, net.ParseIP(reg.Spec.RegistryService.LoadBalancer.IP))
	} else if reg.Spec.RegistryService.ServiceType == regv1.RegServiceTypeIngress {
		template.DNSNames = []string{reg.Spec.RegistryService.Ingress.DomainName}
	}

	parent := &x509.Certificate{}
	parentPrivKey := &rsa.PrivateKey{}
	if parentCert == nil && parentPrivateKey == nil{
		parent = &template
		parentPrivKey = privateKey
	} else {
		parent = parentCert
		parentPrivKey = parentPrivateKey
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, &template, parent, &privateKey.PublicKey, parentPrivKey)
	if err != nil {
		return nil, nil, err
	}

	if _, err = x509.ParseCertificate(serverCertBytes); err != nil {
		return nil, nil, err
	}

	return serverCertBytes, privateKey, nil
}
