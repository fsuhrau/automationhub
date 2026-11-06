package androiddevice

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/tools/exec"
)

var namedKeys = map[string]int{
	"home":        KEYCODE_HOME,
	"back":        KEYCODE_BACK,
	"enter":       KEYCODE_ENTER,
	"menu":        KEYCODE_MENU,
	"delete":      KEYCODE_DEL,
	"tab":         KEYCODE_TAB,
	"app_switch":  KEYCODE_APP_SWITCH,
	"power":       KEYCODE_POWER,
	"volume_up":   KEYCODE_VOLUME_UP,
	"volume_down": KEYCODE_VOLUME_DOWN,
	"escape":      KEYCODE_ESCAPE,
	"search":      KEYCODE_SEARCH,
}

func (d *Device) pressKey(key int) error {
	cmd := exec.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "keyevent", fmt.Sprintf("%d", key))
	return cmd.Run()
}

func (d *Device) swipe(fromX, fromY, toX, toY int) error {
	return d.Swipe(fromX, fromY, toX, toY, 0)
}

func (d *Device) sendText(text string) error {
	return d.InputText(text)
}

// Tap implements device.InputCapable.
func (d *Device) Tap(x, y int) error {
	d.Log("device", "Execute Tap: %d,%d", x, y)
	return exec.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "tap", strconv.Itoa(x), strconv.Itoa(y)).Run()
}

// Swipe implements device.InputCapable. A zero duration uses the system default.
func (d *Device) Swipe(fromX, fromY, toX, toY int, duration time.Duration) error {
	args := []string{"-s", d.DeviceID(), "shell", "input", "swipe",
		strconv.Itoa(fromX), strconv.Itoa(fromY), strconv.Itoa(toX), strconv.Itoa(toY)}
	if duration > 0 {
		args = append(args, strconv.FormatInt(duration.Milliseconds(), 10))
	}
	return exec.NewCommand("adb", args...).Run()
}

// InputText implements device.InputCapable.
func (d *Device) InputText(text string) error {
	if text == "" {
		return nil
	}
	return exec.NewCommand("adb", "-s", d.DeviceID(), "shell", "input", "text", escapeInputText(text)).Run()
}

// PressKey implements device.InputCapable.
func (d *Device) PressKey(key string) error {
	code, ok := namedKeys[strings.ToLower(key)]
	if !ok {
		n, err := strconv.Atoi(key)
		if err != nil || n < 0 {
			return fmt.Errorf("unknown key %q", key)
		}
		code = n
	}
	return d.pressKey(code)
}

// ReadLogs implements device.LogReader using a logcat dump.
func (d *Device) ReadLogs(lines int) ([]string, error) {
	if lines <= 0 {
		lines = 200
	}
	out, err := exec.NewCommand("adb", "-s", d.DeviceID(), "logcat", "-d", "-v", "time", "-t", strconv.Itoa(lines)).Output()
	if err != nil {
		return nil, err
	}
	result := strings.Split(strings.TrimRight(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n"), "\n")
	if len(result) == 1 && result[0] == "" {
		return nil, nil
	}
	return result, nil
}

// escapeInputText prepares text for "adb shell input text": the argument is
// parsed by the device shell, and input text uses %s for spaces.
func escapeInputText(text string) string {
	var b strings.Builder
	for _, r := range text {
		switch r {
		case ' ':
			b.WriteString("%s")
		case '%':
			// input text has no escape for a literal %, keep it apart from 's'
			b.WriteString(`\%`)
		case '\\', '\'', '"', '`', '$', '&', '|', ';', '<', '>', '(', ')', '*', '~', '#', '!', '?', '[', ']', '{', '}':
			b.WriteRune('\\')
			b.WriteRune(r)
		case '\n':
			// newlines can't be sent through input text
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
