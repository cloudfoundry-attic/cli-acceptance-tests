package helpers

import (
	"os"
	"path"
)

type Assets struct {
	ServiceBroker string
}

func NewAssets() Assets {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	return Assets{
		ServiceBroker: path.Join(pwd, "../../assets/service_broker"),
	}
}
