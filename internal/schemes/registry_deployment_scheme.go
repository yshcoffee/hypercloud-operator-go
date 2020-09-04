package schemes

import (
	"encoding/base64"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ConfigMapMountPath   = "/etc/docker/registry"
	SecretCertMountPath  = "/certs"
	RegistryPVCMountPath = "/var/lib/registry"
)

func Deployment(reg *regv1.Registry) *appsv1.Deployment {
	var resName, pvcMountPath, pvcName, configMapName string
	resName = regv1.K8sPrefix + reg.Name
	label := map[string]string{}
	label["app"] = "registry"
	label["apps"] = resName
	label[resName] = "lb"

	if len(reg.Spec.PersistentVolumeClaim.MountPath) == 0 {
		pvcMountPath = RegistryPVCMountPath
	} else {
		pvcMountPath = reg.Spec.PersistentVolumeClaim.MountPath
	}

	if reg.Spec.PersistentVolumeClaim.Exist != nil {
		pvcName = reg.Spec.PersistentVolumeClaim.Exist.PvcName
	} else {
		pvcName = regv1.K8sPrefix + reg.Name
	}

	idPasswd := reg.Spec.LoginId + ":" + reg.Spec.LoginPassword
	loginAuth := base64.StdEncoding.EncodeToString([]byte(idPasswd))

	if len(reg.Spec.CustomConfigYml) != 0 {
		configMapName = reg.Spec.CustomConfigYml
	} else {
		configMapName = regv1.K8sPrefix + reg.Name
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resName,
			Namespace: reg.Namespace,
			Labels:    label,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: label,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: label,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: reg.Spec.Image,
							Name:  "registry",
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "-c", "mkdir /auth; htpasswd -Bbn $ID $PASSWD > /auth/htpasswd"},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("0.2"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          RegistryPortName,
									ContainerPort: RegistryTargetPort,
									Protocol:      RegistryPortProtocol,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "REGISTRY_AUTH",
									Value: "htpasswd",
								},
								{
									Name:  "REGISTRY_AUTH_HTPASSWD_REALM",
									Value: "Registry Realm",
								},
								{
									Name:  "REGISTRY_AUTH_HTPASSWD_PATH",
									Value: "/auth/htpasswd",
								},
								{
									Name:  "REGISTRY_HTTP_ADDR",
									Value: string("0.0.0.0:") + strconv.Itoa(RegistryTargetPort),
								},
								{
									Name:  "REGISTRY_HTTP_TLS_CERTIFICATE",
									Value: "/certs/localhub.crt",
								},
								{
									Name:  "REGISTRY_HTTP_TLS_KEY",
									Value: "/certs/localhub.key",
								},
								// from secret
								{
									Name: "ID",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: regv1.K8sPrefix + reg.Name,
											},
											Key: "ID",
										},
									},
								},
								{
									Name: "PASSWD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: regv1.K8sPrefix + reg.Name,
											},
											Key: "PASSWD",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: ConfigMapMountPath,
								},
								{
									Name:      "certs",
									MountPath: SecretCertMountPath,
								},
							},
							ReadinessProbe: &corev1.Probe{
								PeriodSeconds:       3,
								SuccessThreshold:    1,
								TimeoutSeconds:      1,
								InitialDelaySeconds: 5,
								FailureThreshold:    10,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/v2/_catalog",
										Port: intstr.IntOrString{IntVal: RegistryTargetPort},
										HTTPHeaders: []corev1.HTTPHeader{
											corev1.HTTPHeader{
												Name:  "authorization",
												Value: "Basic " + loginAuth,
											},
										},
										Scheme: corev1.URISchemeHTTPS,
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								TimeoutSeconds:      30,
								InitialDelaySeconds: 5,
								FailureThreshold:    10,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/v2/_catalog",
										Port: intstr.IntOrString{IntVal: RegistryTargetPort},
										HTTPHeaders: []corev1.HTTPHeader{
											corev1.HTTPHeader{
												Name:  "authorization",
												Value: "Basic " + loginAuth,
											},
										},
										Scheme: corev1.URISchemeHTTPS,
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: configMapName},
								},
							},
						},
						corev1.Volume{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: regv1.K8sPrefix + reg.Name,
								},
							},
						},
						corev1.Volume{
							Name: "registry",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
			},
		},
	}

	if reg.Spec.PersistentVolumeClaim.Create != nil {
		if reg.Spec.PersistentVolumeClaim.Create.VolumeMode == "Block" {
			vd := corev1.VolumeDevice{
				Name:       "registry",
				DevicePath: pvcMountPath,
			}

			deployment.Spec.Template.Spec.Containers[0].VolumeDevices = append(deployment.Spec.Template.Spec.Containers[0].VolumeDevices, vd)
		} else {
			vm := corev1.VolumeMount{
				Name:      "registry",
				MountPath: pvcMountPath,
			}

			deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, vm)
		}
	}

	return deployment
}
