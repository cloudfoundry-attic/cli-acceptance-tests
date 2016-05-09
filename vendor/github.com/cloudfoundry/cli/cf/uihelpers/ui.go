package uihelpers

import (
	"fmt"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

func ColoredAppState(app models.ApplicationFields) string {
	appState := strings.ToLower(app.State)

	if app.RunningInstances == 0 {
		if appState == "stopped" {
			return appState
		}
		return terminal.CrashedColor(appState)
	}

	if app.RunningInstances < app.InstanceCount {
		return terminal.WarningColor(appState)
	}

	return appState
}

func ColoredAppInstances(app models.ApplicationFields) string {
	healthString := fmt.Sprintf("%d/%d", app.RunningInstances, app.InstanceCount)

	if app.RunningInstances < 0 {
		healthString = fmt.Sprintf("?/%d", app.InstanceCount)
	}

	if app.RunningInstances == 0 {
		if strings.ToLower(app.State) == "stopped" {
			return healthString
		}
		return terminal.CrashedColor(healthString)
	}

	if app.RunningInstances < app.InstanceCount {
		return terminal.WarningColor(healthString)
	}

	return healthString
}

func ColoredInstanceState(instance models.AppInstanceFields) (colored string) {
	state := string(instance.State)
	switch state {
	case "started", "running":
		colored = T("running")
	case "stopped":
		colored = terminal.StoppedColor(T("stopped"))
	case "crashed":
		colored = terminal.CrashedColor(T("crashed"))
	case "flapping":
		colored = terminal.CrashedColor(T("crashing"))
	case "down":
		colored = terminal.CrashedColor(T("down"))
	case "starting":
		colored = terminal.AdvisoryColor(T("starting"))
	default:
		colored = terminal.WarningColor(state)
	}

	return
}
