package httpx_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/gin-gonic/gin"
)

func TestCurrentUser(t *testing.T) {
	alice := authentication.User{Name: "Alice", Login: "alice"}
	var nilUser *authentication.User
	tests := []struct {
		name   string
		value  any
		set    bool
		want   authentication.User
		wantOK bool
	}{
		{"value", alice, true, alice, true},
		{"pointer", &alice, true, alice, true},
		{"nil pointer", nilUser, true, authentication.User{}, false},
		{"token owner name", "ci", true, authentication.User{}, false},
		{"none", nil, false, authentication.User{}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			if tc.set {
				c.Set(authentication.UserKey, tc.value)
			}
			got, ok := httpx.CurrentUser(c)
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("got %+v %v, want %+v %v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}
