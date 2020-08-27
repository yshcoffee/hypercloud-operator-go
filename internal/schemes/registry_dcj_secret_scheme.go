package schemes

import (
	"encoding/base64"
	"encoding/json"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

const (
	DockerConfigJson = ".dockerconfigjson"
)

type DockerConfig struct {
	Auths map[string]AuthValue `json:"auths"`
}

type AuthValue struct {
	Auth string `json:"auth"`
}

func DCJSecret(reg *regv1.Registry) *corev1.Secret {
	if (!regBodyCheckForDCJSecret(reg)) {
		return nil
	}
	serviceType := reg.Spec.RegistryService.ServiceType
	var domainList []string
	port := reg.Spec.RegistryService.Port
	data := map[string][]byte{}
	domainList = append(domainList, reg.Spec.RegistryService.ClusterIP + ":" + strconv.Itoa(port))
	if serviceType == regv1.RegServiceTypeLoadBalancer {
		domainList = append(domainList, reg.Spec.RegistryService.LoadBalancer.IP + ":" + strconv.Itoa(port))
	} else {
		domainList = append(domainList, reg.Name + "." + reg.Spec.RegistryService.Ingress.DomainName + ":" + strconv.Itoa(port))
	}

	config := DockerConfig{
		Auths: map[string]AuthValue{},
	}
	for _, domain := range domainList {
		config.Auths[domain] = AuthValue{base64.StdEncoding.EncodeToString([]byte(reg.Spec.LoginId + ":" + reg.Spec.LoginPassword))}
	}

	configBytes , _ := json.Marshal(config)
	data[DockerConfigJson] = configBytes

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: regv1.K8sPrefix + regv1.K8sRegistryPrefix + reg.Name,
			Namespace: reg.Namespace,
			Labels: map[string]string {
				"secret": "docker",
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: data,
	}
}

func regBodyCheckForDCJSecret(reg *regv1.Registry) bool {
	regService := reg.Spec.RegistryService
	if (regService.ClusterIP == "") {
		return false
	}
	if (regService.ServiceType == regv1.RegServiceTypeIngress && regService.Ingress.DomainName == "") {
		return false
	} else if (regService.ServiceType == regv1.RegServiceTypeLoadBalancer && regService.LoadBalancer.IP == "") {
		return false
	}
	return true
}