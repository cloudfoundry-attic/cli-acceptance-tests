package organization

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/models"
)

type ShowOrg struct {
	ui          terminal.UI
	config      coreconfig.Reader
	orgReq      requirements.OrganizationRequirement
	pluginModel *plugin_models.GetOrg_Model
	pluginCall  bool
}

func init() {
	commandregistry.Register(&ShowOrg{})
}

func (cmd *ShowOrg) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &flags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given org's guid.  All other output for the org is suppressed.")}
	return commandregistry.CommandMetadata{
		Name:        "org",
		Description: T("Show org info"),
		Usage: []string{
			T("CF_NAME org ORG"),
		},
		Flags: fs,
	}
}

func (cmd *ShowOrg) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("org"))
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return reqs
}

func (cmd *ShowOrg) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.Organization
	return cmd
}

func (cmd *ShowOrg) Execute(c flags.FlagContext) {
	org := cmd.orgReq.GetOrganization()

	if c.Bool("guid") {
		cmd.ui.Say(org.GUID)
	} else {
		cmd.ui.Say(T("Getting info for org {{.OrgName}} as {{.Username}}...",
			map[string]interface{}{
				"OrgName":  terminal.EntityNameColor(org.Name),
				"Username": terminal.EntityNameColor(cmd.config.Username())}))
		cmd.ui.Ok()
		cmd.ui.Say("")

		table := cmd.ui.Table([]string{terminal.EntityNameColor(org.Name) + ":", "", ""})

		domains := []string{}
		for _, domain := range org.Domains {
			domains = append(domains, domain.Name)
		}

		spaces := []string{}
		for _, space := range org.Spaces {
			spaces = append(spaces, space.Name)
		}

		spaceQuotas := []string{}
		for _, spaceQuota := range org.SpaceQuotas {
			spaceQuotas = append(spaceQuotas, spaceQuota.Name)
		}

		quota := org.QuotaDefinition

		appInstanceLimit := strconv.Itoa(quota.AppInstanceLimit)
		if quota.AppInstanceLimit == resources.UnlimitedAppInstances {
			appInstanceLimit = T("unlimited")
		}

		orgQuotaFields := []string{
			T("{{.MemoryLimit}}M memory limit", map[string]interface{}{"MemoryLimit": quota.MemoryLimit}),
			T("{{.InstanceMemoryLimit}} instance memory limit", map[string]interface{}{"InstanceMemoryLimit": formatters.InstanceMemoryLimit(quota.InstanceMemoryLimit)}),
			T("{{.RoutesLimit}} routes", map[string]interface{}{"RoutesLimit": quota.RoutesLimit}),
			T("{{.ServicesLimit}} services", map[string]interface{}{"ServicesLimit": quota.ServicesLimit}),
			T("paid services {{.NonBasicServicesAllowed}}", map[string]interface{}{"NonBasicServicesAllowed": formatters.Allowed(quota.NonBasicServicesAllowed)}),
			T("{{.AppInstanceLimit}} app instance limit", map[string]interface{}{"AppInstanceLimit": appInstanceLimit}),
		}
		orgQuota := fmt.Sprintf("%s (%s)", quota.Name, strings.Join(orgQuotaFields, ", "))

		if cmd.pluginCall {
			cmd.populatePluginModel(org, quota)
		} else {
			table.Add("", T("domains:"), terminal.EntityNameColor(strings.Join(domains, ", ")))
			table.Add("", T("quota:"), terminal.EntityNameColor(orgQuota))
			table.Add("", T("spaces:"), terminal.EntityNameColor(strings.Join(spaces, ", ")))
			table.Add("", T("space quotas:"), terminal.EntityNameColor(strings.Join(spaceQuotas, ", ")))

			table.Print()
		}
	}
}

func (cmd *ShowOrg) populatePluginModel(org models.Organization, quota models.QuotaFields) {
	cmd.pluginModel.Name = org.Name
	cmd.pluginModel.Guid = org.GUID
	cmd.pluginModel.QuotaDefinition.Name = quota.Name
	cmd.pluginModel.QuotaDefinition.MemoryLimit = quota.MemoryLimit
	cmd.pluginModel.QuotaDefinition.InstanceMemoryLimit = quota.InstanceMemoryLimit
	cmd.pluginModel.QuotaDefinition.RoutesLimit = quota.RoutesLimit
	cmd.pluginModel.QuotaDefinition.ServicesLimit = quota.ServicesLimit
	cmd.pluginModel.QuotaDefinition.NonBasicServicesAllowed = quota.NonBasicServicesAllowed

	for _, domain := range org.Domains {
		d := plugin_models.GetOrg_Domains{
			Name: domain.Name,
			Guid: domain.GUID,
			OwningOrganizationGuid: domain.OwningOrganizationGUID,
			Shared:                 domain.Shared,
		}
		cmd.pluginModel.Domains = append(cmd.pluginModel.Domains, d)
	}

	for _, space := range org.Spaces {
		s := plugin_models.GetOrg_Space{
			Name: space.Name,
			Guid: space.GUID,
		}
		cmd.pluginModel.Spaces = append(cmd.pluginModel.Spaces, s)
	}

	for _, spaceQuota := range org.SpaceQuotas {
		sq := plugin_models.GetOrg_SpaceQuota{
			Name:                spaceQuota.Name,
			Guid:                spaceQuota.GUID,
			MemoryLimit:         spaceQuota.MemoryLimit,
			InstanceMemoryLimit: spaceQuota.InstanceMemoryLimit,
		}
		cmd.pluginModel.SpaceQuotas = append(cmd.pluginModel.SpaceQuotas, sq)
	}
}
