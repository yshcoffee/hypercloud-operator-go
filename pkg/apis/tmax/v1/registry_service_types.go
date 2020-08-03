package v1

// use ingress service type
type Ingress struct {
	// external port. Generally use 443 port
	Port int `json:"port",default:"443"`

	// (example: 192.168.6.110.nip.io)
	DomainName string `json:"domainName",default:""`
}

// use loadBalancer service type
type LoadBalancer struct {
	// external port. Generally use 443 port
	Port int `json:"port",default:"443"`
}
