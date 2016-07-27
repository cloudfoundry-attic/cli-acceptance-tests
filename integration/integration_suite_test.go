package integration

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	setAPI()
	RunSpecs(t, "Integration Suite")
}

func setAPI() {
	api := exec.Command("cf", "api", "api.bosh-lite.com", "--skip-ssl-validation")
	apiSession, err := Start(api, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(apiSession).Should(Exit(0))
}
