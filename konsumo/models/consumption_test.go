package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestConsumptionEntryJSONRoundTrip(t *testing.T) {
	entry := ConsumptionEntry{
		Date:             time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		Category:         "electricity",
		ElectricityDay:   120.5,
		ElectricityNight: 80.25,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded ConsumptionEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !decoded.Date.Equal(entry.Date) {
		t.Errorf("date mismatch: got %v want %v", decoded.Date, entry.Date)
	}
	if decoded.Category != entry.Category {
		t.Errorf("category mismatch: got %q want %q", decoded.Category, entry.Category)
	}
	if decoded.ElectricityDay != entry.ElectricityDay || decoded.ElectricityNight != entry.ElectricityNight {
		t.Errorf("electricity mismatch: got day=%f night=%f", decoded.ElectricityDay, decoded.ElectricityNight)
	}
}

func TestConsumptionEntryOmitEmpty(t *testing.T) {
	entry := ConsumptionEntry{
		Date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Category: "water",
		Water:    42.0,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	raw := string(data)
	for _, field := range []string{"gasoline", "electricity_day", "electricity_night"} {
		if strings.Contains(raw, field) {
			t.Errorf("expected field %q to be omitted, got JSON: %s", field, raw)
		}
	}
}
