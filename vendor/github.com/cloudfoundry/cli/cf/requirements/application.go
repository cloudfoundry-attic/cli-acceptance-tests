package requirements

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/models"
)

//go:generate counterfeiter . ApplicationRequirement

type ApplicationRequirement interface {
	Requirement
	GetApplication() models.Application
}

type applicationAPIRequirement struct {
	name        string
	appRepo     applications.ApplicationRepository
	application models.Application
}

func NewApplicationRequirement(name string, aR applications.ApplicationRepository) *applicationAPIRequirement {
	req := &applicationAPIRequirement{}
	req.name = name
	req.appRepo = aR
	return req
}

func (req *applicationAPIRequirement) Execute() error {
	var apiErr error
	req.application, apiErr = req.appRepo.Read(req.name)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *applicationAPIRequirement) GetApplication() models.Application {
	return req.application
}
