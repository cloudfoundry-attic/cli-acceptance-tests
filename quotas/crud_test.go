package quotas_test

import(
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/cf"
)

var _ = Describe("CF Quota commands", func() {
	It("can Create, Read, Update, and Delete quotas", func() {
		AsUser(context.AdminUserContext(), func() {
			Eventually(Cf("create-quota",
				"quota-name-goes-here",
				"-m", "512M",
			), 5.0).Should(Exit(0))

			Eventually(Cf("quota", "quota-name-goes-here"), 5.0).Should(Say("512M"))

			quotaOutput := Cf("quotas")
			Eventually(quotaOutput, 5).Should(Say("quota-name-goes-here"))

			Eventually(Cf("update-quota",
				"quota-name-goes-here",
				"-m", "513M",
			), 5).Should(Exit(0))

			Eventually(Cf("quotas")).Should(Say("513M"))

			Eventually(Cf("delete-quota",
				"quota-name-goes-here",
				"-f,",
			)).Should(Exit(0))

			Eventually(Cf("quotas")).ShouldNot(Say("quota-name-goes-here"))
		})
	})
})
