package coreconfig

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/models"
)

type AuthPromptType string

const (
	AuthPromptTypeText     AuthPromptType = "TEXT"
	AuthPromptTypePassword AuthPromptType = "PASSWORD"
)

type AuthPrompt struct {
	Type        AuthPromptType
	DisplayName string
}

type Data struct {
	ConfigVersion            int
	Target                   string
	APIVersion               string
	AuthorizationEndpoint    string
	LoggregatorEndPoint      string
	DopplerEndPoint          string
	UaaEndpoint              string
	RoutingAPIEndpoint       string
	AccessToken              string
	SSHOAuthClient           string
	RefreshToken             string
	OrganizationFields       models.OrganizationFields
	SpaceFields              models.SpaceFields
	SSLDisabled              bool
	AsyncTimeout             uint
	Trace                    string
	ColorEnabled             string
	Locale                   string
	PluginRepos              []models.PluginRepo
	MinCLIVersion            string
	MinRecommendedCLIVersion string
}

func NewData() (data *Data) {
	data = new(Data)
	return
}

func (d *Data) JSONMarshalV3() (output []byte, err error) {
	d.ConfigVersion = 3
	return json.MarshalIndent(d, "", "  ")
}

func (d *Data) JSONUnmarshalV3(input []byte) (err error) {
	err = json.Unmarshal(input, d)
	if err != nil {
		return
	}

	if d.ConfigVersion != 3 {
		*d = Data{}
		return
	}

	return
}
