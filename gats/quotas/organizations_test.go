package quotas_test

import (
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("CF Organizations Quota", func() {
	var (
		orgName   string
		quotaName string
	)

	BeforeEach(func() {
		AsUser(context.AdminUserContext(), 100*time.Second, func() {
			orgName = generateUniqueName()
			quotaName = generateUniqueName()

			Eventually(Cf("create-org",
				orgName,
			), assertionTimeout).Should(Say("OK"))

			Eventually(Cf("create-quota",
				quotaName, "-m", "100M",
			), assertionTimeout).Should(Say("OK"))

			Eventually(Cf("set-quota",
				orgName, quotaName,
			), assertionTimeout).Should(Say("OK"))
		})
	})

	AfterEach(func() {
		AsUser(context.AdminUserContext(), 100*time.Second, func() {
			Eventually(Cf("delete-org", orgName, "-f"), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("delete-quota", quotaName, "-f"), assertionTimeout).Should(Say("OK"))
		})
	})

	It("can get accurate quotas back", func() {
		AsUser(context.AdminUserContext(), 100*time.Second, func() {
			Eventually(Cf("org", orgName), assertionTimeout).Should(Say("100M"))
		})
	})
})

func generateUniqueName() string {
	uuidBytes, err := uuid.NewV4()
	Expect(err).ToNot(HaveOccurred())
	return uuidBytes.String()
}
