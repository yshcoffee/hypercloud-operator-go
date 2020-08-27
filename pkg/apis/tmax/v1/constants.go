package v1

const (
	K8sPrefix         = "hpcd-"
	OperatorNamespace = "hypercloud4-system"
	TLSPrefix = "tls-"
	K8sRegistryPrefix = "registry-"

	CustomObjectGroup = "tmax.io"

	// OpenSSL Cert File Name
	GenCertScriptFile = "genCert.sh"
	CertKeyFile       = "localhub.key"
	CertCrtFile       = "localhub.crt"
	CertCertFile      = "localhub.cert"
	DockerDir         = "/etc/docker"
	DockerCertDir     = "/etc/docker/certs.d"

	// OpenSSL Certificate Home Directory
	OpenSslHomeDir = "/openssl"

	DockerLoginHomeDir   = "/root/.docker"
	DockerConfigFile     = "config.json"
	DockerConfigJsonFile = ".dockerconfigjson"
)
