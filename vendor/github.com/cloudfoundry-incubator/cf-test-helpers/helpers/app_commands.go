package helpers

import (
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
)

const CURL_TIMEOUT = 30 * time.Second

// Gets an app's endpoint with the specified path
func AppUri(appName, path string) string {
	config := LoadConfig()
	appsDomain := config.AppsDomain

	return config.Protocol() + appName + "." + appsDomain + path
}

// Gets an app's root endpoint
func AppRootUri(appName string) string {
	return AppUri(appName, "/")
}

// Curls an app's endpoint and exit successfully before the specified timeout
func CurlAppWithTimeout(appName, path string, timeout time.Duration, args ...string) string {
	uri := AppUri(appName, path)
	curlCmd := runner.Curl(append([]string{uri}, args...)...).Wait(timeout)
	Expect(curlCmd).To(Exit(0))
	Expect(string(curlCmd.Err.Contents())).To(HaveLen(0))
	return string(curlCmd.Out.Contents())
}

// Curls an app's endpoint and exit successfully before the default timeout
func CurlApp(appName, path string, args ...string) string {
	return CurlAppWithTimeout(appName, path, CURL_TIMEOUT, args...)
}

// Curls an app's root endpoint and exit successfully before the default timeout
func CurlAppRoot(appName string) string {
	return CurlApp(appName, "/")
}

// Returns a function that curls an app's root endpoint and exit successfully before the default timeout
func CurlingAppRoot(appName string) func() string {
	return func() string {
		return CurlAppRoot(appName)
	}
}
