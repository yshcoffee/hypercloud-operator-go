package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	"hypercloud-operator-go/pkg/controller/regctl"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	RegistryEventPath = "/registry/event"
)

var logz = log.Log.WithName("server-handler")

func CreateImageHandler(w http.ResponseWriter, r *http.Request) {
	regEvents := &regv1.RegistryEvents{}
	err := json.NewDecoder(r.Body).Decode(regEvents)
	if err != nil {
		logz.Error(err, "decode error")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	for _, event := range regEvents.Events {
		if event.Action == "push" {
			if len(event.Target.Tag) == 0 {
				logz.Info("tag is nil", "repository", event.Target.Repository)
				continue
			}
			logz.Info("pushed", "image", event.Target.Repository+":"+event.Target.Tag)
			createImage(event)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func createImage(event regv1.RegistryEvent) {
	// Get registry
	reg := registry(k8sClient, event)
	if reg == nil {
		logz.Info("registry not found", "registry_pod", strings.Split(event.Source.Addr, ":")[0])
		return
	}
	logger := logz.WithValues("registry", reg.Name, "ns", reg.Namespace)

	// Check if repository cr is exist
	repositoryCRNotFound := false
	repository := &regv1.Repository{}
	repositoryName := event.Target.Repository
	newImageTag := event.Target.Tag
	repositoryCRName := utils.ParseImageName(repositoryName) + "." + reg.Name
	err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: repositoryCRName, Namespace: reg.Namespace}, repository)
	if err != nil {
		if errors.IsNotFound(err) {
			repositoryCRNotFound = true
		} else {
			logger.Error(err, "Unknown error")
			return
		}
	}

	repoCtl := &regctl.RegistryRepository{}
	if repositoryCRNotFound {
		// If not exist, create repository cr
		logger.Info("create", "repository", repositoryName, "ver", newImageTag)
		repoCtl.Create(k8sClient, reg, repositoryName, []string{newImageTag}, scheme)

	} else {
		// Check if new version is exist
		if isExistVersion(repository.Spec.Versions, newImageTag) {
			logger.Info("version is already exist", "repository", repositoryName, "ver", newImageTag)
			return
		}

		// if exist, patch repository cr
		patchRepo := repository.DeepCopy()
		newVersion := regv1.ImageVersion{Version: newImageTag, CreatedAt: metav1.Now()}

		patchRepo.Spec.Versions = append(patchRepo.Spec.Versions, newVersion)
		logger.Info("repo_new_version", "repository", repositoryName, "ver", newImageTag)

		err := repoCtl.Patch(k8sClient, repository, patchRepo)
		if err != nil {
			logger.Error(err, "repository patch error")
			return
		}
	}

}

func isExistVersion(versions []regv1.ImageVersion, version string) bool {
	for _, ver := range versions {
		if ver.Version == version {
			return true
		}
	}
	return false
}

func registry(c client.Client, event regv1.RegistryEvent) *regv1.Registry {
	searchPodName := strings.Split(event.Source.Addr, ":")[0]
	podList := &corev1.PodList{}

	err := c.List(context.TODO(), podList, &client.ListOptions{})
	if err != nil {
		logz.Error(err, "Failed to list pods.")
		return nil
	}

	for _, pod := range podList.Items {
		if pod.Name == searchPodName {
			// Get registry name
			regName := pod.Labels["apps"][len(regv1.K8sPrefix):]
			logz.Info("msg", "registry_name", regName)
			registry := &regv1.Registry{}

			err := c.Get(context.TODO(), types.NamespacedName{Name: regName, Namespace: pod.Namespace}, registry)
			if err != nil {
				logz.Error(err, "Get regsitry is failed")
				return nil
			}

			return registry
		}
	}

	logz.Info(searchPodName + " pod not found")
	return nil
}
