package integration

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
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
	apiURL string

	// Per Test Level
	homeDir string
)

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func(_ []byte) {
	//Ginkgo Globals
	SetDefaultEventuallyTimeout(3 * time.Second)

	//Setup common environment variables
	apiURL = os.Getenv("CF_API")
})

var _ = BeforeEach(func() {
	setHomeDir()
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

func destroyHomeDir() {
	if homeDir != "" {
		os.RemoveAll(homeDir)
	}
}

func setAPI() {
	if apiURL == "" {
		apiURL = "api.bosh-lite.com"
	}
	api := exec.Command("cf", "api", apiURL, "--skip-ssl-validation")
	apiSession, err := Start(api, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(apiSession).Should(Exit(0))
}
