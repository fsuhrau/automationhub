package slack

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/modules/hooks"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type hook struct {
	client *slack.Client
	cfg    config.Hook
}

// NewHook posts to the Slack channel of cfg; opts adjust the Slack client
// (e.g. its API URL).
func NewHook(cfg config.Hook, opts ...slack.Option) *hook {
	h := &hook{
		cfg: cfg,
	}
	h.client = slack.New(cfg.Token, append([]slack.Option{slack.OptionDebug(true)}, opts...)...)
	return h
}

func mapLevelToColor(level hooks.Level) string {
	colors := map[hooks.Level]string{
		hooks.LevelSuccess:  "#25BA0E",
		hooks.LevelError:    "#cf4449",
		hooks.LevelUnstable: "#f5d22e",
	}

	if color, ok := colors[level]; ok {
		return color
	}

	return "white"
}

func (h *hook) Send(msg hooks.Message) {

	attachment := slack.Attachment{
		Title: msg.Title,
		Text:  msg.Text,
		Color: mapLevelToColor(msg.Level),
	}
	if len(msg.Link) > 0 {
		attachment.Actions = append(attachment.Actions, slack.AttachmentAction{
			Name: "Open",
			Text: "Open",
			Type: "button",
			URL:  msg.Link,
		})
	}
	_, _, err := h.client.PostMessage(h.cfg.Channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		logrus.Errorf("send slack message failed: %v", err)
	}
}
