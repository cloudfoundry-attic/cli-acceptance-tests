package security_groups_test

import (
	"github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var assertionTimeout = 10.0

var _ = PDescribe("CF security group commands", func() {

	var securityGroupName, orgName, spaceName string

	BeforeEach(func() {
		AsUser(context.AdminUserContext(), func() {
			bytes, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			securityGroupName = bytes.String()
			orgName = "org-" + bytes.String()
			spaceName = "space-" + bytes.String()

			Eventually(Cf("create-security-group", securityGroupName), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("create-org", orgName), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("create-space", spaceName), assertionTimeout).Should(Say("OK"))
		})
	})

	AfterEach(func() {
		AsUser(context.AdminUserContext(), func() {
			Eventually(Cf("delete-security-group", securityGroupName), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("security-group", securityGroupName), assertionTimeout).Should(Say("not found"))
			Eventually(Cf("delete-space", spaceName), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("delete-org", orgName), assertionTimeout).Should(Say("OK"))
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
		Eventually(Cf("staging-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))

		Eventually(Cf("add-staging-security-group", securityGroupName), assertionTimeout).Should(Say("OK"))
		Eventually(Cf("staging-security-groups"), assertionTimeout).Should(Say(securityGroupName))

		Eventually(Cf("remove-staging-security-group"), assertionTimeout).ShouldNot(Say("OK"))
		Eventually(Cf("staging-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))
	})

	It("has a workflow for default running security groups", func() {
		Eventually(Cf("running-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))

		Eventually(Cf("add-running-security-group", securityGroupName), assertionTimeout).Should(Say("OK"))
		Eventually(Cf("running-security-groups"), assertionTimeout).Should(Say(securityGroupName))

		Eventually(Cf("remove-running-security-group"), assertionTimeout).ShouldNot(Say("OK"))
		Eventually(Cf("running-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))
	})

	It("has a workflow for assigning and unassigning security groups", func() {
		Eventually(Cf("assign-security-group", securityGroupName, orgName, spaceName), assertionTimeout).Should(Say("OK"))
		Eventually(Cf("security-group", securityGroupName), assertionTimeout).Should(Say(spaceName))

		Eventually(Cf("unassign-security-group", securityGroupName, orgName, spaceName), assertionTimeout).Should(Say("OK"))
		Eventually(Cf("security-group", securityGroupName), assertionTimeout).ShouldNot(Say(spaceName))
	})

})
