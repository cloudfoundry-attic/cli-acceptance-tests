package v3_helpers

import (
	"github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

type Entity struct {
	AppName       string `json:"app_name"`
	AppGuid       string `json:"app_guid"`
	State         string `json:"state"`
	BuildpackName string `json:"buildpack_name"`
	BuildpackGuid string `json:"buildpack_guid"`
	ParentAppName string `json:"parent_app_name"`
	ParentAppGuid string `json:"parent_app_guid"`
	ProcessType   string `json:"process_type"`
	TaskGuid      string `json:"task_guid"`
}
type AppUsageEvent struct {
	Entity `json:"entity"`
}

type AppUsageEvents struct {
	Resources []AppUsageEvent `struct:"resources"`
}

func UsageEventsInclude(events []AppUsageEvent, event AppUsageEvent) bool {
	found := false
	for _, e := range events {
		found = event.Entity.ParentAppName == e.Entity.ParentAppName &&
			event.Entity.ParentAppGuid == e.Entity.ParentAppGuid &&
			event.Entity.ProcessType == e.Entity.ProcessType &&
			event.Entity.State == e.Entity.State &&
			event.Entity.AppGuid == e.Entity.AppGuid &&
			event.Entity.TaskGuid == e.Entity.TaskGuid
		if found {
			break
		}
	}
	return found
}

func LastPageUsageEvents(context helpers.SuiteContext) []AppUsageEvent {
	var response AppUsageEvents

	cf.AsUser(context.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		cf.ApiRequest("GET", "/v2/app_usage_events?order-direction=desc&page=1", &response, DEFAULT_TIMEOUT)
	})

	return response.Resources
}
