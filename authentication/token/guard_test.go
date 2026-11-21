package token

import (
	"net/http"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/storage/models"
)

func TestGuardDenials(t *testing.T) {
	db := testDB(t)
	project := models.Project{Name: "Game", Identifier: "game"}
	db.Create(&project)
	gone := uint(999)
	node := models.Node{Identifier: "n1", Name: "node"}
	db.Create(&node)
	past := time.Now().Add(-time.Hour)
	db.Create(&models.AccessToken{ProjectID: &project.ID, Name: "p", Token: "project"})
	db.Create(&models.AccessToken{ProjectID: &gone, Name: "p", Token: "orphan-project"})
	db.Create(&models.AccessToken{NodeID: &gone, Name: "n", Token: "orphan-node"})
	db.Create(&models.AccessToken{NodeID: &node.ID, Name: "n", Token: "node"})
	db.Create(&models.AccessToken{UserID: &gone, Name: "u", Token: StoredPersonal(PersonalPrefix + "orphan")})
	db.Create(&models.AccessToken{ProjectID: &project.ID, Name: "old", Token: "expired", ExpiresAt: &past})
	r := server(db)

	for name, c := range map[string]struct {
		path, token string
		code        int
	}{
		"expired":                {"/api/game/x", "expired", http.StatusUnauthorized},
		"project token no param": {"/api/projects", "project", http.StatusBadRequest},
		"project token orphaned": {"/api/game/x", "orphan-project", http.StatusUnauthorized},
		"node token orphaned":    {"/upload/x.apk", "orphan-node", http.StatusUnauthorized},
		"personal user missing":  {"/api/game/x", PersonalPrefix + "orphan", http.StatusUnauthorized},
		"node token wrong route": {"/api/projects", "node", http.StatusForbidden},
	} {
		if w := get(r, c.path, c.token); w.Code != c.code {
			t.Errorf("%s: %d %s, want %d", name, w.Code, w.Body, c.code)
		}
	}
}

func TestDenialHelpers(t *testing.T) {
	if personalDenial(true) != nil || personalDenial(false) == nil || *personalDenial(false) != *unauthorized() {
		t.Error("personalDenial should map ok to nil and failure to unauthorized")
	}
}
