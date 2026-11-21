package oauthflow

import (
	"fmt"

	"github.com/fsuhrau/automationhub/storage/models"
)

// freeName picks the hub name of a new account. The name is only a label
// (accounts are identified by id), but it has to be unique.
func (u users) freeName(p Profile) (string, error) {
	for _, name := range nameCandidates(p, u.provider) {
		taken, err := u.nameTaken(name)
		if err != nil {
			return "", err
		}
		if !taken {
			return name, nil
		}
	}
	return "", fmt.Errorf("no free user name for %q", p.Login)
}

func (u users) nameTaken(name string) (bool, error) {
	var n int64
	if err := u.db.Model(&models.User{}).Where("name = ?", name).Count(&n).Error; err != nil {
		return false, fmt.Errorf("find user name: %w", err)
	}
	return n > 0, nil
}

// maxNameCounter bounds the numbered names tried after the plain ones.
const maxNameCounter = 1000

// nameCandidates lists the names to try in order: the display name (else the
// login), the same with the provider, then the provider with a counter.
func nameCandidates(p Profile, provider string) []string {
	base := firstNonEmpty(p.Name, p.Login, p.ProviderUserID, provider)
	names := []string{base, fmt.Sprintf("%s (%s)", base, provider)}
	for i := 2; i <= maxNameCounter; i++ {
		names = append(names, fmt.Sprintf("%s (%s %d)", base, provider, i))
	}
	return names
}

// firstNonEmpty returns the first value that is set, else the last one.
func firstNonEmpty(first string, rest ...string) string {
	if first != "" || len(rest) == 0 {
		return first
	}
	return firstNonEmpty(rest[0], rest[1:]...)
}
