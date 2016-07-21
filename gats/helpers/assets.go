package helpers

type Assets struct {
	ServiceBroker      string
	SecurityRules      string
	EmptySecurityRules string
	DoraApp            string
}

func NewAssets() Assets {
	return Assets{
		ServiceBroker:      "../assets/service_broker",
		SecurityRules:      "../assets/security_groups/security-rules.json",
		EmptySecurityRules: "../assets/security_groups/empty-security-rules.json",
		DoraApp:            "../assets/dora",
	}
}
