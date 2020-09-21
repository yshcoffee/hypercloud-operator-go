package regctl

import (
	"context"
	"hypercloud-operator-go/internal/schemes"
	"hypercloud-operator-go/internal/utils"
	"strings"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"github.com/operator-framework/operator-sdk/pkg/status"
	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	MountPathDiffKey = "MountPath"
	PvcNameDiffKey   = "PvcName"
	ImageDiffKey     = "Image"
)

type RegistryDeployment struct {
	deploy *appsv1.Deployment
	logger *utils.RegistryLogger
}

func (r *RegistryDeployment) Handle(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {
	if err := r.get(c, reg); err != nil {
		if errors.IsNotFound(err) {
			if err := r.create(c, reg, patchReg, scheme); err != nil {
				r.logger.Error(err, "create Deployment error")
				return err
			}
		} else {
			r.logger.Error(err, "Deployment error")
			return err
		}
	}

	r.logger.Info("Check if patch exists.")
	diff := r.compare(reg)
	if diff == nil {
		r.logger.Error(nil, "Invalid deployment!!!")
		r.delete(c, patchReg)
	} else if len(diff) > 0 {
		r.patch(c, reg, patchReg, diff)
	}

	return nil
}

func (r *RegistryDeployment) Ready(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, useGet bool) error {
	var err error = nil
	condition := &status.Condition{
		Status: corev1.ConditionFalse,
		Type:   regv1.ConditionTypeDeployment,
	}
	defer utils.SetError(err, patchReg, condition)
	if useGet {
		err = r.get(c, reg)
		if err != nil {
			r.logger.Error(err, "Deployment error")
			return err
		}
	}

	if r.deploy == nil {
		r.logger.Info("NotReady")

		err = regv1.MakeRegistryError("NotReady")
		return err
	}

	r.logger.Info("Ready")
	condition.Status = corev1.ConditionTrue
	return nil
}

func (r *RegistryDeployment) create(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, scheme *runtime.Scheme) error {
	if err := controllerutil.SetControllerReference(reg, r.deploy, scheme); err != nil {
		r.logger.Error(err, "SetOwnerReference Failed")
		condition := status.Condition{
			Status:  corev1.ConditionFalse,
			Type:    regv1.ConditionTypeDeployment,
			Message: err.Error(),
		}

		patchReg.Status.Conditions.SetCondition(condition)
		return nil
	}

	r.logger.Info("Create registry deployment")
	err := c.Create(context.TODO(), r.deploy)
	if err != nil {
		condition := status.Condition{
			Status:  corev1.ConditionFalse,
			Type:    regv1.ConditionTypeDeployment,
			Message: err.Error(),
		}

		patchReg.Status.Conditions.SetCondition(condition)
		r.logger.Error(err, "Creating registry deployment is failed.")
		return nil
	}

	return nil
}

func (r *RegistryDeployment) get(c client.Client, reg *regv1.Registry) error {
	r.deploy = schemes.Deployment(reg)
	r.logger = utils.NewRegistryLogger(*r, r.deploy.Namespace, r.deploy.Name)

	req := types.NamespacedName{Name: r.deploy.Name, Namespace: r.deploy.Namespace}

	err := c.Get(context.TODO(), req, r.deploy)
	if err != nil {
		r.logger.Error(err, "Get regsitry deployment is failed")
		return err
	}

	return nil
}

func (r *RegistryDeployment) patch(c client.Client, reg *regv1.Registry, patchReg *regv1.Registry, diff []utils.Diff) error {
	target := r.deploy.DeepCopy()
	originObject := client.MergeFrom(r.deploy)

	var deployContainer *corev1.Container = nil
	// var contPvcVm *corev1.VolumeMount = nil
	volumeMap := map[string]corev1.Volume{}
	podSpec := target.Spec.Template.Spec

	r.logger.Info("Get", "Patch Keys", strings.Join(utils.DiffKeyList(diff), ", "))

	// Get registry container
	for i, cont := range podSpec.Containers {
		if cont.Name == "registry" {
			deployContainer = &podSpec.Containers[i]
			break
		}
	}

	if deployContainer == nil {
		r.logger.Error(regv1.MakeRegistryError(regv1.ContainerNotFound), "registry container is nil")
		return nil
	}

	for _, d := range diff {
		switch d.Key {
		case ImageDiffKey:
			deployContainer.Image = reg.Spec.Image

		case MountPathDiffKey:
			found := false
			for i, vm := range deployContainer.VolumeMounts {
				if vm.Name == "registry" {
					deployContainer.VolumeMounts[i].MountPath = reg.Spec.PersistentVolumeClaim.MountPath
					found = true
					break
				}
			}

			if !found {
				r.logger.Error(regv1.MakeRegistryError(regv1.PvcVolumeMountNotFound), "registry pvc volume mount is nil")
				return nil
			}

		case PvcNameDiffKey:
			// Get volumes
			for _, vol := range podSpec.Volumes {
				volumeMap[vol.Name] = vol
			}

			vol, _ := volumeMap["registry"]
			if reg.Spec.PersistentVolumeClaim.Create != nil {
				vol.PersistentVolumeClaim.ClaimName = regv1.K8sPrefix + reg.Name
			} else {
				vol.PersistentVolumeClaim.ClaimName = reg.Spec.PersistentVolumeClaim.Exist.PvcName
			}
		}
	}

	// Patch
	if err := c.Patch(context.TODO(), target, originObject); err != nil {
		r.logger.Error(err, "Unknown error patch")
		return err
	}
	return nil
}

func (r *RegistryDeployment) delete(c client.Client, patchReg *regv1.Registry) error {
	if err := c.Delete(context.TODO(), r.deploy); err != nil {
		r.logger.Error(err, "Unknown error delete deployment")
		return err
	}

	condition := status.Condition{
		Type:   regv1.ConditionTypeDeployment,
		Status: corev1.ConditionFalse,
	}

	patchReg.Status.Conditions.SetCondition(condition)

	return nil
}

func (r *RegistryDeployment) compare(reg *regv1.Registry) []utils.Diff {
	diff := []utils.Diff{}
	var deployContainer *corev1.Container = nil
	podSpec := r.deploy.Spec.Template.Spec
	volumeMap := map[string]corev1.Volume{}

	// Get registry container
	for _, cont := range podSpec.Containers {
		if cont.Name == "registry" {
			deployContainer = &cont
		}
	}

	if deployContainer == nil {
		r.logger.Error(regv1.MakeRegistryError(regv1.ContainerNotFound), "registry container is nil")
		return nil
	}

	// Get volumes
	for _, vol := range podSpec.Volumes {
		volumeMap[vol.Name] = vol
	}

	if reg.Spec.Image != deployContainer.Image {
		diff = append(diff, utils.Diff{Type: utils.Replace, Key: ImageDiffKey})
	}

	if reg.Spec.PersistentVolumeClaim.Create != nil {
		vol, exist := volumeMap["registry"]
		if !exist {
			r.logger.Info("Registry volume is not exist.")
		} else if vol.VolumeSource.PersistentVolumeClaim.ClaimName != (regv1.K8sPrefix + reg.Name) {
			diff = append(diff, utils.Diff{Type: utils.Replace, Key: PvcNameDiffKey})
		}
	} else {
		vol, exist := volumeMap["registry"]
		if !exist {
			r.logger.Info("Registry volume is not exist.")
		} else if vol.VolumeSource.PersistentVolumeClaim.ClaimName != reg.Spec.PersistentVolumeClaim.Exist.PvcName {
			diff = append(diff, utils.Diff{Type: utils.Replace, Key: PvcNameDiffKey})
		}
	}

	var contPvcVm *corev1.VolumeMount = nil
	for i, vm := range deployContainer.VolumeMounts {
		if vm.Name == "registry" {
			contPvcVm = &deployContainer.VolumeMounts[i]
			break
		}
	}

	if contPvcVm == nil {
		r.logger.Error(regv1.MakeRegistryError(regv1.PvcVolumeMountNotFound), "registry pvc volume mount is nil")
		return nil
	}

	if reg.Spec.PersistentVolumeClaim.MountPath != contPvcVm.MountPath {
		diff = append(diff, utils.Diff{Type: utils.Replace, Key: MountPathDiffKey})
	}

	return diff
}
