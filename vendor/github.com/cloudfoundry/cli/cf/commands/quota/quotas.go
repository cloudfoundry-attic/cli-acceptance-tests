package quota

import (
	"fmt"
	"strconv"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ListQuotas struct {
	ui        terminal.UI
	config    coreconfig.Reader
	quotaRepo quotas.QuotaRepository
}

func init() {
	commandregistry.Register(&ListQuotas{})
}

func (cmd *ListQuotas) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "quotas",
		Description: T("List available usage quotas"),
		Usage: []string{
			T("CF_NAME quotas"),
		},
	}
}

func (cmd *ListQuotas) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
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

func (cmd *ListQuotas) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetQuotaRepository()
	return cmd
}

func (cmd *ListQuotas) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Getting quotas as {{.Username}}...", map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	quotas, apiErr := cmd.quotaRepo.FindAll()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	table := cmd.ui.Table([]string{
		T("name"),
		T("total memory"),
		T("instance memory"),
		T("routes"),
		T("service instances"),
		T("paid plans"),
		T("app instances"),
		T("route ports"),
	})

	var megabytes string
	for _, quota := range quotas {
		if quota.InstanceMemoryLimit == -1 {
			megabytes = T("unlimited")
		} else {
			megabytes = formatters.ByteSize(quota.InstanceMemoryLimit * formatters.MEGABYTE)
		}

		servicesLimit := strconv.Itoa(quota.ServicesLimit)
		if quota.ServicesLimit == -1 {
			servicesLimit = T("unlimited")
		}

		appInstanceLimit := strconv.Itoa(quota.AppInstanceLimit)
		if quota.AppInstanceLimit == resources.UnlimitedAppInstances {
			appInstanceLimit = T("unlimited")
		}

		reservedRoutePorts := string(quota.ReservedRoutePorts)
		if reservedRoutePorts == resources.UnlimitedReservedRoutePorts {
			reservedRoutePorts = T("unlimited")
		}

		table.Add(
			quota.Name,
			formatters.ByteSize(quota.MemoryLimit*formatters.MEGABYTE),
			megabytes,
			fmt.Sprintf("%d", quota.RoutesLimit),
			fmt.Sprint(servicesLimit),
			formatters.Allowed(quota.NonBasicServicesAllowed),
			fmt.Sprint(appInstanceLimit),
			fmt.Sprint(reservedRoutePorts),
		)
	}

	table.Print()
}
