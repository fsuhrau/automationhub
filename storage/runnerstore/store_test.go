package runnerstore_test

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/runnerstore"
)

func TestCreateRunAndAddLog(t *testing.T) {
	db := testkit.DB(t)
	store := runnerstore.New(db)
	ctx := context.Background()
	test := testkit.Test(db)

	run := &models.TestRun{TestID: test.ID, SessionID: "s1"}
	if err := store.CreateRun(ctx, run); err != nil || run.ID == 0 {
		t.Fatalf("create run: id %d, %v", run.ID, err)
	}
	entry := &models.TestRunLogEntry{TestRunID: run.ID, Level: "log", Log: "hello"}
	if err := store.AddLog(ctx, entry); err != nil || entry.ID == 0 {
		t.Fatalf("add log: id %d, %v", entry.ID, err)
	}
	var n int64
	db.Model(&models.TestRunLogEntry{}).Where("test_run_id = ?", run.ID).Count(&n)
	if n != 1 {
		t.Fatalf("got %d log entries, want 1", n)
	}
}

func TestUnityTestsOfTheConfig(t *testing.T) {
	db := testkit.DB(t)
	store := runnerstore.New(db)
	db.Create(&models.UnityTestFunction{TestConfigUnityID: 7, Class: "A", Method: "one"})
	db.Create(&models.UnityTestFunction{TestConfigUnityID: 7, Class: "B", Method: "two"})
	db.Create(&models.UnityTestFunction{TestConfigUnityID: 8, Class: "C", Method: "other"})

	got, err := store.UnityTests(context.Background(), 7)
	if err != nil || len(got) != 2 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

// A closed database reports its errors with the operation.
func TestStoreErrorsAreWrapped(t *testing.T) {
	db := testkit.DB(t)
	store := runnerstore.New(db)
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	ctx := context.Background()

	if err := store.CreateRun(ctx, &models.TestRun{}); err == nil {
		t.Error("create run on a closed database succeeded")
	}
	if err := store.AddLog(ctx, &models.TestRunLogEntry{}); err == nil {
		t.Error("add log on a closed database succeeded")
	}
	if _, err := store.UnityTests(ctx, 1); err == nil {
		t.Error("unity tests on a closed database succeeded")
	}
}
