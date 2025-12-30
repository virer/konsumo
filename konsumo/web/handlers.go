package web

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/virer/konsumo/models"
	"github.com/virer/konsumo/storage"
)

// MonthlyDataPoint represents one data point per month
type MonthlyDataPoint struct {
	Year  int     `json:"year"`
	Month int     `json:"month"` // 1-12
	Rate  float64 `json:"rate"`  // daily rate for that month
}

// LatestDataPoint represents the latest entry and daily consumption for a category
type LatestDataPoint struct {
	Date            time.Time `json:"date"`
	Value           float64   `json:"value"`
	DailyConsumption float64  `json:"daily_consumption"`
}

// ChartData contains aggregated data for charts
type ChartData struct {
	Entries           []models.ConsumptionEntry  `json:"entries"`
	Electricity       map[int][]MonthlyDataPoint `json:"electricity"` // year -> monthly points
	Water             map[int][]MonthlyDataPoint `json:"water"`       // year -> monthly points
	Fuel              map[int][]MonthlyDataPoint `json:"fuel"`        // year -> monthly points
	LatestElectricity []LatestDataPoint          `json:"latest_electricity,omitempty"`
	LatestWater       []LatestDataPoint          `json:"latest_water,omitempty"`
	LatestFuel        []LatestDataPoint          `json:"latest_fuel,omitempty"`
}

var funcMap = template.FuncMap{
	"marshal": func(v any) template.JS {
		a, _ := json.Marshal(v)
		return template.JS(a)
	},
}

// HomeHandler serves the dashboard and form
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := storage.LoadData()
	if err != nil {
		log.Printf("Error loading data: %v", err)
		http.Error(w, "Unable to load data", http.StatusInternalServerError)
		return
	}

	// Aggregate data by month
	chartData := ChartData{
		Entries:           entries,
		Electricity:       aggregateElectricity(entries),
		Water:             aggregateWater(entries),
		Fuel:              aggregateFuel(entries),
		LatestElectricity: getLatestElectricity(entries),
		LatestWater:       getLatestWater(entries),
		LatestFuel:        getLatestFuel(entries),
	}

	tmplPath := filepath.Join("web", "templates", "index.html")
	tmpl := template.New("index.html").Funcs(funcMap)
	tmpl, err = tmpl.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, chartData); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// SubmitHandler handles the form submission to add a new entry
func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	category := r.FormValue("category")
	dateStr := r.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	entry := models.ConsumptionEntry{
		Date:     date,
		Category: category,
	}

	switch category {
	case "water":
		entry.Water = parseFloat(r.FormValue("water"))
	case "electricity":
		entry.ElectricityDay = parseFloat(r.FormValue("electricity_day"))
		entry.ElectricityNight = parseFloat(r.FormValue("electricity_night"))
	case "fuel":
		entry.Gasoline = parseFloat(r.FormValue("gasoline"))
	default:
		http.Error(w, "Unknown category", http.StatusBadRequest)
		return
	}

	entries, _ := storage.LoadData()
	entries = append(entries, entry)

	if err := storage.SaveData(entries); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	// Redirect back to form tab with the same category selected
	http.Redirect(w, r, "/?tab=form&category="+category, http.StatusSeeOther)
}

func parseFloat(value string) float64 {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Printf("Error parsing float: %v", err)
		return 0.0
	}
	return result
}

