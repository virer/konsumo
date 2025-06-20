// storage/storage.go
package storage

import (
	"encoding/json"
	"os"

	"github.com/virer/konsumo/models"
)

var filePath = "data/consumption.json"

func LoadData() ([]models.ConsumptionEntry, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return []models.ConsumptionEntry{}, nil // start empty
	}
	var entries []models.ConsumptionEntry
	err = json.Unmarshal(file, &entries)
	return entries, err
}

func SaveData(entries []models.ConsumptionEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
