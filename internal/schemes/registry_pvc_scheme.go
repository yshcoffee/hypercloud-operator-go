package schemes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

func PersistentVolumeClaim(reg *regv1.Registry) *corev1.PersistentVolumeClaim {
	resName := regv1.K8sPrefix + reg.Name
	label := map[string]string{}
	label["app"] = "registry"
	label["apps"] = regv1.K8sPrefix + reg.Name

	var accessModes []corev1.PersistentVolumeAccessMode
	for mode := range reg.Spec.PersistentVolumeClaim.Create.AccessModes {
		accessModes = append(accessModes, corev1.PersistentVolumeAccessMode(mode))
	}

	var v corev1.PersistentVolumeMode
	v = corev1.PersistentVolumeMode(reg.Spec.PersistentVolumeClaim.Create.VolumeMode)

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resName,
			Namespace: reg.Namespace,
			Labels:    label,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(reg.Spec.PersistentVolumeClaim.Create.StorageSize),
				},
			},
			StorageClassName: &reg.Spec.PersistentVolumeClaim.Create.StorageClassName,
			VolumeMode:       &v,
		},
	}
}
