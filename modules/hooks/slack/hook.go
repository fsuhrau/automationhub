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

func NewHook(cfg config.Hook) *hook {
	h := &hook{
		cfg: cfg,
	}
	h.client = slack.New(cfg.Token, slack.OptionDebug(true))
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

func (h *hook) Send(title, message, link string, level hooks.Level) {

	attachment := slack.Attachment{
		Title: title,
		Text:  message,
		Color: mapLevelToColor(level),
	}
	if len(link) > 0 {
		attachment.Actions = append(attachment.Actions, slack.AttachmentAction{
			Name: "Open",
			Text: "Open",
			Type: "button",
			URL:  link,
		})
	}
	_, _, err := h.client.PostMessage(h.cfg.Channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		logrus.Errorf("send slack message failed: %v", err)
	}
}
