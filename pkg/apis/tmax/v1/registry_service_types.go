package v1

// use ingress service type
type Ingress struct {
	// [TODO] Minimum, Maximum
	// (example: 192.168.6.110.nip.io)
	DomainName string `json:"domainName"`
}

// use loadBalancer service type
type LoadBalancer struct {
	// [TODO] Minimum, Maximum
	IP string `json:"ip,omitempty"`
}
