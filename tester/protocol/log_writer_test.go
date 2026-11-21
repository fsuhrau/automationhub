package protocol

import (
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Keys belong to the test that reported them: each test of a run keeps its
// own value, and a key sent again replaces the test's value.
func TestKeysAreStoredPerProtocol(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Device{}, &models.TestRun{}, &models.TestProtocol{}, &models.ProtocolKey{}, &models.ProtocolEntry{}, &models.ProtocolPerformanceEntry{}); err != nil {
		t.Fatal(err)
	}
	dev := models.Device{DeviceIdentifier: "phone"}
	db.Create(&dev)
	run := models.TestRun{}
	db.Create(&run)

	w := NewProtocolWriter(db, RunTarget{ProjectID: "p", AppID: 1, TestName: "t", Run: &run})
	runner, err := w.NewProtocol(dev, "Runner")
	if err != nil {
		t.Fatal(err)
	}
	a, err := w.NewSubProtocol("Suite/A", runner.Writer)
	if err != nil {
		t.Fatal(err)
	}
	a.Writer.SetKey("player_id", "1234")
	a.Writer.SetKey("world", "en1")
	a.Writer.SetKey("player_id", "5678")
	b, err := w.NewSubProtocol("Suite/B", runner.Writer)
	if err != nil {
		t.Fatal(err)
	}
	b.Writer.SetKey("player_id", "42")

	keysOf := func(id uint) []models.ProtocolKey {
		var keys []models.ProtocolKey
		db.Where("test_protocol_id = ?", id).Order("key").Find(&keys)
		return keys
	}
	if keys := keysOf(*a.TestProtocolId()); len(keys) != 2 || keys[0].Key != "player_id" || keys[0].Value != "5678" || keys[1].Key != "world" || keys[1].Value != "en1" {
		t.Fatalf("keys of A: %+v", keys)
	}
	if keys := keysOf(*b.TestProtocolId()); len(keys) != 1 || keys[0].Value != "42" {
		t.Fatalf("keys of B: %+v", keys)
	}
	if keys := keysOf(*runner.TestProtocolId()); len(keys) != 0 {
		t.Fatalf("keys of the runner: %+v", keys)
	}

	// the run endpoint loads them with the protocols
	var loaded models.TestRun
	db.Preload("Protocols").Preload("Protocols.Keys").First(&loaded, run.ID)
	found := 0
	for _, p := range loaded.Protocols {
		found += len(p.Keys)
	}
	if found != 3 {
		t.Fatalf("preloaded keys: %d", found)
	}
	if len(a.Writer.errs) > 0 || len(b.Writer.errs) > 0 {
		t.Fatalf("errors: %v %v", a.Writer.errs, b.Writer.errs)
	}
}

// Everything a test logs lands in the protocol log with its level; an error
// is remembered besides.
func TestLogWriterStoresDataLogsAndErrors(t *testing.T) {
	db := testkit.DB(t)
	run := testkit.TestRun(db)
	dev := testkit.Device(db)
	runner, err := NewProtocolWriter(db, RunTarget{ProjectID: "p", AppID: 1, TestName: "t", Run: run}).NewProtocol(*dev, "Runner")
	if err != nil {
		t.Fatal(err)
	}

	runner.Writer.Data("screen", "shot.png")
	runner.Writer.Log("step", "tap %d", 3)
	runner.Writer.Error("step", "failed %s", "badly")

	var entries []models.ProtocolEntry
	db.Where("test_protocol_id = ?", runner.p.ID).Order("id").Find(&entries)
	if len(entries) != 3 {
		t.Fatalf("entries %+v", entries)
	}
	if entries[0].Level != "log" || entries[0].Data != "shot.png" || entries[0].Message != "" {
		t.Errorf("data entry %+v", entries[0])
	}
	if entries[1].Level != "log" || entries[1].Message != "tap 3" || entries[1].Data != "" {
		t.Errorf("log entry %+v", entries[1])
	}
	if entries[2].Level != "error" || entries[2].Message != "failed badly" || len(runner.Writer.errs) != 1 {
		t.Errorf("error entry %+v, errors %v", entries[2], runner.Writer.errs)
	}
}
