package service

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/models"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ListServices struct {
	ui                 terminal.UI
	config             coreconfig.Reader
	serviceSummaryRepo api.ServiceSummaryRepository
	pluginModel        *[]plugin_models.GetServices_Model
	pluginCall         bool
}

func init() {
	commandregistry.Register(&ListServices{})
}

func (cmd *ListServices) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "services",
		ShortName:   "s",
		Description: T("List all service instances in the target space"),
		Usage: []string{
			"CF_NAME services",
		},
	}
}

func (cmd *ListServices) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	return reqs
}

func (cmd *ListServices) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceSummaryRepo = deps.RepoLocator.GetServiceSummaryRepository()
	cmd.pluginModel = deps.PluginModels.Services
	cmd.pluginCall = pluginCall
	return cmd
}

func (cmd *ListServices) Execute(fc flags.FlagContext) {
	cmd.ui.Say(T("Getting services in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceInstances, apiErr := cmd.serviceSummaryRepo.GetSummariesInCurrentSpace()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(serviceInstances) == 0 {
		cmd.ui.Say(T("No services found"))
		return
	}

	table := cmd.ui.Table([]string{T("name"), T("service"), T("plan"), T("bound apps"), T("last operation")})

	for _, instance := range serviceInstances {
		var serviceColumn string
		var serviceStatus string

		if instance.IsUserProvided() {
			serviceColumn = T("user-provided")
		} else {
			serviceColumn = instance.ServiceOffering.Label
		}
		serviceStatus = ServiceInstanceStateToStatus(instance.LastOperation.Type, instance.LastOperation.State, instance.IsUserProvided())

		table.Add(
			instance.Name,
			serviceColumn,
			instance.ServicePlan.Name,
			strings.Join(instance.ApplicationNames, ", "),
			serviceStatus,
		)
		if cmd.pluginCall {
			s := plugin_models.GetServices_Model{
				Name: instance.Name,
				Guid: instance.GUID,
				ServicePlan: plugin_models.GetServices_ServicePlan{
					Name: instance.ServicePlan.Name,
					Guid: instance.ServicePlan.GUID,
				},
				Service: plugin_models.GetServices_ServiceFields{
					Name: instance.ServiceOffering.Label,
				},
				ApplicationNames: instance.ApplicationNames,
				LastOperation: plugin_models.GetServices_LastOperation{
					Type:  instance.LastOperation.Type,
					State: instance.LastOperation.State,
				},
				IsUserProvided: instance.IsUserProvided(),
			}

			*(cmd.pluginModel) = append(*(cmd.pluginModel), s)
		}

	}

	table.Print()
}
