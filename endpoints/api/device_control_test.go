package api

import (
	"testing"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/gin-gonic/gin"
)

func TestRequestOwner(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"session user", authentication.User{ID: 7, Name: "admin"}, "admin (user:7)"},
		{"same name, other id", authentication.User{ID: 8, Name: "admin"}, "admin (user:8)"},
		{"user pointer", &authentication.User{ID: 7, Name: "admin"}, "admin (user:7)"},
		{"user without name", authentication.User{ID: 7}, "user:7"},
		{"user without id", authentication.User{Name: "admin", Login: "admin"}, "anonymous"},
		{"nil user pointer", (*authentication.User)(nil), "anonymous"},
		{"project token", "ci", "token:ci"},
		{"empty token name", "", "anonymous"},
		{"nobody", nil, "anonymous"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			if tc.value != nil {
				c.Set(authentication.UserKey, tc.value)
			}
			if got := requestOwner(c); got != tc.want {
				t.Fatalf("owner = %q, want %q", got, tc.want)
			}
		})
	}
}
