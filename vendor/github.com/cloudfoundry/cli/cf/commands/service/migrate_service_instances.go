package service

import (
	"fmt"

	"github.com/blang/semver"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type MigrateServiceInstances struct {
	ui          terminal.UI
	configRepo  coreconfig.Reader
	serviceRepo api.ServiceRepository
}

func init() {
	commandregistry.Register(&MigrateServiceInstances{})
}

func (cmd *MigrateServiceInstances) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force migration without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "migrate-service-instances",
		Description: T("Migrate service instances from one service plan to another"),
		Usage: []string{
			T("CF_NAME migrate-service-instances v1_SERVICE v1_PROVIDER v1_PLAN v2_SERVICE v2_PLAN\n\n"),
			migrateServiceInstanceWarning(),
		},
		Flags: fs,
	}
}

func (cmd *MigrateServiceInstances) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 5 {
		cmd.ui.Failed(T("Incorrect Usage. Requires v1_SERVICE v1_PROVIDER v1_PLAN v2_SERVICE v2_PLAN as arguments\n\n") + commandregistry.Commands.CommandUsage("migrate-service-instances"))
	}

	maximumVersion, err := semver.Make("2.46.0")
	if err != nil {
		panic(err.Error())
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewMaxAPIVersionRequirement("migrate-service-instances", maximumVersion),
	}

	return reqs
}

func (cmd *MigrateServiceInstances) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	return cmd
}

func migrateServiceInstanceWarning() string {
	return T("WARNING: This operation is internal to Cloud Foundry; service brokers will not be contacted and resources for service instances will not be altered. The primary use case for this operation is to replace a service broker which implements the v1 Service Broker API with a broker which implements the v2 API by remapping service instances from v1 plans to v2 plans.  We recommend making the v1 plan private or shutting down the v1 broker to prevent additional instances from being created. Once service instances have been migrated, the v1 services and plans can be removed from Cloud Foundry.")
}

func (cmd *MigrateServiceInstances) Execute(c flags.FlagContext) {
	v1 := resources.ServicePlanDescription{
		ServiceLabel:    c.Args()[0],
		ServiceProvider: c.Args()[1],
		ServicePlanName: c.Args()[2],
	}
	v2 := resources.ServicePlanDescription{
		ServiceLabel:    c.Args()[3],
		ServicePlanName: c.Args()[4],
	}
	force := c.Bool("f")

	v1GUID, apiErr := cmd.serviceRepo.FindServicePlanByDescription(v1)
	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Failed(T("Plan {{.ServicePlanName}} cannot be found",
			map[string]interface{}{
				"ServicePlanName": terminal.EntityNameColor(v1.String()),
			}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	v2GUID, apiErr := cmd.serviceRepo.FindServicePlanByDescription(v2)
	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Failed(T("Plan {{.ServicePlanName}} cannot be found",
			map[string]interface{}{
				"ServicePlanName": terminal.EntityNameColor(v2.String()),
			}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	count, apiErr := cmd.serviceRepo.GetServiceInstanceCountForServicePlan(v1GUID)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	} else if count == 0 {
		cmd.ui.Failed(T("Plan {{.ServicePlanName}} has no service instances to migrate", map[string]interface{}{"ServicePlanName": terminal.EntityNameColor(v1.String())}))
		return
	}

	cmd.ui.Warn(migrateServiceInstanceWarning())

	serviceInstancesPhrase := pluralizeServiceInstances(count)

	if !force {
		response := cmd.ui.Confirm(
			T("Really migrate {{.ServiceInstanceDescription}} from plan {{.OldServicePlanName}} to {{.NewServicePlanName}}?>",
				map[string]interface{}{
					"ServiceInstanceDescription": serviceInstancesPhrase,
					"OldServicePlanName":         terminal.EntityNameColor(v1.String()),
					"NewServicePlanName":         terminal.EntityNameColor(v2.String()),
				}))
		if !response {
			return
		}
	}

	cmd.ui.Say(T("Attempting to migrate {{.ServiceInstanceDescription}}...", map[string]interface{}{"ServiceInstanceDescription": serviceInstancesPhrase}))

	changedCount, apiErr := cmd.serviceRepo.MigrateServicePlanFromV1ToV2(v1GUID, v2GUID)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Say(T("{{.CountOfServices}} migrated.", map[string]interface{}{"CountOfServices": pluralizeServiceInstances(changedCount)}))
	cmd.ui.Ok()

	return
}

func pluralizeServiceInstances(count int) string {
	var phrase string
	if count == 1 {
		phrase = T("service instance")
	} else {
		phrase = T("service instances")
	}

	return fmt.Sprintf("%d %s", count, phrase)
}
