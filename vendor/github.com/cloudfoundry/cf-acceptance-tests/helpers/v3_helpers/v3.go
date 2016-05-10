package v3_helpers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/cloudfoundry-incubator/cf-test-helpers/runner"

	. "github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/onsi/gomega"
	. "github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/onsi/gomega/gbytes"
	. "github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/onsi/gomega/gexec"
)

func StartApp(appGuid string) {
	startURL := fmt.Sprintf("/v3/apps/%s/start", appGuid)
	Expect(cf.Cf("curl", startURL, "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
}

func StopApp(appGuid string) {
	stopURL := fmt.Sprintf("/v3/apps/%s/stop", appGuid)
	Expect(cf.Cf("curl", stopURL, "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
}

func CreateApp(appName, spaceGuid, environmentVariables string) string {
	session := cf.Cf("curl", "/v3/apps", "-X", "POST", "-d", fmt.Sprintf(`{"name":"%s", "relationships": {"space": {"guid": "%s"}}, "environment_variables":%s}`, appName, spaceGuid, environmentVariables))
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	var app struct {
		Guid string `json:"guid"`
	}
	json.Unmarshal(bytes, &app)
	return app.Guid
}

func CreateDockerApp(appName, spaceGuid, environmentVariables string) string {
	session := cf.Cf("curl", "/v3/apps", "-X", "POST", "-d", fmt.Sprintf(`{"name":"%s", "relationships": {"space": {"guid": "%s"}}, "environment_variables":%s, "lifecycle": {"type": "docker", "data": {} } }`, appName, spaceGuid, environmentVariables))
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	var app struct {
		Guid string `json:"guid"`
	}
	json.Unmarshal(bytes, &app)
	return app.Guid
}

func DeleteApp(appGuid string) {
	session := cf.Cf("curl", fmt.Sprintf("/v3/apps/%s", appGuid), "-X", "DELETE", "-v")
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	Expect(bytes).To(ContainSubstring("204 No Content"))
}

func WaitForPackageToBeReady(packageGuid string) {
	pkgUrl := fmt.Sprintf("/v3/packages/%s", packageGuid)
	Eventually(func() *Session {
		session := cf.Cf("curl", pkgUrl)
		Expect(session.Wait(DEFAULT_TIMEOUT)).To(Exit(0))
		return session
	}, LONG_CURL_TIMEOUT).Should(Say("READY"))
}

func WaitForDropletToStage(dropletGuid string) {
	dropletPath := fmt.Sprintf("/v3/droplets/%s", dropletGuid)
	Eventually(func() *Session {
		return cf.Cf("curl", dropletPath).Wait(DEFAULT_TIMEOUT)
	}, CF_PUSH_TIMEOUT).Should(Say("STAGED"))
}

func CreatePackage(appGuid string) string {
	packageCreateUrl := fmt.Sprintf("/v3/apps/%s/packages", appGuid)
	session := cf.Cf("curl", packageCreateUrl, "-X", "POST", "-d", fmt.Sprintf(`{"type":"bits"}`))
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	var pac struct {
		Guid string `json:"guid"`
	}
	json.Unmarshal(bytes, &pac)
	return pac.Guid
}

func CreateDockerPackage(appGuid, imagePath string) string {
	packageCreateUrl := fmt.Sprintf("/v3/apps/%s/packages", appGuid)
	session := cf.Cf("curl", packageCreateUrl, "-X", "POST", "-d", fmt.Sprintf(`{"type":"docker", "data": {"image": "%s"}}`, imagePath))
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	var pac struct {
		Guid string `json:"guid"`
	}
	json.Unmarshal(bytes, &pac)
	return pac.Guid
}

func GetSpaceGuidFromName(spaceName string) string {
	session := cf.Cf("space", spaceName, "--guid")
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	return strings.TrimSpace(string(bytes))
}

func GetAuthToken() string {
	bytes := runner.Run("bash", "-c", "cf oauth-token | grep bearer").Wait(DEFAULT_TIMEOUT).Out.Contents()
	return strings.TrimSpace(string(bytes))
}

func UploadPackage(uploadUrl, packageZipPath, token string) {
	bits := fmt.Sprintf(`bits=@%s`, packageZipPath)
	curl := runner.Curl("-v", "-s", uploadUrl, "-F", bits, "-H", fmt.Sprintf("Authorization: %s", token)).Wait(DEFAULT_TIMEOUT)
	Expect(curl).To(Exit(0))
}

func StageBuildpackPackage(packageGuid, buildpack string) string {
	stageBody := fmt.Sprintf(`{"lifecycle":{ "type": "buildpack", "data": { "buildpack": "%s" } }}`, buildpack)
	stageUrl := fmt.Sprintf("/v3/packages/%s/droplets", packageGuid)
	session := cf.Cf("curl", stageUrl, "-X", "POST", "-d", stageBody)
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	var droplet struct {
		Guid string `json:"guid"`
	}
	json.Unmarshal(bytes, &droplet)
	return droplet.Guid
}

func StageDockerPackage(packageGuid string) string {
	stageUrl := fmt.Sprintf("/v3/packages/%s/droplets", packageGuid)
	session := cf.Cf("curl", stageUrl, "-X", "POST", "-d", "")
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	var droplet struct {
		Guid string `json:"guid"`
	}
	json.Unmarshal(bytes, &droplet)
	return droplet.Guid
}

func CreateAndMapRoute(appGuid, space, domain, host string) {
	Expect(cf.Cf("create-route", space, domain, "-n", host).Wait(DEFAULT_TIMEOUT)).To(Exit(0))
	getRoutePath := fmt.Sprintf("/v2/routes?q=host:%s", host)
	routeBody := cf.Cf("curl", getRoutePath).Wait(DEFAULT_TIMEOUT).Out.Contents()
	routeJSON := struct {
		Resources []struct {
			Metadata struct {
				Guid string `json:"guid"`
			} `json:"metadata"`
		} `json:"resources"`
	}{}
	json.Unmarshal([]byte(routeBody), &routeJSON)
	routeGuid := routeJSON.Resources[0].Metadata.Guid
	addRouteBody := fmt.Sprintf(`
	{
		"relationships": {
			"app":   {"guid": "%s"},
			"route": {"guid": "%s"}
		}
	}`, appGuid, routeGuid)
	Expect(cf.Cf("curl", "/v3/route_mappings", "-X", "POST", "-d", addRouteBody).Wait(DEFAULT_TIMEOUT)).To(Exit(0))
}

func AssignDropletToApp(appGuid, dropletGuid string) {
	appUpdatePath := fmt.Sprintf("/v3/apps/%s/droplets/current", appGuid)
	appUpdateBody := fmt.Sprintf(`{"droplet_guid":"%s"}`, dropletGuid)
	Expect(cf.Cf("curl", appUpdatePath, "-X", "PUT", "-d", appUpdateBody).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

	for _, process := range GetProcesses(appGuid, "") {
		ScaleProcess(appGuid, process.Type, DEFAULT_MEMORY_LIMIT)
	}
}

func FetchRecentLogs(appGuid, oauthToken string, config helpers.Config) *Session {
	loggregatorEndpoint := strings.Replace(config.ApiEndpoint, "api", "loggregator", -1)
	logUrl := fmt.Sprintf("%s/recent?app=%s", loggregatorEndpoint, appGuid)
	session := runner.Curl(logUrl, "-H", fmt.Sprintf("Authorization: %s", oauthToken))
	Expect(session.Wait(DEFAULT_TIMEOUT)).To(Exit(0))
	return session
}

func ScaleProcess(appGuid, processType, memoryInMb string) {
	scalePath := fmt.Sprintf("/v3/apps/%s/processes/%s/scale", appGuid, processType)
	scaleBody := fmt.Sprintf(`{"memory_in_mb":"%s"}`, memoryInMb)
	Expect(cf.Cf("curl", scalePath, "-X", "PUT", "-d", scaleBody).Wait(DEFAULT_TIMEOUT)).To(Exit(0))
}
