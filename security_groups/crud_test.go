package security_groups_test

import (
	"github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var assertionTimeout = 10.0

var _ = PDescribe("CF security group commands", func() {

	var securityGroupName string

	BeforeEach(func() {
		AsUser(context.AdminUserContext(), func() {
			quotaBytes, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			securityGroupName = quotaBytes.String()

			Eventually(Cf("create-security-group", securityGroupName), assertionTimeout).Should(Say("OK"))
		})
	})

	AfterEach(func() {
		AsUser(context.AdminUserContext(), func() {
			Eventually(Cf("delete-security-group", securityGroupName), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("security-group", securityGroupName), assertionTimeout).Should(Say("not found"))
		})
	})

	It("has a workflow for CRUD", func() {
		AsUser(context.AdminUserContext(), func() {
			Eventually(Cf("security-group", securityGroupName), assertionTimeout).Should(Say("Rules"))

			Eventually(Cf(
				"update-security-group",
				securityGroupName,
				"--rules",
				`[{"protocol": "tcp", "port": "8081", "destination": "8.8.8.8"}]`,
			), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("security-group", securityGroupName), assertionTimeout).Should(Say("8.8.8.8"))

			Eventually(Cf("security-groups"), assertionTimeout).Should(Say(securityGroupName))
		})
	})

	It("has a workflow for default staging security groups", func() {
		Eventually(Cf("default-staging-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))

		Eventually(Cf("add-default-staging-security-group", securityGroupName), assertionTimeout).Should(Say("OK"))
		Eventually(Cf("default-staging-security-groups"), assertionTimeout).Should(Say(securityGroupName))

		Eventually(Cf("remove-default-staging-security-group"), assertionTimeout).ShouldNot(Say("OK"))
		Eventually(Cf("default-staging-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))
	})

	It("has a workflow for default running security groups", func() {
		Eventually(Cf("default-running-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))

		Eventually(Cf("add-default-running-security-group", securityGroupName), assertionTimeout).Should(Say("OK"))
		Eventually(Cf("default-running-security-groups"), assertionTimeout).Should(Say(securityGroupName))

		Eventually(Cf("remove-default-running-security-group"), assertionTimeout).ShouldNot(Say("OK"))
		Eventually(Cf("default-running-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))
	})

})
