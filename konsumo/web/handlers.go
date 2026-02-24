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
	Day   int     `json:"day"`   // day of month for graph positioning
	Rate  float64 `json:"rate"`  // daily rate for that month
}

// LatestDataPoint represents the latest entry and daily consumption for a category
type LatestDataPoint struct {
	Date             time.Time `json:"date"`
	Value            float64   `json:"value"`
	DailyConsumption float64   `json:"daily_consumption"`
	// For electricity, separate day and night values
	DayValue              float64 `json:"day_value,omitempty"`
	NightValue            float64 `json:"night_value,omitempty"`
	DayDailyConsumption   float64 `json:"day_daily_consumption,omitempty"`
	NightDailyConsumption float64 `json:"night_daily_consumption,omitempty"`
}

// FuelHeatingProjection holds fuel consumption projection for the current heating period (Sept–June).
type FuelHeatingProjection struct {
	PeriodLabel           string    `json:"period_label"`             // e.g. "2025-26"
	PeriodStart           time.Time `json:"period_start"`             // September 1
	PeriodEnd             time.Time `json:"period_end"`                // June 30
	ConsumedSoFar         float64   `json:"consumed_so_far"`          // L consumed in current period
	HistoricalAvgTotal    float64   `json:"historical_avg_total"`     // average L per full heating period (past years)
	HistoricalPeriodsUsed int      `json:"historical_periods_used"`   // number of past periods used for average
	ProjectedTotal        float64   `json:"projected_total"`           // consumed so far + projected rest
	RemainingMonths       float64   `json:"remaining_months"`          // months left in period (for projection)
	HasProjection         bool      `json:"has_projection"`            // true if we have historical data to project
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
	// Fuel heating period (Sept–June) projection for current period
	FuelHeatingProjection *FuelHeatingProjection `json:"fuel_heating_projection,omitempty"`
	// Latest entries for form display
	LatestEntry map[string]models.ConsumptionEntry `json:"latest_entry,omitempty"` // category -> latest entry
	// Last 10 entries for form tab
	Last10Entries []models.ConsumptionEntry `json:"last_10_entries,omitempty"`
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
		Entries:                entries,
		Electricity:            aggregateElectricity(entries),
		Water:                  aggregateWater(entries),
		Fuel:                   aggregateFuel(entries),
		LatestElectricity:      getLatestElectricity(entries),
		LatestWater:            getLatestWater(entries),
		LatestFuel:             getLatestFuel(entries),
		FuelHeatingProjection:  getFuelHeatingProjection(entries, time.Now()),
		LatestEntry:            getLatestEntries(entries),
		Last10Entries:          getLast10Entries(entries),
	}

	tmplPath := filepath.Join("ui", "templates", "index.html")
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

// DeleteHandler handles deletion of an entry by index
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get the index from the form
	indexStr := r.FormValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	// Load entries
	entries, err := storage.LoadData()
	if err != nil {
		log.Printf("Error loading data: %v", err)
		http.Error(w, "Unable to load data", http.StatusInternalServerError)
		return
	}

	// Sort entries by date descending to match the order shown in the form
	sorted := make([]models.ConsumptionEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.After(sorted[j].Date)
	})

	// Check if index is valid
	if index >= len(sorted) {
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}

	// Find the entry to delete in the original entries array
	entryToDelete := sorted[index]

	// Remove the entry from the original entries array
	newEntries := []models.ConsumptionEntry{}
	for i, e := range entries {
		// Compare entries by date and category to find the matching one
		if e.Date.Equal(entryToDelete.Date) && e.Category == entryToDelete.Category {
			// Check if it's the same entry by comparing all fields
			if (e.Category == "water" && e.Water == entryToDelete.Water) ||
				(e.Category == "electricity" && e.ElectricityDay == entryToDelete.ElectricityDay && e.ElectricityNight == entryToDelete.ElectricityNight) ||
				(e.Category == "fuel" && e.Gasoline == entryToDelete.Gasoline) {
				// Skip this entry
				continue
			}
		}
		newEntries = append(newEntries, entries[i])
	}

	// Save the updated entries
	if err := storage.SaveData(newEntries); err != nil {
		log.Printf("Error saving data: %v", err)
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	// Redirect back to form tab
	http.Redirect(w, r, "/?tab=form", http.StatusSeeOther)
}

