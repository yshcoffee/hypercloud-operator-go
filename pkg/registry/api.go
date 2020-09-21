package registry

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Repositories struct {
	Repositories []string `json:"repositories"`
}

type Repository struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type RegistryApi struct {
	*httpClient
}

var logger = log.Log.WithName("registry-api")

func NewRegistryApi(reg *regv1.Registry) *RegistryApi {
	ra := &RegistryApi{}
	regURL := RegistryUrl(reg)

	ra.httpClient = NewHTTPClient(regURL, reg.Spec.LoginId, reg.Spec.LoginPassword)
	return ra
}

func (r *RegistryApi) Catalog() *Repositories {
	repos := &Repositories{}
	req, err := http.NewRequest(http.MethodGet, r.URL+"/v2/_catalog", nil)
	if err != nil {
		logger.Error(err, "Unknown error")
		return nil
	}
	req.SetBasicAuth(r.Login.Username, r.Login.Password)
	res, err := r.Client.Do(req)
	if err != nil {
		logger.Error(err, "Unknown error")
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Error(err, "Unknown error")
		return nil
	}
	logger.Info("contents", "repositories", string(body))
	json.Unmarshal(body, repos)

	return repos
}

func (r *RegistryApi) Tags(imageName string) *Repository {
	repo := &Repository{}
	req, err := http.NewRequest(http.MethodGet, r.URL+"/v2/"+imageName+"/tags/list", nil)
	if err != nil {
		logger.Error(err, "Unknown error")
		return nil
	}
	req.SetBasicAuth(r.Login.Username, r.Login.Password)
	res, err := r.Client.Do(req)
	if err != nil {
		logger.Error(err, "Unknown error")
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Error(err, "Unknown error")
		return nil
	}
	logger.Info("contents", "tags", string(body))
	json.Unmarshal(body, repo)

	return repo
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
