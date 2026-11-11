package control

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

const (
	// how often a followed log is read
	logPollInterval = time.Second
	// the most lines one read returns
	logPollMaxLines = 2000
	// the longest wait after read errors
	logMaxBackoff = 10 * time.Second
)

// logLineTime matches the time at the start of a "logcat -v time" line.
var logLineTime = regexp.MustCompile(`^\d\d-\d\d \d\d:\d\d:\d\d\.\d{3}`)

// logCursor remembers how far a followed log was sent. A read since a time
// returns the lines at that time again; the cursor drops those it sent.
type logCursor struct {
	last string
	// how often each line with the time last was sent
	sent map[string]int
}

// next returns the lines of a read that weren't sent yet and moves the
// cursor behind them.
func (c *logCursor) next(lines []string) []string {
	var out []string
	inBatch := map[string]int{}
	keepUntimed := false
	for _, line := range lines {
		if strings.HasPrefix(line, "--------- ") {
			// "--------- beginning of main"
			continue
		}
		ts := logLineTime.FindString(line)
		if ts == "" {
			// a line without time belongs to the line before it
			if keepUntimed {
				out = append(out, line)
			}
			continue
		}
		keepUntimed = false
		switch {
		case c.last != "" && ts < c.last:
			continue
		case ts == c.last:
			inBatch[line]++
			if inBatch[line] <= c.sent[line] {
				continue
			}
		default:
			c.last = ts
			c.sent = map[string]int{}
			inBatch = map[string]int{line: 1}
		}
		c.sent[line]++
		keepUntimed = true
		out = append(out, line)
	}
	return out
}

func (s *Service) readLogsSince(dev device.Device, since string) ([]string, error) {
	if follower, ok := dev.(device.LogFollower); ok && since != "" {
		return follower.ReadLogsSince(since, logPollMaxLines)
	}
	reader, ok := dev.(device.LogReader)
	if !ok {
		return nil, device.ErrUnsupported
	}
	return reader.ReadLogs(logPollMaxLines)
}

// FollowLogs sends the last backlog log lines of dev and then every new line
// until ctx ends. onLines gets each batch; returning an error stops. onError
// gets read errors, which don't stop following (e.g. a node reconnects).
func (s *Service) FollowLogs(ctx context.Context, dev device.Device, backlog int, onLines func([]string) error, onError func(error) error) error {
	if _, ok := dev.(device.LogReader); !ok {
		return device.ErrUnsupported
	}
	cursor := &logCursor{}
	wait := logPollInterval
	first := true
	for {
		var (
			lines []string
			err   error
		)
		if first {
			lines, err = s.Logs(dev, backlog)
		} else {
			lines, err = s.readLogsSince(dev, cursor.last)
		}
		if err != nil {
			wait = min(wait*2, logMaxBackoff)
			if cbErr := onError(err); cbErr != nil {
				return cbErr
			}
		} else {
			first = false
			wait = logPollInterval
			if fresh := cursor.next(lines); len(fresh) > 0 {
				if err := onLines(fresh); err != nil {
					return err
				}
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
		}
	}
}
