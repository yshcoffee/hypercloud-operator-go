package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	RegistryEventPath = "/registry/event"
)

var logz = log.Log.WithName("registry-handler")

func CreateImageHandler(w http.ResponseWriter, r *http.Request) {
	regEvents := new(regv1.RegistryEvents)
	// err := json.NewDecoder(r.Body).Decode(regEvent)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logz.Error(err, "Unknown error")
		return
	}
	logz.Info("regEvent", "contents", string(body))
	json.Unmarshal(body, regEvents)

	for _, event := range regEvents.Events {
		logz.Info("regEvent", "action", event.Action, "image", event.Target.Rrepository+":"+event.Target.Tag)
	}

}
