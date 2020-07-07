package v1

type Ingress struct {
	Port       int    `json:"port",default:"443"`
	DomainName string `json:"domainName",default:""`
}

type LoadBalancer struct {
	Port int `json:"port",default:"443"`
}