// aggregateElectricity aggregates electricity data by month
func aggregateElectricity(entries []models.ConsumptionEntry) map[int][]MonthlyDataPoint {
	electricityEntries := []models.ConsumptionEntry{}
	for _, e := range entries {
		if e.Category == "electricity" {
			electricityEntries = append(electricityEntries, e)
		}
	}
	sort.Slice(electricityEntries, func(i, j int) bool {
		return electricityEntries[i].Date.Before(electricityEntries[j].Date)
	})

	result := make(map[int][]MonthlyDataPoint)
	monthlyRates := make(map[string]float64) // "year-month" -> rate

	for i := 1; i < len(electricityEntries); i++ {
		prev := electricityEntries[i-1]
		curr := electricityEntries[i]

		if prev.Date.Year() != curr.Date.Year() {
			continue // Skip cross-year calculations
		}

		days := curr.Date.Sub(prev.Date).Hours() / 24
		if days <= 0 {
			continue
		}

		prevTotal := prev.ElectricityDay + prev.ElectricityNight
		currTotal := curr.ElectricityDay + curr.ElectricityNight
		delta := currTotal - prevTotal
		dailyRate := delta / days

		// Assign rate to the previous entry's month (where the consumption period started)
		year := prev.Date.Year()
		month := int(prev.Date.Month())
		key := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		// Average if multiple entries in same month
		if existing, ok := monthlyRates[key]; ok {
			monthlyRates[key] = (existing + dailyRate) / 2
		} else {
			monthlyRates[key] = dailyRate
		}
	}

	// Convert to MonthlyDataPoint structure
	for key, rate := range monthlyRates {
		var t time.Time
		t, _ = time.Parse("2006-01", key)
		year := t.Year()
		month := int(t.Month())

		result[year] = append(result[year], MonthlyDataPoint{
			Year:  year,
			Month: month,
			Rate:  rate,
		})
	}

	// Sort by month for each year
	for year := range result {
		sort.Slice(result[year], func(i, j int) bool {
			return result[year][i].Month < result[year][j].Month
		})
	}

	return result
}

// aggregateWater aggregates water data by month
func aggregateWater(entries []models.ConsumptionEntry) map[int][]MonthlyDataPoint {
	waterEntries := []models.ConsumptionEntry{}
	for _, e := range entries {
		if e.Category == "water" {
			waterEntries = append(waterEntries, e)
		}
	}
	sort.Slice(waterEntries, func(i, j int) bool {
		return waterEntries[i].Date.Before(waterEntries[j].Date)
	})

	result := make(map[int][]MonthlyDataPoint)
	monthlyRates := make(map[string]float64)

	for i := 1; i < len(waterEntries); i++ {
		prev := waterEntries[i-1]
		curr := waterEntries[i]

		if prev.Date.Year() != curr.Date.Year() {
			continue
		}

		days := curr.Date.Sub(prev.Date).Hours() / 24
		if days <= 0 {
			continue
		}

		delta := curr.Water - prev.Water
		dailyRate := delta / days

		// Assign rate to the previous entry's month (where the consumption period started)
		year := prev.Date.Year()
		month := int(prev.Date.Month())
		key := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		if existing, ok := monthlyRates[key]; ok {
			monthlyRates[key] = (existing + dailyRate) / 2
		} else {
			monthlyRates[key] = dailyRate
		}
	}

	for key, rate := range monthlyRates {
		var t time.Time
		t, _ = time.Parse("2006-01", key)
		year := t.Year()
		month := int(t.Month())

		result[year] = append(result[year], MonthlyDataPoint{
			Year:  year,
			Month: month,
			Rate:  rate,
		})
	}

	for year := range result {
		sort.Slice(result[year], func(i, j int) bool {
			return result[year][i].Month < result[year][j].Month
		})
	}

	return result
}

// aggregateFuel aggregates fuel data by month
// Groups data by heating year (August to July) instead of calendar year
func aggregateFuel(entries []models.ConsumptionEntry) map[int][]MonthlyDataPoint {
	fuelEntries := []models.ConsumptionEntry{}
	for _, e := range entries {
		if e.Category == "fuel" {
			fuelEntries = append(fuelEntries, e)
		}
	}
	sort.Slice(fuelEntries, func(i, j int) bool {
		return fuelEntries[i].Date.Before(fuelEntries[j].Date)
	})

	result := make(map[int][]MonthlyDataPoint)
	monthlyRates := make(map[string]float64)

	for i := 1; i < len(fuelEntries); i++ {
		prev := fuelEntries[i-1]
		curr := fuelEntries[i]

		// For heating year, we need to handle cross-year periods
		// A heating year runs from August to July
		// So August-December belong to the current calendar year's heating period
		// January-July belong to the next calendar year's heating period
		prevHeatingYear := getHeatingYear(prev.Date)
		currHeatingYear := getHeatingYear(curr.Date)

		if prevHeatingYear != currHeatingYear {
			continue
		}

		days := curr.Date.Sub(prev.Date).Hours() / 24
		if days <= 0 {
			continue
		}

		delta := prev.Gasoline - curr.Gasoline // tank got smaller
		dailyRate := delta / days

		// Assign rate to the previous entry's month (where the consumption period started)
		year := prev.Date.Year()
		month := int(prev.Date.Month())
		key := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		if existing, ok := monthlyRates[key]; ok {
			monthlyRates[key] = (existing + dailyRate) / 2
		} else {
			monthlyRates[key] = dailyRate
		}
	}

	for key, rate := range monthlyRates {
		var t time.Time
		t, _ = time.Parse("2006-01", key)
		month := int(t.Month())

		// Group by heating year: August-December use current year, January-July use previous year
		heatingYear := getHeatingYear(t)

		result[heatingYear] = append(result[heatingYear], MonthlyDataPoint{
			Year:  heatingYear,
			Month: month,
			Rate:  rate,
		})
	}

	// Sort by heating month order: August (8) first, then 9,10,11,12,1,2,3,4,5,6,7
	for year := range result {
		sort.Slice(result[year], func(i, j int) bool {
			monthI := result[year][i].Month
			monthJ := result[year][j].Month
			// Convert to heating month order: Aug=1, Sep=2, ..., Jul=12
			heatingMonthI := getHeatingMonth(monthI)
			heatingMonthJ := getHeatingMonth(monthJ)
			return heatingMonthI < heatingMonthJ
		})
	}

	return result
}

