package integration

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/utils/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("API Command", func() {
	BeforeEach(func() {
		Skip("Until #126256625 has been completed")
	})

	Context("no arguments", func() {
		Context("when the api is set", func() {
			Context("when the user is not logged in", func() {
				It("outputs the current api", func() {
					command := exec.Command("cf", "api")
					session, err := Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session.Out).Should(Say("API endpoint:\\s+https://%s", getAPI()))
					Eventually(session.Out).Should(Say("API version: \\d+\\.\\d+\\.\\d+"))
					Eventually(session.Out).Should(Say("^User:$"))
					Eventually(session.Out).Should(Say("^Org:$"))
					Eventually(session.Out).Should(Say("^Space:$"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the user is logged in", func() {
				var target, apiVersion, user, org, space string

				BeforeEach(func() {
					target = "api.fake.com"
					apiVersion = "2.59.0"
					user = "faceman@fake.com"
					org = "the-org"
					space = "the-space"

					userConfig := config.Config{
						ConfigFile: config.CFConfig{
							Target:     target,
							APIVersion: apiVersion,
							TargetedOrganization: config.Organization{
								Name: org,
							},
							TargetedSpace: config.Space{
								Name: space,
							},
						},
					}
					err := config.WriteConfig(&userConfig)
					Expect(err).ToNot(HaveOccurred())
				})

				It("outputs the user's target information", func() {
					command := exec.Command("cf", "api")
					session, err := Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session.Out).Should(Say("API endpoint:\\s+https://%s", target))
					Eventually(session.Out).Should(Say("API version: %s", apiVersion))
					// Eventually(session.Out).Should(Say("User:", user))
					Eventually(session.Out).Should(Say("Org:", org))
					Eventually(session.Out).Should(Say("Space:", space))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the api is not set", func() {
			BeforeEach(func() {
				os.RemoveAll(filepath.Join(homeDir, ".cf"))
			})

			It("outputs that nothing is set", func() {
				command := exec.Command("cf", "api")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session.Out).Should(Say("No api endpoint set. Use 'cf api' to set an endpoint"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when Skip SSL Validation is required", func() {
		Context("api has SSL", func() {
			BeforeEach(func() {
				if skipSSLValidation == "" {
					Skip("SKIP_SSL_VALIDATION is not enabled")
				}
			})

			It("warns about skip SSL", func() {
				command := exec.Command("cf", "api", getAPI())
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session.Out).Should(Say("Setting api endpoint to %s...", getAPI()))
				Eventually(session.Err).Should(Say("Invalid SSL Cert for %s", getAPI()))
				Eventually(session.Err).Should(Say("TIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})

			It("sets the API endpoint", func() {
				command := exec.Command("cf", "api", getAPI(), "--skip-ssl-validation")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session.Out).Should(Say("Setting api endpoint to %s...", getAPI()))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say("API endpoint:\\s+https://%s \\(API version: \\d+\\.\\d+\\.\\d+\\)", getAPI()))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("api does not have SSL", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = ghttp.NewServer()
				serverAPIURL := server.URL()[7:]

				response := `{
					"name":"",
					"build":"",
					"support":"http://support.cloudfoundry.com",
					"version":0,
					"description":"",
					"authorization_endpoint":"https://login.APISERVER",
					"token_endpoint":"https://uaa.APISERVER",
					"min_cli_version":null,
					"min_recommended_cli_version":null,
					"api_version":"2.59.0",
					"app_ssh_endpoint":"ssh.APISERVER",
					"app_ssh_host_key_fingerprint":"a6:d1:08:0b:b0:cb:9b:5f:c4:ba:44:2a:97:26:19:8a",
					"app_ssh_oauth_client":"ssh-proxy",
					"logging_endpoint":"wss://loggregator.APISERVER",
					"doppler_logging_endpoint":"wss://doppler.APISERVER"
				}`
				response = strings.Replace(response, "APISERVER", serverAPIURL, -1)
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/info"),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)
			})

			AfterEach(func() {
				server.Close()
			})

			It("falls back to http and gives a warning", func() {
				command := exec.Command("cf", "api", server.URL(), "--skip-ssl-validation")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session.Out).Should(Say("Setting api endpoint to %s...", server.URL()))
				Eventually(session.Out).Should(Say("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended"))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when skip-ssl-validation is not required", func() {
		BeforeEach(func() {
			if skipSSLValidation != "" {
				Skip("SKIP_SSL_VALIDATION is enabled")
			}
		})

		It("logs in without any warnings", func() {
			command := exec.Command("cf", "api", getAPI())
			session, err := Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session.Out).Should(Say("Setting api endpoint to %s...", getAPI()))
			Consistently(session.Out).ShouldNot(Say("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended"))
			Eventually(session.Out).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})
	})

	It("sets the config file", func() {
		command := exec.Command("cf", "api", getAPI(), skipSSLValidation)
		session, err := Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(Exit(0))

		rawConfig, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
		Expect(err).NotTo(HaveOccurred())

		var configFile config.CFConfig
		err = json.Unmarshal(rawConfig, &configFile)
		Expect(err).NotTo(HaveOccurred())

		Expect(configFile.ConfigVersion).To(Equal(3))
		Expect(configFile.Target).To(Equal("https://" + getAPI()))
		Expect(configFile.APIVersion).To(MatchRegexp("\\d+\\.\\d+\\.\\d+"))
		Expect(configFile.AuthorizationEndpoint).ToNot(BeEmpty())
		Expect(configFile.LoggregatorEndpoint).To(MatchRegexp("^wss://"))
		Expect(configFile.DopplerEndpoint).To(MatchRegexp("^wss://"))
		Expect(configFile.UAAEndpoint).ToNot(BeEmpty())
		Expect(configFile.AccessToken).To(BeEmpty())
		Expect(configFile.RefreshToken).To(BeEmpty())
		Expect(configFile.TargetedOrganization.GUID).To(BeEmpty())
		Expect(configFile.TargetedOrganization.Name).To(BeEmpty())
		Expect(configFile.TargetedSpace.GUID).To(BeEmpty())
		Expect(configFile.TargetedSpace.Name).To(BeEmpty())
		Expect(configFile.TargetedSpace.AllowSSH).To(BeFalse())
	})
})
