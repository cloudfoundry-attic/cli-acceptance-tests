package actors

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type RouteActor struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
}

func NewRouteActor(ui terminal.UI, routeRepo api.RouteRepository) RouteActor {
	return RouteActor{ui: ui, routeRepo: routeRepo}
}

func (routeActor RouteActor) CreateRandomTCPRoute(domain models.DomainFields) models.Route {
	routeActor.ui.Say(T("Creating random route for {{.Domain}}", map[string]interface{}{
		"Domain": terminal.EntityNameColor(domain.Name),
	}) + "...")

	route, err := routeActor.routeRepo.Create("", domain, "", true)
	if err != nil {
		routeActor.ui.Failed(err.Error())
		return models.Route{}
	}

	return route
}

func (routeActor RouteActor) FindOrCreateRoute(hostname string, domain models.DomainFields, path string, useRandomPort bool) (route models.Route) {
	var port int
	route, apiErr := routeActor.routeRepo.Find(hostname, domain, path, port)

	switch apiErr.(type) {
	case nil:
		routeActor.ui.Say(T("Using route {{.RouteURL}}", map[string]interface{}{"RouteURL": terminal.EntityNameColor(route.URL())}))
	case *errors.ModelNotFoundError:
		if useRandomPort {
			route = routeActor.CreateRandomTCPRoute(domain)
		} else {
			routeActor.ui.Say(T("Creating route {{.Hostname}}...", map[string]interface{}{"Hostname": terminal.EntityNameColor(domain.URLForHostAndPath(hostname, path, port))}))

			route, apiErr = routeActor.routeRepo.Create(hostname, domain, path, useRandomPort)
			if apiErr != nil {
				routeActor.ui.Failed(apiErr.Error())
			}
		}

		routeActor.ui.Ok()
		routeActor.ui.Say("")
	default:
		routeActor.ui.Failed(apiErr.Error())
	}

	return
}

func (routeActor RouteActor) BindRoute(app models.Application, route models.Route) {
	if !app.HasRoute(route) {
		routeActor.ui.Say(T("Binding {{.URL}} to {{.AppName}}...", map[string]interface{}{"URL": terminal.EntityNameColor(route.URL()), "AppName": terminal.EntityNameColor(app.Name)}))

		apiErr := routeActor.routeRepo.Bind(route.GUID, app.GUID)
		switch apiErr := apiErr.(type) {
		case nil:
			routeActor.ui.Ok()
			routeActor.ui.Say("")
			return
		case errors.HTTPError:
			if apiErr.ErrorCode() == errors.InvalidRelation {
				routeActor.ui.Failed(T("The route {{.URL}} is already in use.\nTIP: Change the hostname with -n HOSTNAME or use --random-route to generate a new route and then push again.", map[string]interface{}{"URL": route.URL()}))
			}
		}
		routeActor.ui.Failed(apiErr.Error())
	}
}

func (routeActor RouteActor) UnbindAll(app models.Application) {
	for _, route := range app.Routes {
		routeActor.ui.Say(T("Removing route {{.URL}}...", map[string]interface{}{"URL": terminal.EntityNameColor(route.URL())}))
		routeActor.routeRepo.Unbind(route.GUID, app.GUID)
	}
}
