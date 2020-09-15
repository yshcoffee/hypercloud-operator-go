package registry

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
)

type Authorizer struct {
	Username string
	Password string
}

type httpClient struct {
	Login Authorizer
	URL   string
	*http.Client
}

func NewHTTPClient(url, username, password string) *httpClient {
	caCert := CAData()
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	return &httpClient{
		URL:    url,
		Login:  Authorizer{Username: username, Password: password},
		Client: c,
	}
}
