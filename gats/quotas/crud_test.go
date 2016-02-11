package quotas_test

import (
	"time"

	"github.com/nu7hatch/gouuid"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var (
	assertionTimeout      = 10.0
	asyncAssertionTimeout = 15.0
)

var _ = Describe("CF Quota commands", func() {
	It("can Create, Read, Update, and Delete quotas", func() {
		AsUser(context.AdminUserContext(), 100*time.Second, func() {
			quotaBytes, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			quotaName := quotaBytes.String()

			Eventually(Cf("create-quota",
				quotaName,
				"-m", "512M", "-i", "1G",
			), assertionTimeout).Should(Say("OK"))

			Eventually(Cf("quota", quotaName), assertionTimeout).Should(Say("512M"))
			Eventually(Cf("quota", quotaName), assertionTimeout).Should(Say("1G"))

			quotaOutput := Cf("quotas")
			Eventually(quotaOutput, assertionTimeout).Should(Say(quotaName))

			Eventually(Cf("update-quota",
				quotaName,
				"-m", "513M",
			), assertionTimeout).Should(Say("OK"))

			Eventually(Cf("quotas"), assertionTimeout).Should(Say("513M"))

			Eventually(
				Cf("delete-quota", quotaName, "-f"),
				asyncAssertionTimeout,
			).Should(Say("OK"))

			Eventually(Cf("quotas"), assertionTimeout).ShouldNot(Say(quotaName))
		})
	})
})
