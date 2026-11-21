// Package validate holds the input rules the services share.
package validate

import (
	"strings"

	"github.com/fsuhrau/automationhub/apperr"
)

// TrimRequired returns name without surrounding space. A name that is empty
// afterwards is an Invalid error with the message msg.
func TrimRequired(name, msg string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", apperr.Invalid("%s", msg)
	}
	return name, nil
}
