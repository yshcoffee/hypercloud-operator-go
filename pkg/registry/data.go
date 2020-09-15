package registry

import (
	"context"
	"hypercloud-operator-go/internal/schemes"
	"hypercloud-operator-go/internal/utils"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func CAData() []byte {
	c, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		logger.Error(err, "Unknown error")
		return nil
	}

	secret := &corev1.Secret{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: schemes.RootCASecretName, Namespace: schemes.RootCASecretNamespace}, secret)
	if err != nil {
		logger.Error(err, "Unknown error")
		return nil
	}

	data := secret.Data
	cacrt, exist := data[schemes.RootCACert]
	if !exist {
		logger.Info("CA is not found")
		return nil
	}

	return cacrt
}

func RegistryUrl(reg *regv1.Registry) string {
	c, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		logger.Error(err, "Unknown error")
		return ""
	}

	secret := &corev1.Secret{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: regv1.K8sPrefix + reg.Name, Namespace: reg.Namespace}, secret)
	if err != nil {
		logger.Error(err, "Unknown error")
		return ""
	}

	data := secret.Data
	url, exist := data["REGISTRY_URL"]
	if !exist {
		logger.Info("CA is not found")
		return ""
	}

	return utils.SCHEME_HTTPS_PREFIX + string(url)
}