func parseFloat(value string) float64 {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Printf("Error parsing float: %v", err)
		return 0.0
	}
	return result
}

// rateWithDay stores a rate calculation with its associated day
type rateWithDay struct {
	day  int
	rate float64
}

// aggregateElectricity aggregates electricity data by month
// Creates graph points on specific days: 8, 16, 24 if multiple points, or 10, 20 if only one
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
	// "year-month" -> []rateWithDay (all rate calculations for that month)
	monthlyRates := make(map[string][]rateWithDay)

	// Collect all rate calculations grouped by month
	for i := 1; i < len(electricityEntries); i++ {
		prev := electricityEntries[i-1]
		curr := electricityEntries[i]

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
		day := prev.Date.Day()
		key := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		monthlyRates[key] = append(monthlyRates[key], rateWithDay{day: day, rate: dailyRate})
	}

	// Create graph points for each month
	for key, rates := range monthlyRates {
		var t time.Time
		t, _ = time.Parse("2006-01", key)
		year := t.Year()
		month := int(t.Month())

		graphPoints := []MonthlyDataPoint{}

		if len(rates) == 1 {
			// Only one data point: create graph points on days 10 and 20
			rate := rates[0].rate
			recordedDay := rates[0].day

			// Day 10: use exact rate if recorded day matches, otherwise use the rate
			if recordedDay == 10 {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   10,
					Rate:  rate,
				})
			} else {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   10,
					Rate:  rate, // Use the single rate
				})
			}

			// Day 20: use exact rate if recorded day matches, otherwise use the rate
			if recordedDay == 20 {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   20,
					Rate:  rate,
				})
			} else {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   20,
					Rate:  rate, // Use the single rate
				})
			}
		} else {
			// Multiple data points: create graph points on days 8, 16, 24
			targetDays := []int{8, 16, 24}

			for _, targetDay := range targetDays {
				// Check if any recorded day matches the target day
				var matchedRate *float64
				for _, r := range rates {
					if r.day == targetDay {
						matchedRate = &r.rate
						break
					}
				}

				if matchedRate != nil {
					// Use exact rate if day matches
					graphPoints = append(graphPoints, MonthlyDataPoint{
						Year:  year,
						Month: month,
						Day:   targetDay,
						Rate:  *matchedRate,
					})
				} else {
					// Average all rates for this month
					sum := 0.0
					for _, r := range rates {
						sum += r.rate
					}
					avgRate := sum / float64(len(rates))
					graphPoints = append(graphPoints, MonthlyDataPoint{
						Year:  year,
						Month: month,
						Day:   targetDay,
						Rate:  avgRate,
					})
				}
			}
		}

		// Add all graph points for this month
		for _, point := range graphPoints {
			result[year] = append(result[year], point)
		}
	}

	// Sort by month and day for each year
	for year := range result {
		sort.Slice(result[year], func(i, j int) bool {
			if result[year][i].Month != result[year][j].Month {
				return result[year][i].Month < result[year][j].Month
			}
			return result[year][i].Day < result[year][j].Day
		})
	}

	return result
}

