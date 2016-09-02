package integration

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Help Flag", func() {
	DescribeTable("displays the master help text",
		func(setup func() *exec.Cmd) {
			cmd := setup()
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("VERSION:"))
			Eventually(session).Should(Say("GETTING STARTED:"))
			Eventually(session).Should(Say("ENVIRONMENT VARIABLES:"))
			Eventually(session).Should(Say("GLOBAL OPTIONS:"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when cf is run without providing a command or a flag", func() *exec.Cmd {
			Skip("Ask dies what should happen in this case")
			return exec.Command("cf", "-a")
		}),

		Entry("when cf is run with -h flag alone", func() *exec.Cmd {
			return exec.Command("cf", "-h", "-a")
		}),

		Entry("when cf is run with --help flag alone", func() *exec.Cmd {
			return exec.Command("cf", "--help", "-a")
		}),
	)
})
