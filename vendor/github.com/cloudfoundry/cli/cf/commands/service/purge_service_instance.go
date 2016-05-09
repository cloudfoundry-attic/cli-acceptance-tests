package service

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type PurgeServiceInstance struct {
	ui          terminal.UI
	serviceRepo api.ServiceRepository
}

func init() {
	commandregistry.Register(&PurgeServiceInstance{})
}

func (cmd *PurgeServiceInstance) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "purge-service-instance",
		Description: T("Recursively remove a service instance and child objects from Cloud Foundry database without making requests to a service broker"),
		Usage: []string{
			T("CF_NAME purge-service-instance SERVICE_INSTANCE"),
			"\n\n",
			cmd.scaryWarningMessage(),
		},
		Flags: fs,
	}
}

func (cmd *PurgeServiceInstance) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("purge-service-instance"))
	}

	minRequiredAPIVersion, err := semver.Make("2.36.0")
	if err != nil {
		panic(err.Error())
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewMinAPIVersionRequirement("purge-service-instance", minRequiredAPIVersion),
	}

	return reqs
}

func (cmd *PurgeServiceInstance) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	return cmd
}

func (cmd *PurgeServiceInstance) scaryWarningMessage() string {
	return T(`WARNING: This operation assumes that the service broker responsible for this service instance is no longer available or is not responding with a 200 or 410, and the service instance has been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service instance will be removed from Cloud Foundry, including service bindings and service keys.`)
}

func (cmd *PurgeServiceInstance) Execute(c flags.FlagContext) {
	instanceName := c.Args()[0]

	instance, err := cmd.serviceRepo.FindInstanceByName(instanceName)
	if err != nil {
		if _, ok := err.(*errors.ModelNotFoundError); ok {
			cmd.ui.Warn(T("Service instance {{.InstanceName}} not found", map[string]interface{}{"InstanceName": instanceName}))
			return
		}

		cmd.ui.Failed(err.Error())
	}

	force := c.Bool("f")
	if !force {
		cmd.ui.Warn(cmd.scaryWarningMessage())
		confirmed := cmd.ui.Confirm(T("Really purge service instance {{.InstanceName}} from Cloud Foundry?",
			map[string]interface{}{"InstanceName": instanceName},
		))

		if !confirmed {
			return
		}
	}

	cmd.ui.Say(T("Purging service {{.InstanceName}}...", map[string]interface{}{"InstanceName": instanceName}))
	err = cmd.serviceRepo.PurgeServiceInstance(instance)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