// aggregateWater aggregates water data by month
// Creates graph points on specific days: 8, 16, 24 if multiple points, or 10, 20 if only one
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
	// "year-month" -> []rateWithDay (all rate calculations for that month)
	monthlyRates := make(map[string][]rateWithDay)

	// Collect all rate calculations grouped by month
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
		day := prev.Date.Day()
		key := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		monthlyRates[key] = append(monthlyRates[key], rateWithDay{day: day, rate: dailyRate})
	}

	// Create graph points for each month
	for key, rates := range monthlyRates {
		var t time.Time
		t, _ = time.Parse("2006-01", key)
		year := t.Year()
		month := int(t.Month())

		graphPoints := []MonthlyDataPoint{}

		if len(rates) == 1 {
			// Only one data point: create graph points on days 10 and 20
			rate := rates[0].rate
			recordedDay := rates[0].day

			// Day 10: use exact rate if recorded day matches, otherwise use the rate
			if recordedDay == 10 {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   10,
					Rate:  rate,
				})
			} else {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   10,
					Rate:  rate, // Use the single rate
				})
			}

			// Day 20: use exact rate if recorded day matches, otherwise use the rate
			if recordedDay == 20 {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   20,
					Rate:  rate,
				})
			} else {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   20,
					Rate:  rate, // Use the single rate
				})
			}
		} else {
			// Multiple data points: create graph points on days 8, 16, 24
			targetDays := []int{8, 16, 24}

			for _, targetDay := range targetDays {
				// Check if any recorded day matches the target day
				var matchedRate *float64
				for _, r := range rates {
					if r.day == targetDay {
						matchedRate = &r.rate
						break
					}
				}

				if matchedRate != nil {
					// Use exact rate if day matches
					graphPoints = append(graphPoints, MonthlyDataPoint{
						Year:  year,
						Month: month,
						Day:   targetDay,
						Rate:  *matchedRate,
					})
				} else {
					// Average all rates for this month
					sum := 0.0
					for _, r := range rates {
						sum += r.rate
					}
					avgRate := sum / float64(len(rates))
					graphPoints = append(graphPoints, MonthlyDataPoint{
						Year:  year,
						Month: month,
						Day:   targetDay,
						Rate:  avgRate,
					})
				}
			}
		}

		// Add all graph points for this month
		for _, point := range graphPoints {
			result[year] = append(result[year], point)
		}
	}

	// Sort by month and day for each year
	for year := range result {
		sort.Slice(result[year], func(i, j int) bool {
			if result[year][i].Month != result[year][j].Month {
				return result[year][i].Month < result[year][j].Month
			}
			return result[year][i].Day < result[year][j].Day
		})
	}

	return result
}

