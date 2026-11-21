package api_test

import (
	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

// seedContract stores the fixed dataset every contract case runs against.
// The ids follow from the insert order: project proj (1) with app 1, test 1,
// bundle 1, parameter 1, runs 1 and 2 with protocols 1 and 2, device 1,
// node 1, user 1 (Alice) and an access token 1 of the project.
func seedContract(env *apitest.Env) {
	db := env.DB
	project := testkit.Project(db, func(p *models.Project) { p.Identifier, p.Name = "proj", "Project" })
	app := testkit.App(db, func(a *models.App) {
		a.ProjectID, a.Identifier, a.Name = project.ID, "com.example.app", "App"
	})
	db.Create(&models.AppParameter{AppID: app.ID, Name: "level", Type: "string"})
	db.Create(&models.AppBinary{AppID: app.ID, Name: "app.apk", Platform: "android", Version: "1.0", Identifier: "com.example.app", Hash: "abc", Tags: "release"})
	test := testkit.Test(db, func(x *models.Test) { x.AppID, x.Name = app.ID, "Smoke" })
	db.Create(&models.TestConfig{TestID: test.ID, Type: models.TestTypeUnity, AllDevices: true})
	device := testkit.Device(db, func(d *models.Device) {
		d.DeviceIdentifier, d.Name, d.Manager, d.PlatformType = "dev-1", "Pixel", "android_device", models.PlatformTypeAndroid
	})
	for i := 0; i < 2; i++ {
		run := testkit.TestRun(db, func(r *models.TestRun) { r.TestID, r.AppBinaryID, r.SessionID = test.ID, 1, "session" })
		testkit.Protocol(db, func(p *models.TestProtocol) {
			p.TestRunID, p.DeviceID, p.TestName = run.ID, &device.ID, "Suite Case"
			p.TestResult = models.TestResultSuccess
		})
	}
	db.Create(&models.Node{Identifier: "node-1", Name: "Node"})
	testkit.User(db, func(u *models.User) { u.Name = "Alice" })
	testkit.Token(db, func(tk *models.AccessToken) {
		tk.ProjectID, tk.Name, tk.Token, tk.Hint = &project.ID, "ci", "seed-token", "seed"
	})
}
