package schemes

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

func Pod(reg *regv1.Registry) *corev1.Pod {
	var resName string
	label := map[string]string{}
	label["app"] = "registry"
	label["apps"] = regv1.K8sPrefix + reg.Name

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resName,
			Namespace: reg.Namespace,
			Labels:    label,
		},
		Spec: corev1.PodSpec{},
	}
}
