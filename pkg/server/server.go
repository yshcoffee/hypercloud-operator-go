package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	port = ":28677"
)

var logger = log.Log.WithName("registry-server")
var scheme *runtime.Scheme
var k8sClient client.Client

func StartServer(m manager.Manager) {
	r := mux.NewRouter()
	scheme = m.GetScheme()
	k8sClient = m.GetClient()
	logger.Info("Handle", "Path", RegistryEventPath)
	r.HandleFunc(RegistryEventPath, CreateImageHandler).Methods(http.MethodPost)

	logger.Info("Listen", "Port", port)
	if err := http.ListenAndServe(port, r); err != nil {
		logger.Error(err, "Server listen error")
	}
}
