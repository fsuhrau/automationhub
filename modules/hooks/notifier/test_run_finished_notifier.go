package notifier

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/modules/hooks"
	"github.com/fsuhrau/automationhub/storage/models"
)

type testRunFinishedNotifier struct {
	hooks  []hooks.Hook
	hostIP string
}

// RegisterEventTestRunFinishedListener sends every finished test run to the
// hooks; hostIP is the host the links in the messages point to.
func RegisterEventTestRunFinishedListener(hooks []hooks.Hook, hostIP string) {
	notifier := testRunFinishedNotifier{
		hooks:  hooks,
		hostIP: hostIP,
	}
	events.TestRunFinished.Register(notifier)
}

func buildMessage(payload events.TestRunFinishedPayload) string {
	return fmt.Sprintf("Results:\n- Successful: %d\n- Unstable: %d\n- Failed: %d ", payload.Succeeded, payload.Unstable, payload.Failed)
}

func buildTitle(payload events.TestRunFinishedPayload) string {
	if len(payload.TestName) > 0 {
		return fmt.Sprintf("TestRun finished: %s", payload.TestName)
	}

	if testRun, ok := payload.TestRun.(*models.TestRun); ok {
		if testRun.Test != nil {
			return fmt.Sprintf("TestRun finished: %s", testRun.Test.Name)
		}
	}

	return "TestRun finished"
}

func getUrl(payload events.TestRunFinishedPayload, ip string) string {
	if testRun, ok := payload.TestRun.(*models.TestRun); ok {
		return fmt.Sprintf("http://%s:8002/project/%s/app/%d/test/%d/run/%d", ip, payload.ProjectID, payload.AppID, testRun.TestID, testRun.ID)
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
		hook.Send(hooks.Message{
			Title: buildTitle(payload),
			Text:  buildMessage(payload),
			Link:  getUrl(payload, u.hostIP),
			Level: getLevel(payload),
		})
	}
}
