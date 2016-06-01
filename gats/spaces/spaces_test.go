package quotas_test

import (
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("CF Space", func() {
	var (
		orgName   string
		spaceName string
	)

	BeforeEach(func() {
		AsUser(context.AdminUserContext(), 100*time.Second, func() {
			orgName = generateUniqueName()
			spaceName = generateUniqueName()

			Eventually(Cf("create-org",
				orgName,
			), assertionTimeout).Should(Say("OK"))

			Eventually(Cf("target",
				"-o", orgName,
			), assertionTimeout).Should(Say(orgName))

			Eventually(Cf("create-space",
				spaceName,
			), assertionTimeout).Should(Say("OK"))

		})
	})

	AfterEach(func() {
		AsUser(context.AdminUserContext(), 100*time.Second, func() {
			Eventually(Cf("delete-org", orgName, "-f"), assertionTimeout).Should(Say("OK"))
		})
	})

	It("can get accurate org for space", func() {
		AsUser(context.AdminUserContext(), 100*time.Second, func() {
			Eventually(Cf("target",
				"-o", orgName,
			), assertionTimeout).Should(Say(orgName))

			Eventually(Cf("space",
				spaceName,
			), assertionTimeout).Should(Say(orgName))
		})
	})
})
