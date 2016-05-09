package environmentvariablegroup

import (
	"github.com/cloudfoundry/cli/cf/api/environmentvariablegroups"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type RunningEnvironmentVariableGroup struct {
	ui                           terminal.UI
	config                       coreconfig.ReadWriter
	environmentVariableGroupRepo environmentvariablegroups.EnvironmentVariableGroupsRepository
}

func init() {
	commandregistry.Register(&RunningEnvironmentVariableGroup{})
}

func (cmd *RunningEnvironmentVariableGroup) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "running-environment-variable-group",
		Description: T("Retrieve the contents of the running environment variable group"),
		ShortName:   "revg",
		Usage: []string{
			T("CF_NAME running-environment-variable-group"),
		},
	}
}

func (cmd *RunningEnvironmentVariableGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs
}

func (cmd *RunningEnvironmentVariableGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.environmentVariableGroupRepo = deps.RepoLocator.GetEnvironmentVariableGroupsRepository()
	return cmd
}

func (cmd *RunningEnvironmentVariableGroup) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Retrieving the contents of the running environment variable group as {{.Username}}...", map[string]interface{}{
		"Username": terminal.EntityNameColor(cmd.config.Username())}))

	runningEnvVars, err := cmd.environmentVariableGroupRepo.ListRunning()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()

	table := cmd.ui.Table([]string{T("Variable Name"), T("Assigned Value")})
	for _, envVar := range runningEnvVars {
		table.Add(envVar.Name, envVar.Value)
	}
	table.Print()
}
