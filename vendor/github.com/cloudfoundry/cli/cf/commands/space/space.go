package space

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/models"

	"github.com/cloudfoundry/cli/cf/api/spacequotas"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ShowSpace struct {
	ui          terminal.UI
	config      coreconfig.Reader
	spaceReq    requirements.SpaceRequirement
	quotaRepo   spacequotas.SpaceQuotaRepository
	pluginModel *plugin_models.GetSpace_Model
	pluginCall  bool
}

func init() {
	commandregistry.Register(&ShowSpace{})
}

func (cmd *ShowSpace) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &flags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given space's guid.  All other output for the space is suppressed.")}
	fs["security-group-rules"] = &flags.BoolFlag{Name: "security-group-rules", Usage: T("Retrieve the rules for all the security groups associated with the space")}
	return commandregistry.CommandMetadata{
		Name:        "space",
		Description: T("Show space info"),
		Usage: []string{
			T("CF_NAME space SPACE"),
		},
		Flags: fs,
	}
}

func (cmd *ShowSpace) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("space"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}

	return reqs
}

func (cmd *ShowSpace) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.Space
	return cmd
}

func (cmd *ShowSpace) Execute(c flags.FlagContext) {
	space := cmd.spaceReq.GetSpace()
	if cmd.pluginCall {
		cmd.populatePluginModel(space)
		return
	}
	if c.Bool("guid") {
		cmd.ui.Say(space.GUID)
	} else {
		cmd.ui.Say(T("Getting info for space {{.TargetSpace}} in org {{.OrgName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"TargetSpace": terminal.EntityNameColor(space.Name),
				"OrgName":     terminal.EntityNameColor(space.Organization.Name),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
			}))

		quotaString := cmd.quotaString(space)
		cmd.ui.Ok()
		cmd.ui.Say("")
		table := cmd.ui.Table([]string{terminal.EntityNameColor(space.Name), "", ""})
		table.Add("", T("Org:"), terminal.EntityNameColor(space.Organization.Name))

		apps := []string{}
		for _, app := range space.Applications {
			apps = append(apps, terminal.EntityNameColor(app.Name))
		}
		table.Add("", T("Apps:"), strings.Join(apps, ", "))

		domains := []string{}
		for _, domain := range space.Domains {
			domains = append(domains, terminal.EntityNameColor(domain.Name))
		}
		table.Add("", T("Domains:"), strings.Join(domains, ", "))

		services := []string{}
		for _, service := range space.ServiceInstances {
			services = append(services, terminal.EntityNameColor(service.Name))
		}
		table.Add("", T("Services:"), strings.Join(services, ", "))

		securityGroups := []string{}
		for _, group := range space.SecurityGroups {
			securityGroups = append(securityGroups, terminal.EntityNameColor(group.Name))
		}
		table.Add("", T("Security Groups:"), strings.Join(securityGroups, ", "))

		table.Add("", T("Space Quota:"), quotaString)

		table.Print()
	}
	if c.Bool("security-group-rules") {
		cmd.ui.Say("")
		for _, group := range space.SecurityGroups {
			cmd.ui.Say(T("Getting rules for the security group  : {{.SecurityGroupName}}...",
				map[string]interface{}{"SecurityGroupName": terminal.EntityNameColor(group.Name)}))
			table := cmd.ui.Table([]string{"", "", "", ""})
			for _, rules := range group.Rules {
				for ruleName, ruleValue := range rules {
					table.Add("", ruleName, ":", fmt.Sprintf("%v", ruleValue))
				}
				table.Add("", "", "", "")
			}
			table.Print()
		}
	}

}

func (cmd *ShowSpace) quotaString(space models.Space) string {
	if space.SpaceQuotaGUID == "" {
		return ""
	}

	quota, err := cmd.quotaRepo.FindByGUID(space.SpaceQuotaGUID)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return ""
	}

	spaceQuota := fmt.Sprintf(
		"%s (%s memory limit, %s instance memory limit, %d routes, %d services, paid services %s, %s app instance limit)",
		quota.Name,
		quota.FormattedMemoryLimit(),
		quota.FormattedInstanceMemoryLimit(),
		quota.RoutesLimit,
		quota.ServicesLimit,
		formatters.Allowed(quota.NonBasicServicesAllowed),
		T(quota.FormattedAppInstanceLimit()),
	)

	return spaceQuota
}

func (cmd *ShowSpace) populatePluginModel(space models.Space) {
	cmd.pluginModel.Name = space.Name
	cmd.pluginModel.Guid = space.GUID

	cmd.pluginModel.Organization.Name = space.Organization.Name
	cmd.pluginModel.Organization.Guid = space.Organization.GUID

	for _, app := range space.Applications {
		a := plugin_models.GetSpace_Apps{
			Name: app.Name,
			Guid: app.GUID,
		}
		cmd.pluginModel.Applications = append(cmd.pluginModel.Applications, a)
	}

	for _, domain := range space.Domains {
		d := plugin_models.GetSpace_Domains{
			Name: domain.Name,
			Guid: domain.GUID,
			OwningOrganizationGuid: domain.OwningOrganizationGUID,
			Shared:                 domain.Shared,
		}
		cmd.pluginModel.Domains = append(cmd.pluginModel.Domains, d)
	}

	for _, service := range space.ServiceInstances {
		si := plugin_models.GetSpace_ServiceInstance{
			Name: service.Name,
			Guid: service.GUID,
		}
		cmd.pluginModel.ServiceInstances = append(cmd.pluginModel.ServiceInstances, si)
	}
	for _, group := range space.SecurityGroups {
		sg := plugin_models.GetSpace_SecurityGroup{
			Name:  group.Name,
			Guid:  group.GUID,
			Rules: group.Rules,
		}
		cmd.pluginModel.SecurityGroups = append(cmd.pluginModel.SecurityGroups, sg)
	}

	quota, err := cmd.quotaRepo.FindByGUID(space.SpaceQuotaGUID)
	if err == nil {
		cmd.pluginModel.SpaceQuota.Name = quota.Name
		cmd.pluginModel.SpaceQuota.Guid = quota.GUID
		cmd.pluginModel.SpaceQuota.MemoryLimit = quota.MemoryLimit
		cmd.pluginModel.SpaceQuota.InstanceMemoryLimit = quota.InstanceMemoryLimit
		cmd.pluginModel.SpaceQuota.RoutesLimit = quota.RoutesLimit
		cmd.pluginModel.SpaceQuota.ServicesLimit = quota.ServicesLimit
		cmd.pluginModel.SpaceQuota.NonBasicServicesAllowed = quota.NonBasicServicesAllowed
	}
}
