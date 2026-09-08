package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/virer/konsumo/models"
)

func TestLoadDataMissingFile(t *testing.T) {
	dir := t.TempDir()
	store := &Store{FilePath: filepath.Join(dir, "missing.json")}

	entries, err := store.LoadData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(entries))
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "consumption.json")
	store := &Store{FilePath: path}

	entries := []models.ConsumptionEntry{
		{
			Date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Category: "water",
			Water:    123.45,
		},
		{
			Date:             time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			Category:         "electricity",
			ElectricityDay:   100,
			ElectricityNight: 50,
		},
	}

	if err := store.SaveData(entries); err != nil {
		t.Fatalf("SaveData failed: %v", err)
	}

	loaded, err := store.LoadData()
	if err != nil {
		t.Fatalf("LoadData failed: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded))
	}
	if loaded[0].Water != 123.45 {
		t.Errorf("expected water 123.45, got %f", loaded[0].Water)
	}
	if loaded[1].ElectricityDay != 100 || loaded[1].ElectricityNight != 50 {
		t.Errorf("unexpected electricity values: day=%f night=%f", loaded[1].ElectricityDay, loaded[1].ElectricityNight)
	}
}

func TestLoadDataInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	store := &Store{FilePath: path}
	_, err := store.LoadData()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestPackageLevelFunctionsUseDefaultStore(t *testing.T) {
	orig := defaultStore
	defer SetStore(orig)

	dir := t.TempDir()
	path := filepath.Join(dir, "consumption.json")
	SetStore(&Store{FilePath: path})

	entry := models.ConsumptionEntry{
		Date:     time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		Category: "fuel",
		Gasoline: 500,
	}
	if err := SaveData([]models.ConsumptionEntry{entry}); err != nil {
		t.Fatalf("SaveData failed: %v", err)
	}

	loaded, err := LoadData()
	if err != nil {
		t.Fatalf("LoadData failed: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Gasoline != 500 {
		t.Fatalf("unexpected loaded data: %+v", loaded)
	}
}
