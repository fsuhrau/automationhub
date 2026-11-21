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

// LogFollow says what FollowLogs sends and where it sends it.
type LogFollow struct {
	// Backlog is how many of the last log lines are sent first.
	Backlog int
	// OnLines gets each batch of new lines; returning an error stops.
	OnLines func([]string) error
	// OnError gets read errors, which don't stop following (e.g. a node
	// reconnects); returning an error stops.
	OnError func(error) error
}

// FollowLogs sends the last backlog log lines of dev and then every new line
// until ctx ends.
func (s *Service) FollowLogs(ctx context.Context, dev device.Device, follow LogFollow) error {
	if _, ok := dev.(device.LogReader); !ok {
		return device.ErrUnsupported
	}
	p := &logPoller{service: s, dev: dev, follow: follow, wait: logPollInterval, first: true}
	for {
		if err := p.poll(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-s.after(p.wait):
		}
	}
}

// logPoller reads a log again and again and sends what is new.
type logPoller struct {
	service *Service
	dev     device.Device
	follow  LogFollow
	cursor  logCursor
	wait    time.Duration
	first   bool
}

// read returns the backlog on the first read, then the lines since the cursor.
func (p *logPoller) read() ([]string, error) {
	if p.first {
		return p.service.Logs(p.dev, p.follow.Backlog)
	}
	return p.service.readLogsSince(p.dev, p.cursor.last)
}

// poll does one read: an error backs the next read off, new lines are sent.
func (p *logPoller) poll() error {
	lines, err := p.read()
	if err != nil {
		p.wait = min(p.wait*2, logMaxBackoff)
		return p.follow.OnError(err)
	}
	p.first, p.wait = false, logPollInterval
	if fresh := p.cursor.next(lines); len(fresh) > 0 {
		return p.follow.OnLines(fresh)
	}
	return nil
}
