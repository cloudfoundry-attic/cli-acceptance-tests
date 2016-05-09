package service

import (
	"encoding/json"
	"strings"

	"github.com/blang/semver"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/util"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type UpdateUserProvidedService struct {
	ui                              terminal.UI
	config                          coreconfig.Reader
	userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository
	serviceInstanceReq              requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&UpdateUserProvidedService{})
}

func (cmd *UpdateUserProvidedService) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["p"] = &flags.StringFlag{ShortName: "p", Usage: T("Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications")}
	fs["l"] = &flags.StringFlag{ShortName: "l", Usage: T("URL to which logs for bound applications will be streamed")}
	fs["r"] = &flags.StringFlag{ShortName: "r", Usage: T("URL to which requests for bound routes will be forwarded. Scheme for this URL must be https")}

	return commandregistry.CommandMetadata{
		Name:        "update-user-provided-service",
		ShortName:   "uups",
		Description: T("Update user-provided service instance"),
		Usage: []string{
			T(`CF_NAME update-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-l SYSLOG_DRAIN_URL] [-r ROUTE_SERVICE_URL]

   Pass comma separated credential parameter names to enable interactive mode:
   CF_NAME update-user-provided-service SERVICE_INSTANCE -p "comma, separated, parameter, names"

   Pass credential parameters as JSON to create a service non-interactively:
   CF_NAME update-user-provided-service SERVICE_INSTANCE -p '{"key1":"value1","key2":"value2"}'

   Specify a path to a file containing JSON:
   CF_NAME update-user-provided-service SERVICE_INSTANCE -p PATH_TO_FILE`),
		},
		Examples: []string{
			`CF_NAME update-user-provided-service my-db-mine -p '{"username":"admin", "password":"pa55woRD"}'`,
			"CF_NAME update-user-provided-service my-db-mine -p /path/to/credentials.json",
			"CF_NAME update-user-provided-service my-drain-service -l syslog://example.com",
			"CF_NAME update-user-provided-service my-route-service -r https://example.com",
		},
		Flags: fs,
	}
}

func (cmd *UpdateUserProvidedService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("update-user-provided-service"))
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.serviceInstanceReq,
	}

	if fc.IsSet("r") {
		minAPIVersion, err := semver.Make("2.51.0")
		if err != nil {
			panic(err.Error())
		}

		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '-r'", minAPIVersion))
	}

	return reqs
}

func (cmd *UpdateUserProvidedService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.userProvidedServiceInstanceRepo = deps.RepoLocator.GetUserProvidedServiceInstanceRepository()
	return cmd
}

func (cmd *UpdateUserProvidedService) Execute(c flags.FlagContext) {
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()
	if !serviceInstance.IsUserProvided() {
		cmd.ui.Failed(T("Service Instance is not user provided"))
		return
	}

	drainURL := c.String("l")
	credentials := strings.Trim(c.String("p"), `'"`)
	routeServiceURL := c.String("r")

	credentialsMap := make(map[string]interface{})

	if c.IsSet("p") {
		jsonBytes, err := util.GetContentsFromFlagValue(credentials)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}

		err = json.Unmarshal(jsonBytes, &credentialsMap)
		if err != nil {
			for _, param := range strings.Split(credentials, ",") {
				param = strings.Trim(param, " ")
				credentialsMap[param] = cmd.ui.Ask(param)
			}
		}
	}

	cmd.ui.Say(T("Updating user provided service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceInstance.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceInstance.Params = credentialsMap
	serviceInstance.SysLogDrainURL = drainURL
	serviceInstance.RouteServiceURL = routeServiceURL

	apiErr := cmd.userProvidedServiceInstanceRepo.Update(serviceInstance.ServiceInstanceFields)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("TIP: Use '{{.CFRestageCommand}}' for any bound apps to ensure your env variable changes take effect",
		map[string]interface{}{
			"CFRestageCommand": terminal.CommandColor(cf.Name + " restage"),
		}))

	if routeServiceURL == "" && credentials == "" && drainURL == "" {
		cmd.ui.Warn(T("No flags specified. No changes were made."))
	}
}
