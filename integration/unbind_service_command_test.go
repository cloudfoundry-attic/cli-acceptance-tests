package integration

import (
	"os"

	. "code.cloudfoundry.org/cli-acceptance-tests/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-service command", func() {
	var (
		org             string
		space           string
		service         string
		servicePlan     string
		serviceInstance string
		appName         string
		broker          ServiceBroker
	)

	BeforeEach(func() {
		Skip("until #129631341")
		org = PrefixedRandomName("ORG")
		space = PrefixedRandomName("SPACE")
		service = PrefixedRandomName("SERVICE")
		servicePlan = PrefixedRandomName("SERVICE-PLAN")
		serviceInstance = "service-instance"
		appName = PrefixedRandomName("app")

		setupCF(org, space)
	})

	AfterEach(func() {
		Eventually(CF("delete-org", "-f", org), CFLongTimeout).Should(Exit(0))
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				unsetAPI()
			})

			AfterEach(func() {
				setAPI()
				loginCF()
			})

			It("fails with no API endpoint set message", func() {
				Eventually(CF("unbind-service", "some-app", "some-service")).Should(SatisfyAll(
					Exit(1),
					Say("FAILED\nNo API endpoint set. Use 'cf login' or 'cf api' to target an endpoint.")),
				)
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				logoutCF()
			})

			AfterEach(func() {
				loginCF()
			})

			It("fails with not logged in message", func() {
				Eventually(CF("unbind-service", "some-app", "some-service")).Should(SatisfyAll(
					Exit(1),
					Say("FAILED\nNot logged in. Use 'cf login' to log in.")),
				)
			})
		})

		Context("when there no space set", func() {
			BeforeEach(func() {
				logoutCF()
				loginCF()
			})

			It("fails with no targeted space error message", func() {
				Eventually(CF("unbind-service", "some-app", "some-service")).Should(SatisfyAll(
					Exit(1),
					Say("FAILED\nServer error, status code: 404, error code: 40004, message: The app space could not be found: apps")),
				)
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		BeforeEach(func() {
			broker = NewServiceBroker(PrefixedRandomName("SERVICE-BROKER"), NewAssets().ServiceBroker, "bosh-lite.com", service, servicePlan)
			broker.Push()
			broker.Configure()
			broker.Create()

			Eventually(CF("enable-service-access", service)).Should(Exit(0))
		})

		AfterEach(func() {
			broker.Destroy()
		})

		Context("when the service is bound to an app", func() {
			BeforeEach(func() {
				Eventually(CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
				Eventually(CF("push", appName, "--no-start", "-p", os.TempDir()), CFLongTimeout).Should(Exit(0))
				Eventually(CF("bind-service", appName, serviceInstance)).Should(Exit(0))
			})

			It("unbinds the service", func() {
				Eventually(CF("services")).Should(SatisfyAll(
					Exit(0),
					Say("%s.*%s", serviceInstance, appName)),
				)
				Eventually(CF("unbind-service", appName, serviceInstance), CFLongTimeout).Should(Exit(0))
				Eventually(CF("services")).Should(SatisfyAll(
					Exit(0),
					Not(Say("%s.*%s", serviceInstance, appName)),
				))
			})
		})

		Context("when the service is not bound to an app", func() {
			BeforeEach(func() {
				Eventually(CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
				Eventually(CF("push", appName, "--no-start", "-p", os.TempDir()), CFLongTimeout).Should(Exit(0))
			})

			It("fails to unbind the service", func() {
				Eventually(CF("unbind-service", appName, serviceInstance), CFLongTimeout).Should(SatisfyAll(
					Exit(0),
					Say("Binding between service-instance and %s did not exist", appName),
				))
			})
		})

		Context("when the service does not exist", func() {
			BeforeEach(func() {
				Eventually(CF("push", appName, "--no-start", "-p", os.TempDir()), CFLongTimeout).Should(Exit(0))
			})

			It("fails to unbind the service", func() {
				Eventually(CF("unbind-service", appName, serviceInstance), CFLongTimeout).Should(SatisfyAll(
					Exit(1),
					Say("Service instance %s not found", serviceInstance),
				))
			})
		})

		Context("when the app does not exist", func() {
			BeforeEach(func() {
				Eventually(CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
			})

			It("fails to unbind the service", func() {
				Eventually(CF("unbind-service", appName, serviceInstance), CFLongTimeout).Should(SatisfyAll(
					Exit(1),
					Say("App %s not found", appName),
				))
			})
		})
	})
})
