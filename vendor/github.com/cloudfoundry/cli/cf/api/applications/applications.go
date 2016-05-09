package applications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

//go:generate counterfeiter . ApplicationRepository

type ApplicationRepository interface {
	Create(params models.AppParams) (createdApp models.Application, apiErr error)
	GetApp(appGUID string) (models.Application, error)
	Read(name string) (app models.Application, apiErr error)
	ReadFromSpace(name string, spaceGUID string) (app models.Application, apiErr error)
	Update(appGUID string, params models.AppParams) (updatedApp models.Application, apiErr error)
	Delete(appGUID string) (apiErr error)
	ReadEnv(guid string) (*models.Environment, error)
	CreateRestageRequest(guid string) (apiErr error)
}

type CloudControllerApplicationRepository struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

func NewCloudControllerApplicationRepository(config coreconfig.Reader, gateway net.Gateway) (repo CloudControllerApplicationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerApplicationRepository) Create(params models.AppParams) (models.Application, error) {
	appResource := resources.NewApplicationEntityFromAppParams(params)
	data, err := json.Marshal(appResource)
	if err != nil {
		return models.Application{}, fmt.Errorf("%s: %s", T("Failed to marshal JSON"), err.Error())
	}

	resource := new(resources.ApplicationResource)
	err = repo.gateway.CreateResource(repo.config.APIEndpoint(), "/v2/apps", bytes.NewReader(data), resource)
	if err != nil {
		return models.Application{}, err
	}

	return resource.ToModel(), nil
}

func (repo CloudControllerApplicationRepository) GetApp(appGUID string) (app models.Application, apiErr error) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.APIEndpoint(), appGUID)
	appResources := new(resources.ApplicationResource)

	apiErr = repo.gateway.GetResource(path, appResources)
	if apiErr != nil {
		return
	}

	app = appResources.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Read(name string) (app models.Application, apiErr error) {
	return repo.ReadFromSpace(name, repo.config.SpaceFields().GUID)
}

func (repo CloudControllerApplicationRepository) ReadFromSpace(name string, spaceGUID string) (app models.Application, apiErr error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=%s&inline-relations-depth=1", repo.config.APIEndpoint(), spaceGUID, url.QueryEscape("name:"+name))
	appResources := new(resources.PaginatedApplicationResources)
	apiErr = repo.gateway.GetResource(path, appResources)
	if apiErr != nil {
		return
	}

	if len(appResources.Resources) == 0 {
		apiErr = errors.NewModelNotFoundError("App", name)
		return
	}

	res := appResources.Resources[0]
	app = res.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Update(appGUID string, params models.AppParams) (updatedApp models.Application, apiErr error) {
	appResource := resources.NewApplicationEntityFromAppParams(params)
	data, err := json.Marshal(appResource)
	if err != nil {
		return models.Application{}, fmt.Errorf("%s: %s", T("Failed to marshal JSON"), err.Error())
	}

	path := fmt.Sprintf("/v2/apps/%s?inline-relations-depth=1", appGUID)
	resource := new(resources.ApplicationResource)
	apiErr = repo.gateway.UpdateResource(repo.config.APIEndpoint(), path, bytes.NewReader(data), resource)
	if apiErr != nil {
		return
	}

	updatedApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Delete(appGUID string) (apiErr error) {
	path := fmt.Sprintf("/v2/apps/%s?recursive=true", appGUID)
	return repo.gateway.DeleteResource(repo.config.APIEndpoint(), path)
}

func (repo CloudControllerApplicationRepository) ReadEnv(guid string) (*models.Environment, error) {
	var (
		err error
	)

	path := fmt.Sprintf("%s/v2/apps/%s/env", repo.config.APIEndpoint(), guid)
	appResource := models.NewEnvironment()

	err = repo.gateway.GetResource(path, appResource)
	if err != nil {
		return &models.Environment{}, err
	}

	return appResource, err
}

func (repo CloudControllerApplicationRepository) CreateRestageRequest(guid string) error {
	path := fmt.Sprintf("/v2/apps/%s/restage", guid)
	return repo.gateway.CreateResource(repo.config.APIEndpoint(), path, strings.NewReader(""), nil)
}
