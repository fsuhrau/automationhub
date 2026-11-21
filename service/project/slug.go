package project

import "strings"

const maxSlugLength = 48

// slug turns a name into an identifier that is safe in URLs: lower case
// letters, digits and single dashes.
func slug(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case b.Len() > 0 && !dash:
			b.WriteByte('-')
			dash = true
		}
	}
	slug := strings.TrimSuffix(b.String(), "-")
	if len(slug) > maxSlugLength {
		slug = strings.TrimSuffix(slug[:maxSlugLength], "-")
	}
	if slug == "" {
		slug = "project"
	}
	return slug
}
