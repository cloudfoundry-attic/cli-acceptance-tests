package service

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/json"
)

//go:generate counterfeiter . ServiceBinder

type ServiceBinder interface {
	BindApplication(app models.Application, serviceInstance models.ServiceInstance, paramsMap map[string]interface{}) (apiErr error)
}

type BindService struct {
	ui                 terminal.UI
	config             coreconfig.Reader
	serviceBindingRepo api.ServiceBindingRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&BindService{})
}

func (cmd *BindService) MetaData() commandregistry.CommandMetadata {
	baseUsage := T("CF_NAME bind-service APP_NAME SERVICE_INSTANCE [-c PARAMETERS_AS_JSON]")
	paramsUsage := T(`   Optionally provide service-specific configuration parameters in a valid JSON object in-line:

   CF_NAME bind-service APP_NAME SERVICE_INSTANCE -c '{"name":"value","name":"value"}'

   Optionally provide a file containing service-specific configuration parameters in a valid JSON object. 
   The path to the parameters file can be an absolute or relative path to a file.
   CF_NAME bind-service APP_NAME SERVICE_INSTANCE -c PATH_TO_FILE

   Example of valid JSON object:
   {
      "permissions": "read-only"
   }`)

	fs := make(map[string]flags.FlagSet)
	fs["c"] = &flags.StringFlag{ShortName: "c", Usage: T("Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.")}

	return commandregistry.CommandMetadata{
		Name:        "bind-service",
		ShortName:   "bs",
		Description: T("Bind a service instance to an app"),
		Usage: []string{
			baseUsage,
			"\n\n",
			paramsUsage,
		},
		Examples: []string{
			fmt.Sprintf("%s:", T(`Linux/Mac`)),
			`   CF_NAME bind-service myapp mydb -c '{"permissions":"read-only"}'`,
			``,
			fmt.Sprintf("%s:", T(`Windows Command Line`)),
			`   CF_NAME bind-service myapp mydb -c "{\"permissions\":\"read-only\"}"`,
			``,
			fmt.Sprintf("%s:", T(`Windows PowerShell`)),
			`   CF_NAME bind-service myapp mydb -c '{\"permissions\":\"read-only\"}'`,
			``,
			`CF_NAME bind-service myapp mydb -c ~/workspace/tmp/instance_config.json`,
		},
		Flags: fs,
	}
}

func (cmd *BindService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME and SERVICE_INSTANCE as arguments\n\n") + commandregistry.Commands.CommandUsage("bind-service"))
	}

	serviceName := fc.Args()[1]

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])
	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(serviceName)

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.appReq,
		cmd.serviceInstanceReq,
	}

	return reqs
}

func (cmd *BindService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceBindingRepo = deps.RepoLocator.GetServiceBindingRepository()
	return cmd
}

func (cmd *BindService) Execute(c flags.FlagContext) {
	app := cmd.appReq.GetApplication()
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()
	params := c.String("c")

	paramsMap, err := json.ParseJSONFromFileOrString(params)
	if err != nil {
		cmd.ui.Failed(T("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
	}

	cmd.ui.Say(T("Binding service {{.ServiceInstanceName}} to app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name),
			"AppName":             terminal.EntityNameColor(app.Name),
			"OrgName":             terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":           terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.BindApplication(app, serviceInstance, paramsMap)
	if err != nil {
		if httperr, ok := err.(errors.HTTPError); ok && httperr.ErrorCode() == errors.ServiceBindingAppServiceTaken {
			cmd.ui.Ok()
			cmd.ui.Warn(T("App {{.AppName}} is already bound to {{.ServiceName}}.",
				map[string]interface{}{
					"AppName":     app.Name,
					"ServiceName": serviceInstance.Name,
				}))
			return
		}

		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("TIP: Use '{{.CFCommand}} {{.AppName}}' to ensure your env variable changes take effect",
		map[string]interface{}{"CFCommand": terminal.CommandColor(cf.Name + " restage"), "AppName": app.Name}))
}

func (cmd *BindService) BindApplication(app models.Application, serviceInstance models.ServiceInstance, paramsMap map[string]interface{}) (apiErr error) {
	apiErr = cmd.serviceBindingRepo.Create(serviceInstance.GUID, app.GUID, paramsMap)
	return
}
