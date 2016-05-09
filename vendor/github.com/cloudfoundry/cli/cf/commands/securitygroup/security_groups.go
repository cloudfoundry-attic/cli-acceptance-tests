package securitygroup

import (
	"fmt"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/securitygroups"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type SecurityGroups struct {
	ui                terminal.UI
	securityGroupRepo security_groups.SecurityGroupRepo
	configRepo        coreconfig.Reader
}

func init() {
	commandregistry.Register(&SecurityGroups{})
}

func (cmd *SecurityGroups) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "security-groups",
		Description: T("List all security groups"),
		Usage: []string{
			"CF_NAME security-groups",
		},
	}
}

func (cmd *SecurityGroups) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
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

func (cmd *SecurityGroups) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	return cmd
}

func (cmd *SecurityGroups) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Getting security groups as {{.username}}",
		map[string]interface{}{
			"username": terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	securityGroups, err := cmd.securityGroupRepo.FindAll()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(securityGroups) == 0 {
		cmd.ui.Say(T("No security groups"))
		return
	}

	table := cmd.ui.Table([]string{"", T("Name"), T("Organization"), T("Space")})

	for index, securityGroup := range securityGroups {
		if len(securityGroup.Spaces) > 0 {
			cmd.printSpaces(table, securityGroup, index)
		} else {
			table.Add(fmt.Sprintf("#%d", index), securityGroup.Name, "", "")
		}
	}
	table.Print()
}

type table interface {
	Add(row ...string)
	Print()
}

func (cmd SecurityGroups) printSpaces(table table, securityGroup models.SecurityGroup, index int) {
	outputted_index := false

	for _, space := range securityGroup.Spaces {
		if !outputted_index {
			table.Add(fmt.Sprintf("#%d", index), securityGroup.Name, space.Organization.Name, space.Name)
			outputted_index = true
		} else {
			table.Add("", securityGroup.Name, space.Organization.Name, space.Name)
		}
	}
}
