package application_test

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	catshelpers "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry/cli-acceptance-tests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Push", func() {
	var (
		assets        helpers.Assets
		setupTimeout  time.Duration
		targetTimeout time.Duration
		pushTimeout   time.Duration
		context       *catshelpers.ConfiguredContext
		env           *catshelpers.Environment
	)

	BeforeEach(func() {
		assets = helpers.NewAssets()
		setupTimeout = 20 * time.Second
		targetTimeout = 10 * time.Second
		pushTimeout = 1 * time.Minute

		config := catshelpers.LoadConfig()
		context = catshelpers.NewContext(config)
		env = catshelpers.NewEnvironment(context)

		env.Setup()
	})

	AfterEach(func() {
		env.Teardown()
	})

	Context("when pushing an app with >260 character paths", func() {
		var (
			fullPath string
			cwd      string
		)

		BeforeEach(func() {
			dirName := "dir_name"
			dirNames := []string{}
			for i := 0; i < 32; i++ { // minimum 300 chars, including separators
				dirNames = append(dirNames, dirName)
			}

			longPath := filepath.Join(dirNames...)
			fullPath = filepath.Join(assets.ServiceBroker, longPath)

			if runtime.GOOS == "windows" {
				var err error
				cwd, err = os.Getwd()
				Expect(err).NotTo(HaveOccurred())

				// `\\?\` is used to skip Windows' file name processor, which imposes
				// length limits. Search MSDN for 'Maximum Path Length Limitation' for
				// more.
				err = os.MkdirAll(`\\?\`+filepath.Join(cwd, fullPath), os.ModeDir|os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			} else {
				err := os.MkdirAll(fullPath, os.ModeDir|os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		AfterEach(func() {
			if runtime.GOOS == "windows" {
				// `\\?\` is used to skip Windows' file name processor, which imposes
				// length limits. Search MSDN for 'Maximum Path Length Limitation' for
				// more.
				err := os.RemoveAll(`\\?\` + filepath.Join(cwd, assets.ServiceBroker, "dir_name"))
				Expect(err).NotTo(HaveOccurred())
			} else {
				err := os.RemoveAll(filepath.Join(cwd, assets.ServiceBroker, "dir_name"))
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("is successful", func() {
			cf.AsUser(context.RegularUserContext(), setupTimeout, func() {
				space := context.RegularUserContext().Space
				org := context.RegularUserContext().Org

				target := cf.Cf("target", "-o", org, "-s", space).Wait(targetTimeout)
				Expect(target.ExitCode()).To(Equal(0))

				appName := generator.RandomName()
				session := cf.Cf("push", appName, "-p", assets.ServiceBroker).Wait(pushTimeout)
				Expect(session).To(gexec.Exit(0))
			})
		})
	})
})
