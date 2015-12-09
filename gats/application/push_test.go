package application_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

	Context("when pushing an app with >248 character paths, which are .cfignored", func() {
		var (
			longPath     string
			cfIgnorePath string
		)

		BeforeEach(func() {
			longDirName := strings.Repeat("i", 247)
			longPath = filepath.Join(assets.ServiceBroker, longDirName)

			if runtime.GOOS == "windows" {
				cwd, err := os.Getwd()
				Expect(err).NotTo(HaveOccurred())

				// `\\?\` is used to skip Windows' file name processor, which imposes
				// length limits. Search MSDN for 'Maximum Path Length Limitation' for
				// more.
				err := os.MkdirAll(`\\?\`+filepath.Join(cwd, longPath), os.ModeDir|os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			} else {
				err := os.MkdirAll(longPath, os.ModeDir|os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			}

			cfIgnorePath = filepath.Join(assets.ServiceBroker, ".cfignore")
			cfIgnoreContents := []byte(longDirName + "\n")
			err = ioutil.WriteFile(cfIgnorePath, cfIgnoreContents, 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(cfIgnorePath)
			Expect(err).NotTo(HaveOccurred())
			err = os.RemoveAll(longPath)
			Expect(err).NotTo(HaveOccurred())
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
