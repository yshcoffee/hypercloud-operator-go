package utils

import (
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
)

func GetServiceName(reg *regv1.Registry) string {
	return reg.Name + "-svc"
}

func GetLabel(reg *regv1.Registry) map[string]string {
	return map[string]string{}
}
