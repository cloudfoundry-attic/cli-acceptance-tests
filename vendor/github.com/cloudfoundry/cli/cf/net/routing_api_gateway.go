package net

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
)

type errorResponse struct {
	Name    string
	Message string
}

func errorHandler(statusCode int, body []byte) error {
	response := errorResponse{}
	err := json.Unmarshal(body, &response)
	if err != nil {
		return errors.NewHTTPError(http.StatusInternalServerError, "", "")
	}

	return errors.NewHTTPError(statusCode, response.Name, response.Message)
}

func NewRoutingAPIGateway(config coreconfig.Reader, clock func() time.Time, ui terminal.UI, logger trace.Printer) Gateway {
	gateway := newGateway(errorHandler, config, ui, logger)
	gateway.Clock = clock
	gateway.PollingEnabled = true
	return gateway
}
