package regctl

import (
	"context"
	"hypercloud-operator-go/internal/schemes"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RegistryRepository struct {
	repo *regv1.Repository
}

var logger = log.Log.WithName("registry_repository")

func (r *RegistryRepository) Create(c client.Client, reg *regv1.Registry, imageName string, tags []string, scheme *runtime.Scheme) error {
	r.repo = schemes.Repository(reg, imageName, tags)
	if err := controllerutil.SetControllerReference(reg, r.repo, scheme); err != nil {
		logger.Error(err, "Controller reference failed")
		return err
	}

	if err := c.Create(context.TODO(), r.repo); err != nil {
		logger.Error(err, "Create failed")
		return err
	}

	logger.Info("Created", "Registry", reg.Name, "Repository", r.repo.Name, "Namespace", reg.Namespace)
	return nil
}

func (r *RegistryRepository) Patch(c client.Client, repo *regv1.Repository, patchRepo *regv1.Repository) error {
	originObject := client.MergeFrom(repo)

	logger.Info("Patch", "Repository", patchRepo.Name+"/"+patchRepo.Namespace)

	// Patch
	if err := c.Patch(context.TODO(), patchRepo, originObject); err != nil {
		logger.Error(err, "Unknown error patching status")
		return err
	}
	return nil
}

func (r *RegistryRepository) Delete(c client.Client, reg *regv1.Registry, imageName string, scheme *runtime.Scheme) error {
	r.repo = schemes.Repository(reg, imageName, nil)
	if err := c.Delete(context.TODO(), r.repo); err != nil {
		logger.Error(err, "Delete failed")
		return err
	}

	logger.Info("Deleted", "Registry", reg.Name, "Repository", r.repo.Name, "Namespace", reg.Namespace)
	return nil
}
