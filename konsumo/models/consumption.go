// models/consumption.go
package models

import "time"

type ConsumptionEntry struct {
	Date             time.Time `json:"date"`
	Category         string    `json:"category"` // "water", "fuel", "electricity"
	Water            float64   `json:"water,omitempty"`
	ElectricityDay   float64   `json:"electricity_day,omitempty"`
	ElectricityNight float64   `json:"electricity_night,omitempty"`
	Gasoline         float64   `json:"gasoline,omitempty"`
}
