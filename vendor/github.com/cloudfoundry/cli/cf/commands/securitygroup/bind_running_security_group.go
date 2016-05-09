package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/securitygroups"
	"github.com/cloudfoundry/cli/cf/api/securitygroups/defaults/running"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type bindToRunningGroup struct {
	ui                terminal.UI
	configRepo        coreconfig.Reader
	securityGroupRepo security_groups.SecurityGroupRepo
	runningGroupRepo  running.RunningSecurityGroupsRepo
}

func init() {
	commandregistry.Register(&bindToRunningGroup{})
}

func (cmd *bindToRunningGroup) MetaData() commandregistry.CommandMetadata {
	primaryUsage := T("CF_NAME bind-running-security-group SECURITY_GROUP")
	tipUsage := T("TIP: Changes will not apply to existing running applications until they are restarted.")
	return commandregistry.CommandMetadata{
		Name:        "bind-running-security-group",
		Description: T("Bind a security group to the list of security groups to be used for running applications"),
		Usage: []string{
			primaryUsage,
			"\n\n",
			tipUsage,
		},
	}
}

func (cmd *bindToRunningGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("bind-running-security-group"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs
}

func (cmd *bindToRunningGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	cmd.runningGroupRepo = deps.RepoLocator.GetRunningSecurityGroupsRepository()
	return cmd
}

func (cmd *bindToRunningGroup) Execute(context flags.FlagContext) {
	name := context.Args()[0]

	securityGroup, err := cmd.securityGroupRepo.Read(name)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say(T("Binding security group {{.security_group}} to defaults for running as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(securityGroup.Name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	err = cmd.runningGroupRepo.BindToRunningSet(securityGroup.GUID)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n\n")
	cmd.ui.Say(T("TIP: Changes will not apply to existing running applications until they are restarted."))
}