// getHeatingYear returns the heating year for a given date
// Heating year starts in August, so Aug-Dec belong to current year's heating period
// Jan-Jul belong to the previous year's heating period
func getHeatingYear(date time.Time) int {
	month := int(date.Month())
	if month >= 8 {
		return date.Year()
	}
	return date.Year() - 1
}

// getHeatingMonth converts calendar month to heating month position (1-12)
// August (8) -> 1, September (9) -> 2, ..., July (7) -> 12
func getHeatingMonth(month int) int {
	if month >= 8 {
		return month - 7 // Aug=1, Sep=2, Oct=3, Nov=4, Dec=5
	}
	return month + 5 // Jan=6, Feb=7, Mar=8, Apr=9, May=10, Jun=11, Jul=12
}

// getLatestElectricity returns the 5 latest electricity entries and daily consumption
func getLatestElectricity(entries []models.ConsumptionEntry) []LatestDataPoint {
	electricityEntries := []models.ConsumptionEntry{}
	for _, e := range entries {
		if e.Category == "electricity" {
			electricityEntries = append(electricityEntries, e)
		}
	}
	if len(electricityEntries) < 2 {
		return nil
	}
	sort.Slice(electricityEntries, func(i, j int) bool {
		return electricityEntries[i].Date.Before(electricityEntries[j].Date)
	})

	result := []LatestDataPoint{}
	// Need at least 6 entries to show 5 data points (each needs a previous entry)
	if len(electricityEntries) < 6 {
		// If we have 2-5 entries, show what we can
		if len(electricityEntries) >= 2 {
			for i := 1; i < len(electricityEntries); i++ {
				curr := electricityEntries[i]
				prev := electricityEntries[i-1]

				days := curr.Date.Sub(prev.Date).Hours() / 24
				if days <= 0 {
					continue
				}

				prevTotal := prev.ElectricityDay + prev.ElectricityNight
				currTotal := curr.ElectricityDay + curr.ElectricityNight
				delta := currTotal - prevTotal
				dailyRate := delta / days

				result = append(result, LatestDataPoint{
					Date:             curr.Date,
					Value:            currTotal,
					DailyConsumption: dailyRate,
				})
			}
			// Reverse to show most recent first
			for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
				result[i], result[j] = result[j], result[i]
			}
		}
		return result
	}

	// Get the last 5 data points (need 6 entries total: last 5 + 1 previous)
	// Process from len-5 to len-1 (5 entries), each calculated from previous
	for i := len(electricityEntries) - 5; i < len(electricityEntries); i++ {
		curr := electricityEntries[i]
		prev := electricityEntries[i-1]

		days := curr.Date.Sub(prev.Date).Hours() / 24
		if days <= 0 {
			continue
		}

		prevTotal := prev.ElectricityDay + prev.ElectricityNight
		currTotal := curr.ElectricityDay + curr.ElectricityNight
		delta := currTotal - prevTotal
		dailyRate := delta / days

		result = append(result, LatestDataPoint{
			Date:             curr.Date,
			Value:            currTotal,
			DailyConsumption: dailyRate,
		})
	}

	// Reverse to show most recent first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// getLatestWater returns the 5 latest water entries and daily consumption
