package androiddevice

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice/scrcpy"
	"github.com/fsuhrau/automationhub/tools/exec"
)

var (
	physicalSizeRegex = regexp.MustCompile(`Physical size:\s*(\d+)x(\d+)`)
	overrideSizeRegex = regexp.MustCompile(`Override size:\s*(\d+)x(\d+)`)
)

type videoStream struct {
	*scrcpy.Session
	meta device.VideoMeta
}

func (s *videoStream) Meta() device.VideoMeta { return s.meta }

var touchActions = map[string]int{
	device.InputDown:   scrcpy.ActionDown,
	device.InputMove:   scrcpy.ActionMove,
	device.InputUp:     scrcpy.ActionUp,
	device.InputCancel: scrcpy.ActionCancel,
}

// InjectInput implements device.VideoInput through the scrcpy control socket.
func (s *videoStream) InjectInput(ev device.InputEvent) error {
	switch ev.Kind {
	case device.InputTouch:
		action, ok := touchActions[ev.Action]
		if !ok {
			return fmt.Errorf("unknown touch action %q", ev.Action)
		}
		if ev.Width <= 0 || ev.Height <= 0 {
			return fmt.Errorf("touch needs the video width and height")
		}
		return s.InjectTouch(action, ev.PointerID, ev.X, ev.Y, ev.Width, ev.Height)
	case device.InputKey:
		code, err := keyCode(ev.Key)
		if err != nil {
			return err
		}
		switch ev.Action {
		case device.InputDown:
			return s.InjectKeycode(scrcpy.ActionDown, code)
		case device.InputUp:
			return s.InjectKeycode(scrcpy.ActionUp, code)
		case device.InputPress, "":
			if err := s.InjectKeycode(scrcpy.ActionDown, code); err != nil {
				return err
			}
			return s.InjectKeycode(scrcpy.ActionUp, code)
		}
		return fmt.Errorf("unknown key action %q", ev.Action)
	case device.InputText:
		return s.InjectText(ev.Text)
	case device.InputScroll:
		return s.InjectScroll(ev.X, ev.Y, ev.Width, ev.Height, ev.ScrollX, ev.ScrollY)
	case device.InputResetVideo:
		return s.ResetVideo()
	}
	return fmt.Errorf("unknown input kind %q", ev.Kind)
}

// StartVideo implements device.VideoStreamer with the scrcpy server.
func (d *Device) StartVideo(ctx context.Context, opts device.VideoOptions) (device.VideoStream, error) {
	if state := d.DeviceState(); state != device.StateBooted {
		return nil, fmt.Errorf("device %s is %s", d.DeviceID(), device.StateToString(state))
	}
	d.Log("device", "Start video stream (scrcpy %s)", scrcpy.Version)
	session, err := scrcpy.Start(ctx, d.DeviceID(), scrcpy.Options{MaxSize: opts.MaxSize, MaxFPS: opts.MaxFPS, BitRate: opts.BitRate, Control: true})
	if err != nil {
		return nil, err
	}
	w, h := session.Size()
	meta := device.VideoMeta{Codec: session.Codec(), Width: w, Height: h}
	if out, err := exec.NewCommand("adb", "-s", d.DeviceID(), "shell", "wm", "size").Output(); err == nil {
		meta.ScreenWidth, meta.ScreenHeight = parseWmSize(string(out))
	}
	return &videoStream{Session: session, meta: meta}, nil
}

// parseWmSize parses "wm size"; an override size wins, since input
// coordinates follow it.
func parseWmSize(out string) (int, int) {
	for _, re := range []*regexp.Regexp{overrideSizeRegex, physicalSizeRegex} {
		if m := re.FindStringSubmatch(out); m != nil {
			w, _ := strconv.Atoi(m[1])
			h, _ := strconv.Atoi(m[2])
			return w, h
		}
	}
	return 0, 0
}
