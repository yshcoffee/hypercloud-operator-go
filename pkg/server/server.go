package server

import (
	"hypercloud-operator-go/pkg/server/handler"
	"net/http"

	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	port = ":28677"
)

var logger = log.Log.WithName("registry-server")

func StartServer() {
	r := mux.NewRouter()

	logger.Info("Handle", "Path", handler.RegistryEventPath)
	r.HandleFunc(handler.RegistryEventPath, handler.CreateImageHandler).Methods(http.MethodPost)

	logger.Info("Listen", "Port", port)
	if err := http.ListenAndServe(port, r); err != nil {
		logger.Error(err, "Server listen error")
	}
}
