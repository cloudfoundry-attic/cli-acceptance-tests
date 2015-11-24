package config_test

import (
	"time"

	CATS_helper "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Config", func() {
	var (
		config         CATS_helper.Config
		context        *CATS_helper.ConfiguredContext
		setupTimeout   time.Duration
		commandTimeout time.Duration
	)

	BeforeEach(func() {
		config = CATS_helper.LoadConfig()
		config.UseExistingUser = true
		context = CATS_helper.NewContext(config)
		setupTimeout = 20 * time.Second
		commandTimeout = 10 * time.Second
	})

	It("allows setting locale to de_DE", func() {
		cf.AsUser(context.AdminUserContext(), setupTimeout, func() {
			session := cf.Cf("config", "-locale", "de_DE").Wait(commandTimeout)
			Expect(session).To(Exit(0))
		})
	})

	It("allows setting locale to en_US", func() {
		cf.AsUser(context.AdminUserContext(), setupTimeout, func() {
			session := cf.Cf("config", "-locale", "en_US").Wait(commandTimeout)
			Expect(session).To(Exit(0))
		})
	})

	It("allows setting locale to es_ES", func() {
		cf.AsUser(context.AdminUserContext(), setupTimeout, func() {
			session := cf.Cf("config", "-locale", "es_ES").Wait(commandTimeout)
			Expect(session).To(Exit(0))
		})
	})

	It("allows setting locale to fr_FR", func() {
		cf.AsUser(context.AdminUserContext(), setupTimeout, func() {
			session := cf.Cf("config", "-locale", "fr_FR").Wait(commandTimeout)
			Expect(session).To(Exit(0))
		})
	})

	It("allows setting locale to it_IT", func() {
		cf.AsUser(context.AdminUserContext(), setupTimeout, func() {
			session := cf.Cf("config", "-locale", "it_IT").Wait(commandTimeout)
			Expect(session).To(Exit(0))
		})
	})

	It("allows setting locale to ja_JA", func() {
		cf.AsUser(context.AdminUserContext(), setupTimeout, func() {
			session := cf.Cf("config", "-locale", "ja_JA").Wait(commandTimeout)
			Expect(session).To(Exit(0))
		})
	})

	It("allows setting locale to ko_KR", func() {
		cf.AsUser(context.AdminUserContext(), setupTimeout, func() {
			session := cf.Cf("config", "-locale", "ko_KR").Wait(commandTimeout)
			Expect(session).To(Exit(0))
		})
	})

	It("allows setting locale to pt_BR", func() {
		cf.AsUser(context.AdminUserContext(), setupTimeout, func() {
			session := cf.Cf("config", "-locale", "pt_BR").Wait(commandTimeout)
			Expect(session).To(Exit(0))
		})
	})

	It("allows setting locale to zh_Hans", func() {
		cf.AsUser(context.AdminUserContext(), setupTimeout, func() {
			session := cf.Cf("config", "-locale", "zh_Hans").Wait(commandTimeout)
			Expect(session).To(Exit(0))
		})
	})

	It("allows setting locale to zh_Hant", func() {
		cf.AsUser(context.AdminUserContext(), setupTimeout, func() {
			session := cf.Cf("config", "-locale", "zh_Hant").Wait(commandTimeout)
			Expect(session).To(Exit(0))
		})
	})
})
