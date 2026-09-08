// storage/storage.go
package storage

import (
	"encoding/json"
	"os"

	"github.com/virer/konsumo/models"
)

type Store struct {
	FilePath string
}

func DefaultStore() *Store {
	return &Store{FilePath: "data/consumption.json"}
}

var defaultStore = DefaultStore()

func SetStore(s *Store) {
	defaultStore = s
}

func LoadData() ([]models.ConsumptionEntry, error) {
	return defaultStore.LoadData()
}

func SaveData(entries []models.ConsumptionEntry) error {
	return defaultStore.SaveData(entries)
}

func (s *Store) LoadData() ([]models.ConsumptionEntry, error) {
	file, err := os.ReadFile(s.FilePath)
	if err != nil {
		return []models.ConsumptionEntry{}, nil // start empty
	}
	var entries []models.ConsumptionEntry
	err = json.Unmarshal(file, &entries)
	return entries, err
}

func (s *Store) SaveData(entries []models.ConsumptionEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.FilePath, data, 0644)
}
