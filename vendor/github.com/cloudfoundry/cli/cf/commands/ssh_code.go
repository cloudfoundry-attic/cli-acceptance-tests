package commands

import (
	"errors"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter . SSHCodeGetter

type SSHCodeGetter interface {
	commandregistry.Command
	Get() (string, error)
}

type OneTimeSSHCode struct {
	ui           terminal.UI
	config       coreconfig.ReadWriter
	authRepo     authentication.AuthenticationRepository
	endpointRepo coreconfig.EndpointRepository
}

func init() {
	commandregistry.Register(&OneTimeSSHCode{})
}

func (cmd *OneTimeSSHCode) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "ssh-code",
		Description: T("Get a one time password for ssh clients"),
		Usage: []string{
			T("CF_NAME ssh-code"),
		},
	}
}

func (cmd *OneTimeSSHCode) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewAPIEndpointRequirement(),
	}

	return reqs
}

func (cmd *OneTimeSSHCode) SetDependency(deps commandregistry.Dependency, _ bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.authRepo = deps.RepoLocator.GetAuthenticationRepository()
	cmd.endpointRepo = deps.RepoLocator.GetEndpointRepository()

	return cmd
}

func (cmd *OneTimeSSHCode) Execute(c flags.FlagContext) {
	code, err := cmd.Get()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say(code)
}

func (cmd *OneTimeSSHCode) Get() (string, error) {
	refresher := coreconfig.APIConfigRefresher{
		Endpoint:     cmd.config.APIEndpoint(),
		EndpointRepo: cmd.endpointRepo,
		Config:       cmd.config,
	}

	_, err := refresher.Refresh()
	if err != nil {
		return "", errors.New("Error refreshing config: " + err.Error())
	}

	token, err := cmd.authRepo.RefreshAuthToken()
	if err != nil {
		return "", errors.New(T("Error refreshing oauth token: ") + err.Error())
	}

	sshCode, err := cmd.authRepo.Authorize(token)
	if err != nil {
		return "", errors.New(T("Error getting SSH code: ") + err.Error())
	}

	return sshCode, nil
}
