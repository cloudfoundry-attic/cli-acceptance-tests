package integration

import (
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Help Command", func() {
	DescribeTable("displays help for common commands",
		func(setup func() *exec.Cmd) {
			cmd := setup()
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(Say("Cloud Foundry command line tool"))
			Eventually(session).Should(Say("\\[global options\\] command \\[arguments...\\] \\[command options\\]"))
			Eventually(session).Should(Say("Before getting started:"))
			Eventually(session).Should(Say("config\\s+login,l\\s+target,t"))
			Eventually(session).Should(Say("Global options:"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when cf is run without providing a command or a flag", func() *exec.Cmd {
			return exec.Command("cf")
		}),

		Entry("when cf help is run", func() *exec.Cmd {
			return exec.Command("cf", "help")
		}),

		Entry("when cf is run with -h flag alone", func() *exec.Cmd {
			return exec.Command("cf", "-h")
		}),

		Entry("when cf is run with --help flag alone", func() *exec.Cmd {
			return exec.Command("cf", "--help")
		}),
	)

	DescribeTable("displays help for all commands",
		func(setup func() *exec.Cmd) {
			cmd := setup()
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("VERSION:"))
			Eventually(session).Should(Say("GETTING STARTED:"))
			Eventually(session).Should(Say("ENVIRONMENT VARIABLES:"))
			Eventually(session).Should(Say("CF_DIAL_TIMEOUT=5\\s+Max wait time to establish a connection, including name resolution, in seconds"))
			Eventually(session).Should(Say("GLOBAL OPTIONS:"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when cf help is run", func() *exec.Cmd {
			return exec.Command("cf", "help", "-a")
		}),

		Entry("when cf is run with -h -a flag", func() *exec.Cmd {
			return exec.Command("cf", "-h", "-a")
		}),

		Entry("when cf is run with --help -a flag", func() *exec.Cmd {
			return exec.Command("cf", "--help", "-a")
		}),
	)

	Context("displays the help text for a given command", func() {
		DescribeTable("displays the help",
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

		Context("when the command uses timeout environment variables", func() {
			DescribeTable("shows the CF_STAGING_TIMEOUT and CF_STARTUP_TIMEOUT environment variables",
				func(setup func() (*exec.Cmd, int)) {
					cmd, exitCode := setup()
					session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session).Should(Say(`
ENVIRONMENT:
   CF_STAGING_TIMEOUT=15        Max wait time for buildpack staging, in minutes
   CF_STARTUP_TIMEOUT=5         Max wait time for app instance startup, in minutes
`))
					Eventually(session).Should(Exit(exitCode))
				},

				Entry("cf push", func() (*exec.Cmd, int) {
					return exec.Command("cf", "h", "push"), 0
				}),

				Entry("cf start", func() (*exec.Cmd, int) {
					return exec.Command("cf", "h", "start"), 0
				}),

				Entry("cf restart", func() (*exec.Cmd, int) {
					return exec.Command("cf", "h", "restart"), 0
				}),

				Entry("cf restage", func() (*exec.Cmd, int) {
					return exec.Command("cf", "h", "restage"), 0
				}),

				Entry("cf copy-source", func() (*exec.Cmd, int) {
					return exec.Command("cf", "h", "copy-source"), 0
				}),
			)
		})
	})

	Context("when the command does not exist", func() {
		DescribeTable("help displays an error message",
			func(command func() *exec.Cmd) {
				session, err := Start(command(), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session.Err).Should(Say("'rock' is not a registered command. See 'cf help'"))
				Eventually(session).Should(Exit(1))
			},

			Entry("passing the --help flag", func() *exec.Cmd {
				return exec.Command("cf", "--help", "rock")
			}),

			Entry("calling the help command directly", func() *exec.Cmd {
				return exec.Command("cf", "help", "rock")
			}),
		)

	})

	Context("when the option does not exist", func() {
		DescribeTable("help display an error message as well as help for common commands",

			func(command func() *exec.Cmd) {
				session, err := Start(command(), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(Exit(1))
				Eventually(session).Should(Say("Before getting started:")) // common help
				Expect(strings.Count(string(session.Err.Contents()), "unknown flag")).To(Equal(1))
			},

			Entry("passing invalid option", func() *exec.Cmd {
				return exec.Command("cf", "-c")
			}),

			Entry("passing -a option", func() *exec.Cmd {
				return exec.Command("cf", "-a")
			}),
		)
	})
})
