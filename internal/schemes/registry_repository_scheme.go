package schemes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"hypercloud-operator-go/internal/utils"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

func Repository(reg *regv1.Registry, imageName string, tags []string) *regv1.Repository {
	label := map[string]string{}
	label["app"] = "registry"
	label["apps"] = regv1.K8sPrefix + reg.Name

	versions := []regv1.ImageVersion{}
	if tags != nil {
		for _, ver := range tags {
			newVersion := regv1.ImageVersion{CreatedAt: metav1.Now(), Version: ver}
			versions = append(versions, newVersion)
		}
	}

	return &regv1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.ParseImageName(imageName) + "." + reg.Name,
			Namespace: reg.Namespace,
			Labels:    label,
		},
		Spec: regv1.RepositorySpec{
			Name:     imageName,
			Registry: reg.Name,
			Versions: versions,
		},
	}
}
