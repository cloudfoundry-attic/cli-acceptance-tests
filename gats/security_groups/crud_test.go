package security_groups_test

import (
	"github.com/nu7hatch/gouuid"
	"github.com/pivotal-cf-experimental/GATS/helpers"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var assertionTimeout = 10.0

var _ = Describe("CF security group commands", func() {

	var securityGroupName, orgName, spaceName string

	BeforeEach(func() {
		AsUser(context.AdminUserContext(), func() {
			bytes, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())

			securityGroupName = bytes.String()
			orgName = "org-" + bytes.String()
			spaceName = "space-" + bytes.String()
			emptySecRulesPath := helpers.NewAssets().EmptySecurityRules

			Eventually(Cf("create-security-group", securityGroupName, emptySecRulesPath), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("create-org", orgName), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("create-space", spaceName, "-o", orgName), assertionTimeout).Should(Say("OK"))
		})
	})

	AfterEach(func() {
		AsUser(context.AdminUserContext(), func() {
			// Must target org in order to delete security-group.
			Eventually(Cf("target", "-o", orgName, "-s", spaceName), assertionTimeout).ShouldNot(Say("FAILED"))

			Eventually(Cf("delete-security-group", securityGroupName, "-f"), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("security-group", securityGroupName), assertionTimeout).Should(Say("not found"))
			Eventually(Cf("delete-space", spaceName, "-f"), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("delete-org", orgName, "-f"), assertionTimeout).Should(Say("OK"))
		})
	})

	It("has a workflow for CRUD", func() {
		AsUser(context.AdminUserContext(), func() {
			Eventually(Cf("security-group", securityGroupName), assertionTimeout).Should(Say("Rules"))

			Eventually(Cf(
				"update-security-group",
				securityGroupName,
				helpers.NewAssets().SecurityRules,
			), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("security-group", securityGroupName), assertionTimeout).Should(Say("8.8.8.8"))

			Eventually(Cf("security-groups"), assertionTimeout).Should(Say(securityGroupName))
		})
	})

	It("has a workflow for default staging security groups", func() {
		AsUser(context.AdminUserContext(), func() {
			Eventually(Cf("staging-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))

			Eventually(Cf("bind-staging-security-group", securityGroupName), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("staging-security-groups"), assertionTimeout).Should(Say(securityGroupName))

			Eventually(Cf("unbind-staging-security-group"), assertionTimeout).ShouldNot(Say("OK"))
			Eventually(Cf("staging-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))
		})
	})

	It("has a workflow for default running security groups", func() {
		AsUser(context.AdminUserContext(), func() {
			Eventually(Cf("running-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))

			Eventually(Cf("bind-running-security-group", securityGroupName), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("running-security-groups"), assertionTimeout).Should(Say(securityGroupName))

			Eventually(Cf("unbind-running-security-group"), assertionTimeout).ShouldNot(Say("OK"))
			Eventually(Cf("running-security-groups"), assertionTimeout).ShouldNot(Say(securityGroupName))
		})
	})

	It("has a workflow for binding and unbinding security groups", func() {
		AsUser(context.AdminUserContext(), func() {
			Eventually(Cf("bind-security-group", securityGroupName, orgName, spaceName), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("security-group", securityGroupName), assertionTimeout).Should(Say(spaceName))

			Eventually(Cf("unbind-security-group", securityGroupName, orgName, spaceName), assertionTimeout).Should(Say("OK"))
			Eventually(Cf("security-group", securityGroupName), assertionTimeout).ShouldNot(Say(spaceName))
		})
	})
})