func getLatestWater(entries []models.ConsumptionEntry) []LatestDataPoint {
	waterEntries := []models.ConsumptionEntry{}
	for _, e := range entries {
		if e.Category == "water" {
			waterEntries = append(waterEntries, e)
		}
	}
	if len(waterEntries) < 2 {
		return nil
	}
	sort.Slice(waterEntries, func(i, j int) bool {
		return waterEntries[i].Date.Before(waterEntries[j].Date)
	})

	result := []LatestDataPoint{}
	// Need at least 6 entries to show 5 data points (each needs a previous entry)
	if len(waterEntries) < 6 {
		// If we have 2-5 entries, show what we can
		if len(waterEntries) >= 2 {
			for i := 1; i < len(waterEntries); i++ {
				curr := waterEntries[i]
				prev := waterEntries[i-1]

				days := curr.Date.Sub(prev.Date).Hours() / 24
				if days <= 0 {
					continue
				}

				delta := curr.Water - prev.Water
				dailyRate := delta / days

				result = append(result, LatestDataPoint{
					Date:             curr.Date,
					Value:            curr.Water,
					DailyConsumption: dailyRate,
				})
			}
			// Reverse to show most recent first
			for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
				result[i], result[j] = result[j], result[i]
			}
		}
		return result
	}

	// Get the last 5 data points (need 6 entries total: last 5 + 1 previous)
	// Process from len-5 to len-1 (5 entries), each calculated from previous
	for i := len(waterEntries) - 5; i < len(waterEntries); i++ {
		curr := waterEntries[i]
		prev := waterEntries[i-1]

		days := curr.Date.Sub(prev.Date).Hours() / 24
		if days <= 0 {
			continue
		}

		delta := curr.Water - prev.Water
		dailyRate := delta / days

		result = append(result, LatestDataPoint{
			Date:             curr.Date,
			Value:            curr.Water,
			DailyConsumption: dailyRate,
		})
	}

	// Reverse to show most recent first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// getLatestFuel returns the 5 latest fuel entries and daily consumption
func getLatestFuel(entries []models.ConsumptionEntry) []LatestDataPoint {
	fuelEntries := []models.ConsumptionEntry{}
	for _, e := range entries {
		if e.Category == "fuel" {
			fuelEntries = append(fuelEntries, e)
		}
	}
	if len(fuelEntries) < 2 {
		return nil
	}
	sort.Slice(fuelEntries, func(i, j int) bool {
		return fuelEntries[i].Date.Before(fuelEntries[j].Date)
	})

	result := []LatestDataPoint{}
	// Need at least 6 entries to show 5 data points (each needs a previous entry)
	if len(fuelEntries) < 6 {
		// If we have 2-5 entries, show what we can
		if len(fuelEntries) >= 2 {
			for i := 1; i < len(fuelEntries); i++ {
				curr := fuelEntries[i]
				prev := fuelEntries[i-1]

				days := curr.Date.Sub(prev.Date).Hours() / 24
				if days <= 0 {
					continue
				}

				delta := prev.Gasoline - curr.Gasoline // tank got smaller
				dailyRate := delta / days

				result = append(result, LatestDataPoint{
					Date:             curr.Date,
					Value:            curr.Gasoline,
					DailyConsumption: dailyRate,
				})
			}
			// Reverse to show most recent first
			for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
				result[i], result[j] = result[j], result[i]
			}
		}
		return result
	}

	// Get the last 5 data points (need 6 entries total: last 5 + 1 previous)
	// Process from len-5 to len-1 (5 entries), each calculated from previous
	for i := len(fuelEntries) - 5; i < len(fuelEntries); i++ {
		curr := fuelEntries[i]
		prev := fuelEntries[i-1]

		days := curr.Date.Sub(prev.Date).Hours() / 24
		if days <= 0 {
			continue
		}

		delta := prev.Gasoline - curr.Gasoline // tank got smaller
		dailyRate := delta / days

		result = append(result, LatestDataPoint{
			Date:             curr.Date,
			Value:            curr.Gasoline,
			DailyConsumption: dailyRate,
		})
	}

	// Reverse to show most recent first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}
