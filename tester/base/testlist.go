package base

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	tester_action "github.com/fsuhrau/automationhub/tester/action"
)

// testListTimeout bounds how long the app gets to list its tests.
const testListTimeout = 5 * time.Minute

// ErrNoUnityConfig says the test config has no unity settings to list tests.
var ErrNoUnityConfig = errors.New("test config has no unity settings")

// ListTests returns the tests of the unity config: the app lists them for
// the types "all" and "all of category", the others are stored in the config.
func (r *Runner) ListTests(ctx context.Context, connected []DeviceMap) ([]models.UnityTestFunction, error) {
	unity := r.Config.Unity
	if unity == nil {
		return nil, ErrNoUnityConfig
	}
	switch unity.UnityTestCategoryType {
	case models.AllTest, models.AllOfCategory:
		return r.appTests(ctx, connected[0], unity)
	default:
		return r.Store.UnityTests(ctx, unity.ID)
	}
}

// appTests asks the app of the device which tests it has.
func (r *Runner) appTests(ctx context.Context, dev DeviceMap, unity *models.TestConfigUnity) ([]models.UnityTestFunction, error) {
	r.LogInfo("UnityTestCategoryType active requesting PlayMode tests")
	get := &action.TestsGet{}
	if err := tester_action.NewExecutor(r.Devices).Execute(ctx, tester_action.Call{Device: dev.Device, Action: get, Timeout: testListTimeout}); err != nil {
		return nil, fmt.Errorf("send action to select all tests failed: %w", err)
	}
	var list []models.UnityTestFunction
	for _, t := range get.Tests {
		if unity.UnityTestCategoryType == models.AllTest || inCategories(t.Categories, unity.Categories) {
			list = append(list, models.UnityTestFunction{Assembly: t.Assembly, Class: t.Class, Method: t.Method})
		}
	}
	return list, nil
}

// inCategories reports whether one of the categories of the test is in the
// comma separated list.
func inCategories(testCategories []string, wanted string) bool {
	list := strings.Split(wanted, ",")
	return slices.ContainsFunc(testCategories, func(c string) bool { return slices.Contains(list, c) })
}
