package cmd

import (
	"fmt"
	"strconv"

	"github.com/fsuhrau/automationhub/cli/api"
)

// target names the app of a project on a hub.
type target struct {
	apiURL    string
	projectID string
	appID     uint
}

// client is the API client of the target's app.
func (t target) client(token string) *api.Client {
	return api.NewClient(api.Config{URL: t.apiURL, Token: token, ProjectID: t.projectID, AppID: t.appID})
}

// parseTarget reads "<url> <projectID> <appID>" from the start of args.
func parseTarget(args []string) (target, error) {
	if len(args) < 3 {
		return target{}, fmt.Errorf("missing parameter: <url> <projectID> <appID>")
	}
	appID, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return target{}, fmt.Errorf("invalid appID %q: %w", args[2], err)
	}
	return target{apiURL: args[0], projectID: args[1], appID: uint(appID)}, nil
}
