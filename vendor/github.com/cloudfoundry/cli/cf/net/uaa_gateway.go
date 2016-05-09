package net

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
)

type uaaErrorResponse struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

var uaaErrorHandler = func(statusCode int, body []byte) error {
	response := uaaErrorResponse{}
	json.Unmarshal(body, &response)

	if response.Code == "invalid_token" {
		return errors.NewInvalidTokenError(response.Description)
	}

	return errors.NewHTTPError(statusCode, response.Code, response.Description)
}

func NewUAAGateway(config coreconfig.Reader, ui terminal.UI, logger trace.Printer) Gateway {
	return newGateway(uaaErrorHandler, config, ui, logger)
}
