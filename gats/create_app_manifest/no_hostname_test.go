package create_app_manifest_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	acceptanceTestHelpers "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	gatsHelpers "github.com/cloudfoundry/cli-acceptance-tests/gats/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("CreateAppManifest", func() {
	const (
		assertionTimeout         = 10 * time.Second
		appTimeout               = 5 * time.Minute
		createAppManifestTimeout = 20 * time.Second
		manifestFileName         = "create-app-manifest-test-manifest.yml"
	)

	var (
		context      *acceptanceTestHelpers.ConfiguredContext
		env          *acceptanceTestHelpers.Environment
		manifestPath string
	)

	config := acceptanceTestHelpers.LoadConfig()

	BeforeEach(func() {
		manifestPath = filepath.Join(os.TempDir(), manifestFileName)

		context = acceptanceTestHelpers.NewContext(config)
		env = acceptanceTestHelpers.NewEnvironment(context)

		env.Setup()
	})

	AfterEach(func() {
		env.Teardown()
	})

	It("includes a no-hostname: true section for apps pushed with no hostname", func() {
		AsUser(context.AdminUserContext(), 180*time.Second, func() {
			space := context.RegularUserContext().Space
			org := context.RegularUserContext().Org
			domainName := fmt.Sprintf("%s.com", generator.RandomName())

			target := Cf("target", "-o", org, "-s", space).Wait(assertionTimeout)
			Expect(target.ExitCode()).To(Equal(0))

			createDomain := Cf("create-domain", org, domainName).Wait(assertionTimeout)
			Expect(createDomain.ExitCode()).To(Equal(0))

			appName1 := generator.RandomName()
			app1 := Cf("push", appName1, "-p", gatsHelpers.NewAssets().ServiceBroker, "-d", domainName, "--no-hostname").Wait(appTimeout)
			Expect(app1).To(Exit(0))

			createAppManifest := Cf("create-app-manifest", appName1, "-p", manifestPath).Wait(createAppManifestTimeout)
			Expect(createAppManifest).To(Exit(0))

			manifestContents, err := ioutil.ReadFile(manifestPath)
			Expect(err).NotTo(HaveOccurred())

			manifest := string(manifestContents)

			Expect(manifest).To(ContainSubstring("routes"))
			Expect(manifest).To(ContainSubstring(domainName))
		})
	})
})
