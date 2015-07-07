package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/plugin"
)

type Test1 struct {
}

func (c *Test1) Run(cliConnection plugin.CliConnection, args []string) {
	switch args[0] {
	case "GetCurrentOrg":
		result, _ := cliConnection.GetCurrentOrg()
		fmt.Println("Done GetCurrentOrg:", result)
	case "GetCurrentSpace":
		result, _ := cliConnection.GetCurrentSpace()
		fmt.Println("Done GetCurrentSpace:", result)
	case "Username":
		result, _ := cliConnection.Username()
		fmt.Println("Done Username:", result)
	case "UserGuid":
		result, _ := cliConnection.UserGuid()
		fmt.Println("Done UserGuid:", result)
	case "UserEmail":
		result, _ := cliConnection.UserEmail()
		fmt.Println("Done UserEmail:", result)
	case "IsLoggedIn":
		result, _ := cliConnection.IsLoggedIn()
		fmt.Println("Done IsLoggedIn:", result)
	case "IsSSLDisabled":
		_, err := cliConnection.IsSSLDisabled()
		if err != nil {
			fmt.Println("Error in IsSSLDisabled()", err)
		}
	case "ApiEndpoint":
		result, _ := cliConnection.ApiEndpoint()
		fmt.Println("Done ApiEndpoint:", result)
	case "ApiVersion":
		result, _ := cliConnection.ApiVersion()
		fmt.Println("Done ApiVersion:", result)
	case "HasAPIEndpoint":
		_, err := cliConnection.HasAPIEndpoint()
		if err != nil {
			fmt.Println("Error in HasAPIEndpoint()", err)
		}
	case "HasOrganization":
		result, _ := cliConnection.HasOrganization()
		fmt.Println("Done HasOrganization:", result)
	case "HasSpace":
		result, _ := cliConnection.HasSpace()
		fmt.Println("Done HasSpace:", result)
	case "LoggregatorEndpoint":
		result, _ := cliConnection.LoggregatorEndpoint()
		fmt.Println("Done LoggregatorEndpoint:", result)
	case "DopplerEndpoint":
		result, _ := cliConnection.DopplerEndpoint()
		fmt.Println("Done DopplerEndpoint:", result)
	case "AccessToken":
		result, _ := cliConnection.AccessToken()
		fmt.Println("Done AccessToken:", result)
	case "GetOrg":
		result, _ := cliConnection.GetOrg(args[1])
		fmt.Println("Done GetOrg:", result)
	case "GetOrgs":
		result, _ := cliConnection.GetOrgs()
		fmt.Println("Done GetOrgs:", result)
	}

	// } else if args[0] == "CLI-MESSAGE-UNINSTALL" {
	// uninstalling()
	// }
}

func (c *Test1) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "GatsPlugin",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 2,
			Build: 4,
		},
		MinCliVersion: plugin.VersionType{
			Major: 5,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			{Name: "GetCurrentSpace"},
			{Name: "GetCurrentOrg"},
			{Name: "Username"},
			{Name: "UserGuid"},
			{Name: "UserEmail"},
			{Name: "IsLoggedIn"},
			{Name: "IsSSLDisabled"},
			{Name: "ApiEndpoint"},
			{Name: "ApiVersion"},
			{Name: "HasAPIEndpoint"},
			{Name: "HasOrganization"},
			{Name: "HasSpace"},
			{Name: "LoggregatorEndpoint"},
			{Name: "DopplerEndpoint"},
			{Name: "AccessToken"},
			{Name: "GetOrg"},
			{Name: "GetOrgs"},
		},
	}
}

func uninstalling() {
	os.Remove(filepath.Join(os.TempDir(), "uninstall-test-file-for-test_1.exe"))
}

func main() {
	plugin.Start(new(Test1))
}
