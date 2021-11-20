package notifier

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/modules/hooks"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/spf13/viper"
)

type testRunFinishedNotifier struct {
	hooks []hooks.Hook
}

func RegisterEventTestRunFinishedListener(hooks []hooks.Hook) {
	notifier := testRunFinishedNotifier{
		hooks: hooks,
	}
	events.TestRunFinished.Register(notifier)
}

func buildMessage(payload events.TestRunFinishedPayload) string {
	return fmt.Sprintf("Results:\n- Successful: %d\n- Unstable: %d\n- Failed: %d ", payload.Succeeded, payload.Unstable, payload.Failed)
}

func buildTitle(payload events.TestRunFinishedPayload) string {
	if testRun, ok := payload.TestRun.(*models.TestRun); ok {
		if testRun.Test != nil {
			return fmt.Sprintf("TestRun finished: %s", testRun.Test.Name)
		}
	}
	return "TestRun finished"
}

func getUrl(payload events.TestRunFinishedPayload) string {
	if testRun, ok := payload.TestRun.(*models.TestRun); ok {
		ip := viper.GetString("host_ip")
		return fmt.Sprintf("http://%s:8002/web/test/%d/run/%d", ip, testRun.TestID, testRun.ID)
	}
	return ""
}

func getLevel(payload events.TestRunFinishedPayload) hooks.Level {
	if payload.Unstable > 0 {
		return hooks.LevelUnstable
	}
	if payload.Failed > 0 {
		return hooks.LevelError
	}
	if payload.Succeeded > 0 {
		return hooks.LevelSuccess
	}
	return hooks.LevelUnstable
}

func (u testRunFinishedNotifier) Handle(payload events.TestRunFinishedPayload) {
	for _, hook := range u.hooks {
		hook.Send(buildTitle(payload), buildMessage(payload), getUrl(payload), getLevel(payload))
	}
}
