package integration

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Help Command", func() {
	DescribeTable("displays the help text for a given command",
		func(setup func() (*exec.Cmd, int)) {
			cmd, exitCode := setup()
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("create-user-provided-service - Make a user-provided service instance available to CF apps"))
			Eventually(session).Should(Say("cf create-user-provided-service SERVICE_INSTANCE \\[-p CREDENTIALS\\] \\[-l SYSLOG_DRAIN_URL\\] \\[-r ROUTE_SERVICE_URL\\]"))
			Eventually(session).Should(Say("-l\\s+URL to which logs for bound applications will be streamed"))
			Eventually(session).Should(Exit(exitCode))
		},

		Entry("when a command is called with the --help flag", func() (*exec.Cmd, int) {
			return exec.Command("cf", "create-user-provided-service", "--help"), 0
		}),

		Entry("when a command is called with the --help flag and command arguments", func() (*exec.Cmd, int) {
			return exec.Command("cf", "create-user-provided-service", "-l", "http://example.com", "--help"), 0
		}),

		Entry("when a command is called with the --help flag and command arguments prior to the command", func() (*exec.Cmd, int) {
			return exec.Command("cf", "-l", "create-user-provided-service", "--help"), 1
		}),

		Entry("when the help command is passed a command name", func() (*exec.Cmd, int) {
			return exec.Command("cf", "help", "create-user-provided-service"), 0
		}),

		Entry("when the --help flag is passed with a command name", func() (*exec.Cmd, int) {
			return exec.Command("cf", "--help", "create-user-provided-service"), 0
		}),

		Entry("when the help command is passed a command alias", func() (*exec.Cmd, int) {
			return exec.Command("cf", "help", "cups"), 0
		}),

		Entry("when the --help flag is passed with a command alias", func() (*exec.Cmd, int) {
			return exec.Command("cf", "--help", "cups"), 0
		}),

		Entry("when the --help flag is passed after a command alias", func() (*exec.Cmd, int) {
			return exec.Command("cf", "cups", "--help"), 0
		}),

		Entry("when an invalid flag is passed", func() (*exec.Cmd, int) {
			return exec.Command("cf", "create-user-provided-service", "--invalid-flag"), 1
		}),

		Entry("when missing required arguments", func() (*exec.Cmd, int) {
			return exec.Command("cf", "create-user-provided-service"), 1
		}),

		Entry("when missing arguments to flags", func() (*exec.Cmd, int) {
			return exec.Command("cf", "create-user-provided-service", "foo", "-l"), 1
		}),
	)
})
