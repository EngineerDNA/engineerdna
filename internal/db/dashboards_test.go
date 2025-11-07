package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, *DashboardStore) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE dashboards (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			persona TEXT,
			is_template BOOLEAN NOT NULL DEFAULT FALSE,
			is_system BOOLEAN NOT NULL DEFAULT FALSE,
			layout TEXT NOT NULL,
			filters TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE metric_snapshots (
			metric_name TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id TEXT,
			period_start TEXT NOT NULL,
			period_end TEXT NOT NULL,
			value REAL NOT NULL,
			metadata TEXT,
			computed_at TEXT NOT NULL,
			PRIMARY KEY (metric_name, entity_type, entity_id, period_start)
		);
		CREATE TABLE teams (id TEXT PRIMARY KEY, name TEXT NOT NULL);
		CREATE TABLE engineers (id TEXT PRIMARY KEY, canonical_name TEXT NOT NULL);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	db.Exec("INSERT INTO teams (id, name) VALUES ('team1', 'Test Team')")
	db.Exec("INSERT INTO engineers (id, canonical_name) VALUES ('eng1', 'Test Engineer')")

	return db, NewDashboardStore(db)
}

func TestDashboard_CreateAndGet(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	dashboard := &models.Dashboard{
		Name:        "Test Dashboard",
		Description: "Test description",
		Persona:     "engineer",
		Layout:      "[]",
	}

	if err := store.CreateDashboard(dashboard); err != nil {
		t.Fatalf("CreateDashboard failed: %v", err)
	}

	retrieved, err := store.GetDashboard(dashboard.ID)
	if err != nil || retrieved == nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}

	if retrieved.Name != dashboard.Name {
		t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, dashboard.Name)
	}
}

func TestDashboard_Update(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	dashboard := &models.Dashboard{Name: "Original", Layout: "[]"}
	store.CreateDashboard(dashboard)

	time.Sleep(10 * time.Millisecond)
	dashboard.Name = "Updated"
	if err := store.UpdateDashboard(dashboard); err != nil {
		t.Fatalf("UpdateDashboard failed: %v", err)
	}

	retrieved, _ := store.GetDashboard(dashboard.ID)
	if retrieved.Name != "Updated" {
		t.Errorf("Update failed: got %s", retrieved.Name)
	}
}

func TestDashboard_DeleteProtection(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	dashboard := &models.Dashboard{Name: "System", IsSystem: true, Layout: "[]"}
	store.CreateDashboard(dashboard)

	if err := store.DeleteDashboard(dashboard.ID); err == nil {
		t.Error("Expected error deleting system dashboard")
	}
}

func TestDashboard_List(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	for i := 0; i < 3; i++ {
		d := &models.Dashboard{Name: "Dashboard", Persona: "engineer", Layout: "[]"}
		d.IsTemplate = (i%2 == 0)
		store.CreateDashboard(d)
	}

	dashboards, total, err := store.ListDashboards("", false, 10, 0)
	if err != nil {
		t.Fatalf("ListDashboards failed: %v", err)
	}
	if total != 3 || len(dashboards) != 3 {
		t.Errorf("Expected 3 dashboards, got %d total, %d returned", total, len(dashboards))
	}
}

func TestDashboard_Clone(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	source := &models.Dashboard{
		Name: "Source", Layout: "[{}]", IsTemplate: true, IsSystem: true,
	}
	store.CreateDashboard(source)

	clone, err := store.CloneDashboard(source.ID, "Clone")
	if err != nil {
		t.Fatalf("CloneDashboard failed: %v", err)
	}

	if clone.IsTemplate || clone.IsSystem {
		t.Error("Clone should not be template or system")
	}
	if clone.Layout != source.Layout {
		t.Error("Layout not copied")
	}
}

func TestMetricSnapshot_CreateAndGet(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	snapshot := &models.MetricSnapshot{
		MetricName:  "test_metric",
		EntityType:  "team",
		EntityID:    "team1",
		PeriodStart: "2025-01-01",
		PeriodEnd:   "2025-01-31",
		Value:       42.5,
		Metadata:    "{\"test\":\"data\"}",
	}

	if err := store.CreateMetricSnapshot(snapshot); err != nil {
		t.Fatalf("CreateMetricSnapshot failed: %v", err)
	}

	retrieved, err := store.GetMetricSnapshot(snapshot.MetricName, snapshot.EntityType, snapshot.EntityID, snapshot.PeriodStart)
	if err != nil || retrieved == nil {
		t.Fatalf("GetMetricSnapshot failed: %v", err)
	}

	if retrieved.Value != snapshot.Value {
		t.Errorf("Value mismatch: got %f, want %f", retrieved.Value, snapshot.Value)
	}
}

func TestMetricSnapshot_Batch(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	snapshots := []*models.MetricSnapshot{
		{MetricName: "metric1", EntityType: "team", EntityID: "team1", PeriodStart: "2025-01-01", PeriodEnd: "2025-01-31", Value: 1.0},
		{MetricName: "metric1", EntityType: "team", EntityID: "team1", PeriodStart: "2025-02-01", PeriodEnd: "2025-02-28", Value: 2.0},
		{MetricName: "metric1", EntityType: "team", EntityID: "team1", PeriodStart: "2025-03-01", PeriodEnd: "2025-03-31", Value: 3.0},
	}

	if err := store.CreateMetricSnapshotsBatch(snapshots); err != nil {
		t.Fatalf("CreateMetricSnapshotsBatch failed: %v", err)
	}

	result, err := store.GetMetricSnapshotsInRange("metric1", "team", "team1", "2025-01-01", "2025-12-31", 100)
	if err != nil {
		t.Fatalf("GetMetricSnapshotsInRange failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 snapshots, got %d", len(result))
	}
}
