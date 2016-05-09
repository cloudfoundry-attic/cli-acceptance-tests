package pluginrepo

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/actors/pluginrepo"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/models"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type RepoPlugins struct {
	ui         terminal.UI
	config     coreconfig.Reader
	pluginRepo pluginrepo.PluginRepo
}

func init() {
	commandregistry.Register(&RepoPlugins{})
}

func (cmd *RepoPlugins) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["r"] = &flags.StringFlag{ShortName: "r", Usage: T("Name of a registered repository")}

	return commandregistry.CommandMetadata{
		Name:        T("repo-plugins"),
		Description: T("List all available plugins in specified repository or in all added repositories"),
		Usage: []string{
			T(`CF_NAME repo-plugins [-r REPO_NAME]`),
		},
		Examples: []string{
			"CF_NAME repo-plugins -r PrivateRepo",
		},
		Flags: fs,
	}
}

func (cmd *RepoPlugins) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	reqs := []requirements.Requirement{}
	return reqs
}

func (cmd *RepoPlugins) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.pluginRepo = deps.PluginRepo
	return cmd
}

func (cmd *RepoPlugins) Execute(c flags.FlagContext) {
	var repos []models.PluginRepo
	repoName := c.String("r")

	repos = cmd.config.PluginRepos()
	for i, _ := range repos {
		if repos[i].URL == "http://plugins.cloudfoundry.org" {
			repos[i].URL = "https://plugins.cloudfoundry.org"
		}
	}

	if repoName == "" {
		cmd.ui.Say(T("Getting plugins from all repositories ... "))
	} else {
		index := cmd.findRepoIndex(repoName)
		if index != -1 {
			cmd.ui.Say(T("Getting plugins from repository '") + repoName + "' ...")
			repos = []models.PluginRepo{repos[index]}
		} else {
			cmd.ui.Failed(repoName + T(" does not exist as an available plugin repo."+"\nTip: use `add-plugin-repo` command to add repos."))
		}
	}

	cmd.ui.Say("")

	repoPlugins, repoError := cmd.pluginRepo.GetPlugins(repos)

	cmd.printTable(repoPlugins)

	cmd.printErrors(repoError)
}

func (cmd RepoPlugins) printTable(repoPlugins map[string][]clipr.Plugin) {
	for k, plugins := range repoPlugins {
		cmd.ui.Say(terminal.ColorizeBold(T("Repository: ")+k, 33))
		table := cmd.ui.Table([]string{T("name"), T("version"), T("description")})
		for _, p := range plugins {
			table.Add(p.Name, p.Version, p.Description)
		}
		table.Print()
		cmd.ui.Say("")
	}
}

func (cmd RepoPlugins) printErrors(repoError []string) {
	if len(repoError) > 0 {
		cmd.ui.Say(terminal.ColorizeBold(T("Logged errors:"), 31))
		for _, e := range repoError {
			cmd.ui.Say(terminal.Colorize(e, 31))
		}
		cmd.ui.Say("")
	}
}

func (cmd RepoPlugins) findRepoIndex(repoName string) int {
	repos := cmd.config.PluginRepos()
	for i, repo := range repos {
		if strings.ToLower(repo.Name) == strings.ToLower(repoName) {
			return i
		}
	}
	return -1
}
