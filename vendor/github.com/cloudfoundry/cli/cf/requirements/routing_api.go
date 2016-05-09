package requirements

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
)

type RoutingAPIRequirement struct {
	config coreconfig.Reader
}

func NewRoutingAPIRequirement(config coreconfig.Reader) RoutingAPIRequirement {
	return RoutingAPIRequirement{
		config,
	}
}

func (req RoutingAPIRequirement) Execute() error {
	if len(req.config.RoutingAPIEndpoint()) == 0 {
		return errors.New(T("This command requires the Routing API. Your targeted endpoint reports it is not enabled."))
	}

	return nil
}