// aggregateFuel aggregates fuel data by month
// Groups data by calendar year (January to December)
// Creates graph points on specific days: 8, 16, 24 if multiple points, or 10, 20 if only one
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
	// "year-month" -> []rateWithDay (all rate calculations for that month)
	monthlyRates := make(map[string][]rateWithDay)

	// Collect all rate calculations grouped by month
	for i := 1; i < len(fuelEntries); i++ {
		prev := fuelEntries[i-1]
		curr := fuelEntries[i]

		days := curr.Date.Sub(prev.Date).Hours() / 24
		if days <= 0 {
			continue
		}

		delta := prev.Gasoline - curr.Gasoline // positive = consumption, negative = refuel
		if delta <= 0 {
			continue // refuel or no change: only graph consumption (decreases)
		}
		dailyRate := delta / days

		// Assign rate to the previous entry's month (where the consumption period started)
		year := prev.Date.Year()
		month := int(prev.Date.Month())
		day := prev.Date.Day()
		key := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		monthlyRates[key] = append(monthlyRates[key], rateWithDay{day: day, rate: dailyRate})
	}

	// Create graph points for each month
	for key, rates := range monthlyRates {
		var t time.Time
		t, _ = time.Parse("2006-01", key)
		year := t.Year()
		month := int(t.Month())

		graphPoints := []MonthlyDataPoint{}

		if len(rates) == 1 {
			// Only one data point: create graph points on days 10 and 20
			rate := rates[0].rate
			recordedDay := rates[0].day

			// Day 10: use exact rate if recorded day matches, otherwise use the rate
			if recordedDay == 10 {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   10,
					Rate:  rate,
				})
			} else {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   10,
					Rate:  rate, // Use the single rate
				})
			}

			// Day 20: use exact rate if recorded day matches, otherwise use the rate
			if recordedDay == 20 {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   20,
					Rate:  rate,
				})
			} else {
				graphPoints = append(graphPoints, MonthlyDataPoint{
					Year:  year,
					Month: month,
					Day:   20,
					Rate:  rate, // Use the single rate
				})
			}
		} else {
			// Multiple data points: create graph points on days 8, 16, 24
			targetDays := []int{8, 16, 24}

			for _, targetDay := range targetDays {
				// Check if any recorded day matches the target day
				var matchedRate *float64
				for _, r := range rates {
					if r.day == targetDay {
						matchedRate = &r.rate
						break
					}
				}

				if matchedRate != nil {
					// Use exact rate if day matches
					graphPoints = append(graphPoints, MonthlyDataPoint{
						Year:  year,
						Month: month,
						Day:   targetDay,
						Rate:  *matchedRate,
					})
				} else {
					// Average all rates for this month
					sum := 0.0
					for _, r := range rates {
						sum += r.rate
					}
					avgRate := sum / float64(len(rates))
					graphPoints = append(graphPoints, MonthlyDataPoint{
						Year:  year,
						Month: month,
						Day:   targetDay,
						Rate:  avgRate,
					})
				}
			}
		}

		// Add all graph points for this month
		for _, point := range graphPoints {
			result[year] = append(result[year], point)
		}
	}

	// Sort by month and day for each year
	for year := range result {
		sort.Slice(result[year], func(i, j int) bool {
			if result[year][i].Month != result[year][j].Month {
				return result[year][i].Month < result[year][j].Month
			}
			return result[year][i].Day < result[year][j].Day
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

// Heating period for fuel projection: September (month 9) to June (month 6) of next year.
// heatingPeriodStart(2024) = Sept 1, 2024; heatingPeriodEnd(2024) = June 30, 2025.
func heatingPeriodStart(year int) time.Time {
	return time.Date(year, time.September, 1, 0, 0, 0, 0, time.UTC)
}
func heatingPeriodEnd(year int) time.Time {
	return time.Date(year+1, time.June, 30, 23, 59, 59, 999999999, time.UTC)
}

// fuelConsumptionInPeriod returns total fuel consumption (L) in [start, end] from consecutive readings.
// Each positive delta (prev.Gasoline - curr.Gasoline) is attributed to the period where prev.Date falls.
func fuelConsumptionInPeriod(entries []models.ConsumptionEntry, start, end time.Time) float64 {
	fuelEntries := []models.ConsumptionEntry{}
	for _, e := range entries {
		if e.Category == "fuel" {
			fuelEntries = append(fuelEntries, e)
		}
	}
	sort.Slice(fuelEntries, func(i, j int) bool {
		return fuelEntries[i].Date.Before(fuelEntries[j].Date)
	})
	var total float64
	for i := 1; i < len(fuelEntries); i++ {
		prev := fuelEntries[i-1]
		curr := fuelEntries[i]
		delta := prev.Gasoline - curr.Gasoline
		if delta <= 0 {
			continue
		}
		// Attribute consumption to the period of the start of the interval
		if !prev.Date.Before(start) && !prev.Date.After(end) {
			total += delta
		}
	}
	return total
}

// getFuelHeatingProjection returns projection for the current heating period (Sept–June), fuel only.
func getFuelHeatingProjection(entries []models.ConsumptionEntry, now time.Time) *FuelHeatingProjection {
	// Current heating period: Sept (this year) to June (next year). If we're in Jan–June, period started last Sept.
	var currentYear int
	if now.Month() >= time.September {
		currentYear = now.Year()
	} else {
		currentYear = now.Year() - 1
	}
	start := heatingPeriodStart(currentYear)
	end := heatingPeriodEnd(currentYear)

	consumedSoFar := fuelConsumptionInPeriod(entries, start, end)

	// Historical complete periods: only past periods that have ended (before today)
	var historicalTotals []float64
	for y := currentYear - 1; y >= currentYear-10; y-- {
		pe := heatingPeriodEnd(y)
		if !pe.Before(now) {
			continue // period not yet ended, skip
		}
		ps := heatingPeriodStart(y)
		total := fuelConsumptionInPeriod(entries, ps, pe)
		if total > 0 {
			historicalTotals = append(historicalTotals, total)
		}
	}

	var historicalAvg float64
	n := len(historicalTotals)
	if n > 0 {
		sum := 0.0
		for _, t := range historicalTotals {
			sum += t
		}
		historicalAvg = sum / float64(n)
	}

	// Remaining full months from next month through June (inclusive)
	var remainingMonths float64
	if now.After(end) {
		remainingMonths = 0
	} else if now.Before(start) {
		remainingMonths = 10 // full period ahead (Sept–June)
	} else if now.Month() >= time.September {
		remainingMonths = float64(18 - int(now.Month())) // e.g. Sept -> 9 months left (Oct..Jun)
	} else {
		remainingMonths = float64(6 - int(now.Month())) // Jan->5, Feb->4, ..., Jun->0
	}

	projectedTotal := consumedSoFar
	if n > 0 && remainingMonths > 0 {
		avgMonthly := historicalAvg / 10.0 // 10 months Sept–June
		projectedTotal = consumedSoFar + avgMonthly*remainingMonths
	}

	nextYearShort := (currentYear + 1) % 100
	periodLabel := strconv.Itoa(currentYear) + "-" + strconv.Itoa(nextYearShort)
	if nextYearShort < 10 {
		periodLabel = strconv.Itoa(currentYear) + "-0" + strconv.Itoa(nextYearShort)
	}

	return &FuelHeatingProjection{
		PeriodLabel:           periodLabel,
		PeriodStart:           start,
		PeriodEnd:             end,
		ConsumedSoFar:         consumedSoFar,
		HistoricalAvgTotal:    historicalAvg,
		HistoricalPeriodsUsed: n,
		ProjectedTotal:        projectedTotal,
		RemainingMonths:       remainingMonths,
		HasProjection:         n > 0,
	}
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

				// Calculate day and night separately
				dayDelta := curr.ElectricityDay - prev.ElectricityDay
				nightDelta := curr.ElectricityNight - prev.ElectricityNight
				dayDailyRate := dayDelta / days
				nightDailyRate := nightDelta / days

				result = append(result, LatestDataPoint{
					Date:                  curr.Date,
					Value:                 currTotal,
					DailyConsumption:      dailyRate,
					DayValue:              curr.ElectricityDay,
					NightValue:            curr.ElectricityNight,
					DayDailyConsumption:   dayDailyRate,
					NightDailyConsumption: nightDailyRate,
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

		// Calculate day and night separately
		dayDelta := curr.ElectricityDay - prev.ElectricityDay
		nightDelta := curr.ElectricityNight - prev.ElectricityNight
		dayDailyRate := dayDelta / days
		nightDailyRate := nightDelta / days

		result = append(result, LatestDataPoint{
			Date:                  curr.Date,
			Value:                 currTotal,
			DailyConsumption:      dailyRate,
			DayValue:              curr.ElectricityDay,
			NightValue:            curr.ElectricityNight,
			DayDailyConsumption:   dayDailyRate,
			NightDailyConsumption: nightDailyRate,
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

				delta := prev.Gasoline - curr.Gasoline // positive = consumption
				if delta <= 0 {
					continue // refuel: skip for latest consumption list
				}
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

	// Get the last 5 consumption intervals (skip refuels: only delta > 0)
	allConsumption := []LatestDataPoint{}
	for i := 1; i < len(fuelEntries); i++ {
		curr := fuelEntries[i]
		prev := fuelEntries[i-1]

		days := curr.Date.Sub(prev.Date).Hours() / 24
		if days <= 0 {
			continue
		}

		delta := prev.Gasoline - curr.Gasoline // positive = consumption
		if delta <= 0 {
			continue // refuel: skip
		}
		dailyRate := delta / days

		allConsumption = append(allConsumption, LatestDataPoint{
			Date:             curr.Date,
			Value:            curr.Gasoline,
			DailyConsumption: dailyRate,
		})
	}
	// Take the last 5 consumption data points
	start := 0
	if len(allConsumption) > 5 {
		start = len(allConsumption) - 5
	}
	result = append(result, allConsumption[start:]...)

	// Reverse to show most recent first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// getLatestEntries returns the latest entry for each category
func getLatestEntries(entries []models.ConsumptionEntry) map[string]models.ConsumptionEntry {
	result := make(map[string]models.ConsumptionEntry)

	// Find latest entry for each category
	for _, entry := range entries {
		if existing, ok := result[entry.Category]; !ok || entry.Date.After(existing.Date) {
			result[entry.Category] = entry
		}
	}

	return result
}

// getLast10Entries returns the last 10 entries sorted by date (most recent first)
func getLast10Entries(entries []models.ConsumptionEntry) []models.ConsumptionEntry {
	// Create a copy to avoid modifying the original slice
	sorted := make([]models.ConsumptionEntry, len(entries))
	copy(sorted, entries)

	// Sort by date descending (most recent first)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.After(sorted[j].Date)
	})

	// Return last 10 (or all if less than 10)
	if len(sorted) > 10 {
		return sorted[:10]
	}
	return sorted
}
