package app

import (
	"regexp"
	"strings"
)

var (
	badgingAttrRegex  = regexp.MustCompile(`([A-Za-z][\w-]*)='((?:[^'\\]|\\.)*)'`)
	badgingQuoteRegex = regexp.MustCompile(`'((?:[^'\\]|\\.)*)'`)
)

// badgingLine is one "key: rest" line of the aapt output
type badgingLine struct {
	key   string
	rest  string
	attrs map[string]string
}

func newBadgingLine(key, rest string) badgingLine {
	rest = strings.TrimSpace(rest)
	attrs := map[string]string{}
	for _, m := range badgingAttrRegex.FindAllStringSubmatch(rest, -1) {
		attrs[m[1]] = m[2]
	}
	return badgingLine{key: key, rest: rest, attrs: attrs}
}

// value is the first quoted value of the line, or the whole line without one
func (l badgingLine) value() string {
	if m := badgingQuoteRegex.FindStringSubmatch(l.rest); m != nil {
		return m[1]
	}
	return l.rest
}

// list is every quoted value of the line
func (l badgingLine) list() []string {
	var out []string
	for _, m := range badgingQuoteRegex.FindAllStringSubmatch(l.rest, -1) {
		out = append(out, m[1])
	}
	return out
}
