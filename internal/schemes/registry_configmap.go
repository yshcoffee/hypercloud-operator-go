package schemes

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

func ConfigMap(reg *regv1.Registry) *corev1.ConfigMap {
	var resName string
	label := map[string]string{}
	label["app"] = "registry"
	label["apps"] = regv1.K8sPrefix + reg.Name

	if len(reg.Spec.CustomConfigYml) != 0 {
		resName = reg.Spec.CustomConfigYml
	} else {
		resName = regv1.K8sPrefix + reg.Name
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resName,
			Namespace: reg.Namespace,
			Labels:    label,
		},
	}
}
