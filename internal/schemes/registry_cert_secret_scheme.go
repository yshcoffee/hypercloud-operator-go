package schemes

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"math/big"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
	"time"
)

const (
	RootCASecretName = "registry-ca"
	RootCASecretNamespace = "hypercloud4-system"
	RootCACert = "ca.crt"
	RootCAPriv = "ca.key"
	CertKeyFile = "localhub.key"
	CertCrtFile = "localhub.crt"
	TLSCert = "tls.crt"
	TLSKey = "tls.key"
)

func Secrets(reg *regv1.Registry, c client.Client) (*corev1.Secret, *corev1.Secret) {
	logger := utils.GetRegistryLogger(corev1.Secret{}, reg.Namespace, reg.Name + "secret")
	if (!regBodyCheckForSecrets(reg)) {
		return nil, nil
	}
	secretType := corev1.SecretTypeOpaque
	serviceType := reg.Spec.RegistryService.ServiceType
	port := 443
	if serviceType == regv1.RegServiceTypeLoadBalancer {
		port = reg.Spec.RegistryService.LoadBalancer.Port
	}
	data := map[string][]byte{}
	data["ID"] = []byte(reg.Spec.LoginId)
	data["PASSWD"] = []byte(reg.Spec.LoginPassword)
	data["CLUSTER_IP"] = []byte(reg.Status.ClusterIP)

	if serviceType == "Ingress" {
		registryDomainName := reg.Name +  "." + reg.Spec.RegistryService.Ingress.DomainName
		data["DOMAIN_NAME"] = []byte(registryDomainName)
		data["REGISTRY_URL"] = []byte(registryDomainName + ":" + strconv.Itoa(port))
	} else if serviceType == regv1.RegServiceTypeLoadBalancer {
		data["LB_IP"] = []byte(reg.Status.LoadBalancerIP)
		data["REGISTRY_URL"] = []byte(reg.Status.LoadBalancerIP + ":" + strconv.Itoa(port))
	} else {
		data["REGISTRY_URL"] = []byte(reg.Status.ClusterIP + ":" + strconv.Itoa(port))
	}

	parentCert, parentPrivKey := getCertificateFromFile(c)

	certificateBytes, privateKey, err := makeCertificate(reg, parentCert, parentPrivKey)
	if err != nil {
		// ERROR
		logger.Error(err, "Create certificate failed")
		return nil, nil
	}
	logger.Info("Create Certificate Succeed")
	data[CertCrtFile] = certificateBytes // have to do parse
	privateBytes, _ := x509.MarshalPKCS8PrivateKey(privateKey) // have to do unmarshal

	data[CertKeyFile] = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateBytes})

	logger.Info("Create Secret Opaque Succeed")

	tlsSecretType := corev1.SecretTypeTLS
	tlsData := map[string][]byte{}
	tlsData[TLSCert] = data[CertCrtFile]
	tlsData[TLSKey] = data[CertKeyFile]

	logger.Info("Create Secret TLS Succeed")



	return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: regv1.K8sPrefix + strings.ToLower(reg.Name),
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
				Name: regv1.K8sPrefix + regv1.TLSPrefix + strings.ToLower(reg.Name),
				Namespace: reg.Namespace,
				Labels: map[string]string {
					"secret": "tls",
				},
			},
			Type: tlsSecretType,
			Data: tlsData,
		}
}

func getCertificateFromFile(c client.Client) (*x509.Certificate, *rsa.PrivateKey) {
	logger := utils.GetRegistryLogger(corev1.Secret{}, "CertScheme", "secret")

	rootSecret := corev1.Secret{}
	req := types.NamespacedName{Name: RootCASecretName, Namespace: RootCASecretNamespace}
	if err := c.Get(context.TODO(), req, &rootSecret); err != nil {
		logger.Error(err, "Get Root Secret Error")
		return nil, nil
	}

	block, rest := pem.Decode(rootSecret.Data[RootCACert])
	if len(rest) != 0 {
		logger.Info("Cert is not PEM format", "Rest", rest)
		return nil, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		logger.Error(err, "Parse Root CA block Error")
		return nil, nil
	}

	privBlock, privRest := pem.Decode(rootSecret.Data[RootCAPriv])
	if len(privRest) != 0 {
		logger.Info("Private key is not PEM format", "Rest", privRest)
		return nil, nil
	}

	key, privKeyErr := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if privKeyErr != nil {
		logger.Error(privKeyErr, "Parse private key Error")
		return nil, nil
	}

	return cert, key.(*rsa.PrivateKey)
}

// [TODO] Logging
func makeCertificate(reg *regv1.Registry, parentCert *x509.Certificate,
	parentPrivateKey *rsa.PrivateKey) ([]byte, *rsa.PrivateKey, error) {
	logger := utils.GetRegistryLogger(corev1.Secret{}, "CertScheme", "secret")
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
			CommonName: reg.Status.ClusterIP,
		},
		NotBefore: time.Now(),
		NotAfter: time.Now().Add(time.Hour * 24 * 1000),

		KeyUsage: x509.KeyUsageCRLSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA: false,
		BasicConstraintsValid: true,
	}

	template.IPAddresses = []net.IP{net.ParseIP(reg.Status.ClusterIP)}
	if reg.Spec.RegistryService.ServiceType == regv1.RegServiceTypeLoadBalancer  {
		template.IPAddresses = append(template.IPAddresses, net.ParseIP(reg.Status.LoadBalancerIP))
	} else if reg.Spec.RegistryService.ServiceType == regv1.RegServiceTypeIngress {
		template.DNSNames = []string{reg.Spec.RegistryService.Ingress.DomainName}
	}

	parent := &x509.Certificate{}
	parentPrivKey := &rsa.PrivateKey{}
	if parentCert == nil || parentPrivateKey == nil{
		logger.Info("There is no parent")
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

	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes});

	_, erro := x509.ParseCertificate(serverCertBytes)
	if erro != nil {
		return nil, nil, err
	}
	//utils.NewRegistryLogger(regv1.Registry{}, reg.Namespace, reg.Name).Info("Cert Test", "Cert", string(serverCertPEM), "Private", privateKey)

	return serverCertPEM, privateKey, nil
}

func regBodyCheckForSecrets(reg *regv1.Registry) bool {
	regService := reg.Spec.RegistryService
	if (reg.Status.ClusterIP == "") {
		return false
	}
	if (regService.ServiceType == regv1.RegServiceTypeLoadBalancer && reg.Status.LoadBalancerIP == "" ) {
		return false
	} else if (regService.ServiceType == regv1.RegServiceTypeIngress && regService.Ingress.DomainName == "") {
		return false
	}
	return true
}
