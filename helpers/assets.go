package helpers

import (
	"os"
	"path"
	"path/filepath"
)

type Assets struct {
	ServiceBroker      string
	SecurityRules      string
	EmptySecurityRules string
}

func NewAssets() Assets {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	return Assets{
		ServiceBroker:      filepath.ToSlash(path.Join(pwd, "../../assets/service_broker")),
		SecurityRules:      filepath.ToSlash(path.Join(pwd, "../../assets/security_groups/security-rules.json")),
		EmptySecurityRules: filepath.ToSlash(path.Join(pwd, "../../assets/security_groups/empty-security-rules.json")),
	}
}
