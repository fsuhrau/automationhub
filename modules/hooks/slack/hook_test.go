package slack

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/modules/hooks"
	slackapi "github.com/slack-go/slack"
)

func TestMapLevelToColor(t *testing.T) {
	tests := map[string]struct {
		level hooks.Level
		want  string
	}{
		"success":  {hooks.LevelSuccess, "#25BA0E"},
		"error":    {hooks.LevelError, "#cf4449"},
		"unstable": {hooks.LevelUnstable, "#f5d22e"},
		"unknown":  {hooks.Level(42), "white"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := mapLevelToColor(tc.level); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

type posted struct {
	channel     string
	attachments string
}

func slackServer(t *testing.T, status int, out chan<- posted) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		out <- posted{channel: r.PostForm.Get("channel"), attachments: r.PostForm.Get("attachments")}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestSendPostsAnAttachmentToTheChannel(t *testing.T) {
	tests := map[string]struct {
		link       string
		wantButton bool
	}{
		"with link":    {"http://hub/run/1", true},
		"without link": {"", false},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			out := make(chan posted, 1)
			srv := slackServer(t, http.StatusOK, out)
			h := NewHook(config.Hook{Token: "tok", Channel: "#ci"}, slackapi.OptionAPIURL(srv.URL+"/"))

			h.Send(hooks.Message{Title: "title", Text: "message", Link: tc.link, Level: hooks.LevelError})

			got := <-out
			if got.channel != "#ci" {
				t.Errorf("channel = %q", got.channel)
			}
			for _, want := range []string{`"title":"title"`, `"text":"message"`, `"color":"#cf4449"`} {
				if !strings.Contains(got.attachments, want) {
					t.Errorf("attachments %s lack %s", got.attachments, want)
				}
			}
			if has := strings.Contains(got.attachments, `"url":"http://hub/run/1"`); has != tc.wantButton {
				t.Errorf("button = %v, want %v: %s", has, tc.wantButton, got.attachments)
			}
		})
	}
}

func TestSendLogsAFailedPostWithoutPanicking(t *testing.T) {
	out := make(chan posted, 1)
	srv := slackServer(t, http.StatusInternalServerError, out)
	h := NewHook(config.Hook{Channel: "#ci"}, slackapi.OptionAPIURL(srv.URL+"/"))

	h.Send(hooks.Message{Title: "t", Text: "m", Level: hooks.LevelSuccess})

	if got := <-out; got.channel != "#ci" {
		t.Errorf("channel = %q", got.channel)
	}
}
