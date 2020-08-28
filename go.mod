module hypercloud-operator-go

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/goharbor/harbor/src v0.0.0-20200819104903-8d7d5790b872
	github.com/nsf/jsondiff v0.0.0-20200515183724-f29ed568f4ce // indirect
	github.com/operator-framework/operator-sdk v0.18.1
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
