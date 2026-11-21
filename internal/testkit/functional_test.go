package testkit

import (
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// TestSeededProjectTreeReadsBackThroughGorm seeds the whole tree and loads it
// again with preloads, the way the hub's endpoints read it.
func TestSeededProjectTreeReadsBackThroughGorm(t *testing.T) {
	db := DB(t)
	project := Project(db, WithName[models.Project]("game"))
	app := App(db, func(a *models.App) { a.ProjectID = project.ID })
	test := Test(db, func(m *models.Test) { m.AppID = app.ID })
	run := TestRun(db, func(r *models.TestRun) { r.TestID = test.ID })
	dev := Device(db)
	Protocol(db, func(p *models.TestProtocol) { p.TestRunID = run.ID; p.DeviceID = &dev.ID })
	user := User(db)
	Token(db, func(tk *models.AccessToken) { tk.ProjectID = &project.ID; tk.UserID = &user.ID })

	var got models.Project
	err := db.Preload("Apps.Parameter").Preload("AccessTokens").First(&got, project.ID).Error
	if err != nil || got.Name != "game" || len(got.Apps) != 1 || len(got.AccessTokens) != 1 {
		t.Fatalf("project = %+v err=%v", got, err)
	}
	var gotRun models.TestRun
	err = db.Preload("Test.App.Project").Preload("Protocols.Device").First(&gotRun, run.ID).Error
	if err != nil || gotRun.Test.App.Project.ID != project.ID || len(gotRun.Protocols) != 1 ||
		gotRun.Protocols[0].Device.ID != dev.ID {
		t.Fatalf("run = %+v err=%v", gotRun, err)
	}
}

// TestHTTPHelpersAgainstGinHandler reads the seeded data through a gin route.
func TestHTTPHelpersAgainstGinHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := DB(t)
	Project(db, WithName[models.Project]("game"))
	r := gin.New()
	r.POST("/projects", func(c *gin.Context) {
		var in struct{ Name string }
		if err := c.BindJSON(&in); err != nil {
			return
		}
		var ps []models.Project
		db.Where("name = ?", in.Name).Find(&ps)
		c.JSON(http.StatusOK, ps)
	})
	rec := DoJSON(t, r, http.MethodPost, "/projects", map[string]string{"Name": "game"})
	if got := Decode[[]models.Project](t, rec); rec.Code != http.StatusOK || len(got) != 1 {
		t.Fatalf("code=%d projects=%v", rec.Code, got)
	}
}
