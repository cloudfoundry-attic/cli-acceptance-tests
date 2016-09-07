package integration

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var (
	// Suite Level
	apiURL            string
	skipSSLValidation string
	originalColor     string

	// Per Test Level
	homeDir string
)

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func(_ []byte) {
	// Ginkgo Globals
	SetDefaultEventuallyTimeout(3 * time.Second)

	// Setup common environment variables
	apiURL = os.Getenv("CF_API")
	turnOffColors()
})

var _ = SynchronizedAfterSuite(func() {},
	func() {
		setColor()
	})

var _ = BeforeEach(func() {
	setHomeDir()
	setSkipSSLValidation()
	setAPI()
})

var _ = AfterEach(func() {
	destroyHomeDir()
})

func setHomeDir() {
	var err error
	homeDir, err = ioutil.TempDir("", "cli-gats-test")
	Expect(err).NotTo(HaveOccurred())

	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", homeDir)
	} else {
		os.Setenv("HOME", homeDir)
	}
}

func setSkipSSLValidation() {
	if skip, err := strconv.ParseBool(os.Getenv("SKIP_SSL_VALIDATION")); err == nil && !skip {
		skipSSLValidation = ""
		return
	}
	skipSSLValidation = "--skip-ssl-validation"
}

func getAPI() string {
	if apiURL == "" {
		apiURL = "api.bosh-lite.com"
	}
	return apiURL
}

func setAPI() {
	api := exec.Command("cf", "api", getAPI(), skipSSLValidation)
	apiSession, err := Start(api, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(apiSession).Should(Exit(0))
}

func destroyHomeDir() {
	if homeDir != "" {
		os.RemoveAll(homeDir)
	}
}

func turnOffColors() {
	originalColor = os.Getenv("CF_COLOR")
	os.Setenv("CF_COLOR", "false")
}

func setColor() {
	os.Setenv("CF_COLOR", originalColor)
}
