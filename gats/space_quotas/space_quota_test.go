package space_quotas_test

import (
	"fmt"

	CATS_helper "github.com/cloudfoundry/cf-acceptance-tests/helpers"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("CF space quota commands", func() {
	const (
		assertionTimeout = 10.0
	)

	var (
		context *CATS_helper.ConfiguredContext
		env     *CATS_helper.Environment
		orgName string
	)

	config := CATS_helper.LoadConfig()

	BeforeEach(func() {
		context = CATS_helper.NewContext(config)
		env = CATS_helper.NewEnvironment(context)
		orgName = context.RegularUserContext().Org

		env.Setup()
	})

	AfterEach(func() {
		env.Teardown()
	})

	Describe("space quotas", func() {
		It("can create, read, update, assign to a space, remove a space and delete a space quota", func() {
			AsUser(context.AdminUserContext(), func() {
				target := Cf("target", "-o", orgName).Wait(assertionTimeout)
				Expect(target).To(Exit(0))

				spaceQuota := Cf("create-space-quota", "foo").Wait(assertionTimeout)
				Expect(spaceQuota).To(Exit(0))
				output := spaceQuota.Out.Contents()
				Expect(output).To(ContainSubstring(fmt.Sprintf("Creating space quota foo for org %s", orgName)))

				spaceQuota = Cf("space-quotas").Wait(assertionTimeout)
				Expect(spaceQuota).To(Exit(0))
				output = spaceQuota.Out.Contents()
				Expect(output).To(ContainSubstring("Getting space quotas as"))
				Expect(output).To(ContainSubstring("foo"))

				spaceQuota = Cf("update-space-quota", "foo", "-i", "-1").Wait(assertionTimeout)
				Expect(spaceQuota).To(Exit(0))
				output = spaceQuota.Out.Contents()
				Expect(output).To(ContainSubstring("Updating space quota foo"))

				spaceQuota = Cf("space-quota", "foo").Wait(assertionTimeout)
				Expect(spaceQuota).To(Exit(0))
				output = spaceQuota.Out.Contents()
				Expect(output).To(ContainSubstring("instance memory limit   -1"))

				spaceName := context.RegularUserContext().Space
				spaceQuota = Cf("set-space-quota", spaceName, "foo").Wait(assertionTimeout)
				Expect(spaceQuota).To(Exit(0))
				output = spaceQuota.Out.Contents()
				Expect(output).To(ContainSubstring(fmt.Sprintf("Assigning space quota foo to space %s", spaceName)))

				spaceQuota = Cf("unset-space-quota", spaceName, "foo").Wait(assertionTimeout)
				Expect(spaceQuota).To(Exit(0))
				output = spaceQuota.Out.Contents()
				Expect(output).To(ContainSubstring(fmt.Sprintf("Unassigning space quota foo from space %s", spaceName)))

				spaceQuota = Cf("delete-space-quota", "foo", "-f").Wait(assertionTimeout)
				Expect(spaceQuota).To(Exit(0))
				output = spaceQuota.Out.Contents()
				Expect(output).To(ContainSubstring("Deleting space quota foo"))
			})
		})
	})
})
