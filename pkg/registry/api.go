package registry

import (
	"encoding/json"
	"fmt"
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
	req, err := http.NewRequest(http.MethodGet, r.URL+"/v2/_catalog", nil)
	if err != nil {
		logger.Error(err, "")
		return nil
	}
	req.SetBasicAuth(r.Login.Username, r.Login.Password)
	res, err := r.Client.Do(req)
	if err != nil {
		logger.Error(err, "")
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Error(err, "")
		return nil
	}
	logger.Info("contents", "repositories", string(body))

	rawRepos := &Repositories{}
	repos := &Repositories{}

	json.Unmarshal(body, rawRepos)

	for _, repo := range rawRepos.Repositories {
		tags := r.Tags(repo).Tags
		if tags != nil && len(tags) > 0 {
			repos.Repositories = append(repos.Repositories, repo)
		}
	}

	return repos
}

func (r *RegistryApi) Tags(imageName string) *Repository {
	repo := &Repository{}
	req, err := http.NewRequest(http.MethodGet, r.URL+"/v2/"+imageName+"/tags/list", nil)
	if err != nil {
		logger.Error(err, "")
		return nil
	}
	req.SetBasicAuth(r.Login.Username, r.Login.Password)
	res, err := r.Client.Do(req)
	if err != nil {
		logger.Error(err, "")
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Error(err, "")
		return nil
	}
	logger.Info("contents", "tags", string(body))
	json.Unmarshal(body, repo)

	return repo
}

func (r *RegistryApi) DockerContentDigest(imageName, tag string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, r.URL+"/v2/"+imageName+"/manifests/"+tag, nil)
	if err != nil {
		logger.Error(err, "")
		return "", err
	}

	req.SetBasicAuth(r.Login.Username, r.Login.Password)
	res, err := r.Client.Do(req)
	if err != nil {
		logger.Error(err, "")
		return "", err
	}

	for key, val := range res.Header {
		if key == "Docker-Content-Digest" {
			return val[0], nil
		}
	}

	if res.StatusCode >= 400 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Error(err, "")
			return "", err
		}
		logger.Error(nil, "err", "err", fmt.Sprintf("%s", string(body)))
		return "", fmt.Errorf("error!! %s", string(body))
	}

	return "", nil
}

func (r *RegistryApi) DeleteManifest(imageName, digest string) error {
	req, err := http.NewRequest(http.MethodDelete, r.URL+"/v2/"+imageName+"/manifests/"+digest, nil)
	if err != nil {
		logger.Error(err, "")
		return err
	}

	req.SetBasicAuth(r.Login.Username, r.Login.Password)
	res, err := r.Client.Do(req)
	if err != nil {
		logger.Error(err, "")
		return err
	}

	if res.StatusCode >= 400 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Error(err, "")
			return nil
		}
		logger.Error(nil, "err", "err", fmt.Sprintf("%s", string(body)))
		return fmt.Errorf("error!! %s", string(body))
	}

	return nil
}
