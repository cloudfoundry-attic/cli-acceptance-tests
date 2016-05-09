package application

import (
	"strconv"

	"github.com/cloudfoundry/cli/cf/api/appinstances"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type RestartAppInstance struct {
	ui               terminal.UI
	config           coreconfig.Reader
	appReq           requirements.ApplicationRequirement
	appInstancesRepo appinstances.AppInstancesRepository
}

func init() {
	commandregistry.Register(&RestartAppInstance{})
}

func (cmd *RestartAppInstance) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "restart-app-instance",
		Description: T("Terminate the running application Instance at the given index and instantiate a new instance of the application with the same index"),
		Usage: []string{
			T("CF_NAME restart-app-instance APP_NAME INDEX"),
		},
	}
}

func (cmd *RestartAppInstance) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		usage := commandregistry.Commands.CommandUsage("restart-app-instance")
		cmd.ui.Failed(T("Incorrect Usage. Requires arguments\n\n") + usage)
	}

	appName := fc.Args()[0]

	cmd.appReq = requirementsFactory.NewApplicationRequirement(appName)

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs
}

func (cmd *RestartAppInstance) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appInstancesRepo = deps.RepoLocator.GetAppInstancesRepository()
	return cmd
}

func (cmd *RestartAppInstance) Execute(fc flags.FlagContext) {
	app := cmd.appReq.GetApplication()

	instance, err := strconv.Atoi(fc.Args()[1])

	if err != nil {
		cmd.ui.Failed(T("Instance must be a non-negative integer"))
	}

	cmd.ui.Say(T("Restarting instance {{.Instance}} of application {{.AppName}} as {{.Username}}",
		map[string]interface{}{
			"Instance": instance,
			"AppName":  terminal.EntityNameColor(app.Name),
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.appInstancesRepo.DeleteInstance(app.GUID, instance)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
}
