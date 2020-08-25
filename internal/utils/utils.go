package utils

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetLabel(reg *regv1.Registry) map[string]string {
	return map[string]string{}
}

// Use for GetRegistryLogger
func getFuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(4, pc) //Skip: 3 (Callers, getFuncName, GetRegistryLogger, get)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}

// [TODO] API is not worked well
func GetRegistryLogger(subresource interface{}, resNamespace, resName string) logr.Logger {
	typeName := reflect.TypeOf(subresource).Name()
	funcName := getFuncName()
	path := strings.Split(funcName, ".")
	funcName = path[len(path)-1]

	return log.Log.WithValues(typeName+".Namespace", resNamespace, typeName+".Name", resName, typeName+".Api", funcName)
}

func SetError(error error, patchReg *regv1.Registry, condition status.Condition) {
	if error != nil {
		condition.Message = error.Error()
	}
	patchReg.Status.Conditions.SetCondition(condition)
}