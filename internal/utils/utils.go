package utils

import (
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetServiceName(reg *regv1.Registry) string {
	return reg.Name + "-svc"
}

func GetLabel(reg *regv1.Registry) map[string]string {
	return map[string]string{}
}

// Use for GetRegistryLogger
func getFuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc) //Skip: 3 (Callers, getFuncName, GetRegistryLogger)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}

func GetRegistryLogger(subresource interface{}, resNamespace, resName string) logr.Logger {
	typeName := reflect.TypeOf(subresource).Name()
	funcName := getFuncName()
	path := strings.Split(funcName, ".")
	funcName = path[len(path)-1]

	return log.Log.WithValues(typeName+".Namespace", resNamespace, typeName+".Name", resName, typeName+".Api", funcName)
}
