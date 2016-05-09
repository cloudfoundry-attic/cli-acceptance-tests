package requirements

import (
	"fmt"

	"errors"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type APIEndpointRequirement struct {
	config coreconfig.Reader
}

func NewAPIEndpointRequirement(config coreconfig.Reader) APIEndpointRequirement {
	return APIEndpointRequirement{config}
}

func (req APIEndpointRequirement) Execute() error {
	if req.config.APIEndpoint() == "" {
		loginTip := terminal.CommandColor(fmt.Sprintf(T("{{.CFName}} login", map[string]interface{}{"CFName": cf.Name})))
		apiTip := terminal.CommandColor(fmt.Sprintf(T("{{.CFName}} api", map[string]interface{}{"CFName": cf.Name})))
		return errors.New(T("No API endpoint set. Use '{{.LoginTip}}' or '{{.APITip}}' to target an endpoint.",
			map[string]interface{}{
				"LoginTip": loginTip,
				"APITip":   apiTip,
			}))
	}

	return nil
}
