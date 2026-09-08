package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/virer/konsumo/models"
	"github.com/virer/konsumo/storage"
)

func TestMain(m *testing.M) {
	// Tests run from package web/; module root is parent directory.
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func setupTestStore(t *testing.T) {
	orig := storage.DefaultStore()
	t.Cleanup(func() { storage.SetStore(orig) })
	dir := t.TempDir()
	storage.SetStore(&storage.Store{FilePath: filepath.Join(dir, "consumption.json")})
}

func date(y int, m int, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"123.45", 123.45},
		{"0", 0},
		{"", 0},
		{"invalid", 0},
	}
	for _, tt := range tests {
		got := parseFloat(tt.input)
		if got != tt.want {
			t.Errorf("parseFloat(%q) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestGetHeatingYear(t *testing.T) {
	tests := []struct {
		date time.Time
		want int
	}{
		{date(2024, 8, 1), 2024},
		{date(2024, 9, 1), 2024},
		{date(2024, 12, 31), 2024},
		{date(2025, 1, 1), 2024},
		{date(2025, 7, 31), 2024},
	}
	for _, tt := range tests {
		got := getHeatingYear(tt.date)
		if got != tt.want {
			t.Errorf("getHeatingYear(%v) = %d, want %d", tt.date, got, tt.want)
		}
	}
}

func TestGetHeatingMonth(t *testing.T) {
	tests := []struct {
		month int
		want  int
	}{
		{8, 1},
		{9, 2},
		{12, 5},
		{1, 6},
		{7, 12},
	}
	for _, tt := range tests {
		got := getHeatingMonth(tt.month)
		if got != tt.want {
			t.Errorf("getHeatingMonth(%d) = %d, want %d", tt.month, got, tt.want)
		}
	}
}

func TestHeatingPeriodBounds(t *testing.T) {
	start := heatingPeriodStart(2024)
	end := heatingPeriodEnd(2024)
	if start != date(2024, 9, 1) {
		t.Errorf("heatingPeriodStart(2024) = %v", start)
	}
	if end.Year() != 2025 || end.Month() != time.June || end.Day() != 30 {
		t.Errorf("heatingPeriodEnd(2024) = %v", end)
	}
}

func TestFuelConsumptionInPeriod(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 9, 1), Category: "fuel", Gasoline: 500},
		{Date: date(2024, 10, 1), Category: "fuel", Gasoline: 450},
		{Date: date(2024, 11, 1), Category: "fuel", Gasoline: 600}, // refuel
		{Date: date(2024, 12, 1), Category: "fuel", Gasoline: 550},
	}
	start := date(2024, 9, 1)
	end := date(2024, 12, 31)

	got := fuelConsumptionInPeriod(entries, start, end)
	// Sept: 500-450=50, Oct: skipped (refuel), Nov: 600-550=50
	if got != 100 {
		t.Errorf("fuelConsumptionInPeriod = %f, want 100", got)
	}
}

func TestFuelConsumptionInPeriodOutsideRange(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2023, 1, 1), Category: "fuel", Gasoline: 500},
		{Date: date(2023, 2, 1), Category: "fuel", Gasoline: 400},
	}
	got := fuelConsumptionInPeriod(entries, date(2024, 1, 1), date(2024, 12, 31))
	if got != 0 {
		t.Errorf("expected 0 outside period, got %f", got)
	}
}

func TestWaterConsumptionInPeriod(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "water", Water: 100},
		{Date: date(2024, 2, 1), Category: "water", Water: 110},
		{Date: date(2024, 3, 1), Category: "water", Water: 105}, // negative delta skipped
	}
	total, hasData := waterConsumptionInPeriod(entries, date(2024, 1, 1), date(2024, 3, 31))
	if !hasData {
		t.Fatal("expected hasData true")
	}
	if total != 10 {
		t.Errorf("waterConsumptionInPeriod = %f, want 10", total)
	}
}

func TestElectricityPeriodTotalsSorted(t *testing.T) {
	elec := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "electricity", ElectricityDay: 100, ElectricityNight: 50},
		{Date: date(2024, 2, 1), Category: "electricity", ElectricityDay: 130, ElectricityNight: 60},
		{Date: date(2024, 3, 1), Category: "electricity", ElectricityDay: 140, ElectricityNight: 70},
	}
	total, day, night, hasData := electricityPeriodTotalsSorted(elec, date(2024, 1, 1), date(2024, 2, 28))
	if !hasData {
		t.Fatal("expected hasData true")
	}
	// Jan->Feb and Feb->Mar intervals both attributed via prev.Date in range
	if total != 60 || day != 40 || night != 20 {
		t.Errorf("got total=%f day=%f night=%f, want 60/40/20", total, day, night)
	}
}

func TestGetFuelHeatingProjection(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2023, 9, 1), Category: "fuel", Gasoline: 500},
		{Date: date(2024, 6, 1), Category: "fuel", Gasoline: 200},
		{Date: date(2024, 9, 1), Category: "fuel", Gasoline: 500},
		{Date: date(2024, 11, 1), Category: "fuel", Gasoline: 450},
	}
	now := date(2024, 12, 15)
	proj := getFuelHeatingProjection(entries, now)
	if proj == nil {
		t.Fatal("expected projection")
	}
	if proj.ConsumedSoFar != 50 {
		t.Errorf("ConsumedSoFar = %f, want 50", proj.ConsumedSoFar)
	}
	if !proj.HasProjection {
		t.Error("expected HasProjection true with historical data")
	}
	if proj.PeriodLabel != "2024-25" {
		t.Errorf("PeriodLabel = %q, want 2024-25", proj.PeriodLabel)
	}
}

func TestGetElectricityYearSummary(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2023, 7, 1), Category: "electricity", ElectricityDay: 100, ElectricityNight: 50},
		{Date: date(2024, 6, 1), Category: "electricity", ElectricityDay: 200, ElectricityNight: 100},
	}
	summary := getElectricityYearSummary(entries)
	if summary == nil {
		t.Fatal("expected summary")
	}
	if summary.Last12MonthsTotal <= 0 {
		t.Errorf("expected positive total, got %f", summary.Last12MonthsTotal)
	}
	if len(summary.PastYears) == 0 {
		t.Error("expected at least one past year period")
	}
}

func TestGetElectricityYearSummaryInsufficientData(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "electricity", ElectricityDay: 100, ElectricityNight: 50},
	}
	if getElectricityYearSummary(entries) != nil {
		t.Error("expected nil with fewer than 2 entries")
	}
}

func TestGetWaterYearSummary(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2023, 7, 1), Category: "water", Water: 100},
		{Date: date(2024, 6, 1), Category: "water", Water: 200},
	}
	summary := getWaterYearSummary(entries)
	if summary == nil {
		t.Fatal("expected summary")
	}
	if summary.Last12MonthsTotal <= 0 {
		t.Errorf("expected positive total, got %f", summary.Last12MonthsTotal)
	}
}

func TestGetWaterYearSummaryInsufficientData(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "water", Water: 100},
	}
	if getWaterYearSummary(entries) != nil {
		t.Error("expected nil with fewer than 2 entries")
	}
}

func TestGetLatestElectricity(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "electricity", ElectricityDay: 100, ElectricityNight: 50},
		{Date: date(2024, 1, 11), Category: "electricity", ElectricityDay: 110, ElectricityNight: 55},
	}
	latest := getLatestElectricity(entries)
	if len(latest) != 1 {
		t.Fatalf("expected 1 latest point, got %d", len(latest))
	}
	if latest[0].DailyConsumption != 1.5 { // 15 kWh / 10 days
		t.Errorf("DailyConsumption = %f, want 1.5", latest[0].DailyConsumption)
	}
}

func TestGetLatestWater(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "water", Water: 100},
		{Date: date(2024, 1, 11), Category: "water", Water: 110},
	}
	latest := getLatestWater(entries)
	if len(latest) != 1 {
		t.Fatalf("expected 1 latest point, got %d", len(latest))
	}
	if latest[0].DailyConsumption != 1.0 {
		t.Errorf("DailyConsumption = %f, want 1.0", latest[0].DailyConsumption)
	}
}

func TestGetLatestFuelSkipsRefuel(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "fuel", Gasoline: 400},
		{Date: date(2024, 1, 11), Category: "fuel", Gasoline: 500}, // refuel
		{Date: date(2024, 1, 21), Category: "fuel", Gasoline: 480},
	}
	latest := getLatestFuel(entries)
	if len(latest) != 1 {
		t.Fatalf("expected 1 latest point (refuel skipped), got %d", len(latest))
	}
	if latest[0].DailyConsumption != 2.0 { // 20L / 10 days
		t.Errorf("DailyConsumption = %f, want 2.0", latest[0].DailyConsumption)
	}
}

func TestGetLatestEntries(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "water", Water: 100},
		{Date: date(2024, 2, 1), Category: "water", Water: 110},
		{Date: date(2024, 1, 15), Category: "electricity", ElectricityDay: 50, ElectricityNight: 25},
	}
	latest := getLatestEntries(entries)
	if latest["water"].Water != 110 {
		t.Errorf("latest water = %f, want 110", latest["water"].Water)
	}
	if latest["electricity"].ElectricityDay != 50 {
		t.Errorf("latest electricity day = %f, want 50", latest["electricity"].ElectricityDay)
	}
}

func TestGetLast10Entries(t *testing.T) {
	entries := make([]models.ConsumptionEntry, 15)
	for i := 0; i < 15; i++ {
		entries[i] = models.ConsumptionEntry{
			Date:     date(2024, 1, i+1),
			Category: "water",
			Water:    float64(i),
		}
	}
	last10 := getLast10Entries(entries)
	if len(last10) != 10 {
		t.Fatalf("expected 10 entries, got %d", len(last10))
	}
	// Most recent first
	if last10[0].Date.Day() != 15 {
		t.Errorf("expected most recent day 15, got %d", last10[0].Date.Day())
	}
}

func TestAggregateElectricitySingleReading(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "electricity", ElectricityDay: 100, ElectricityNight: 0},
		{Date: date(2024, 1, 11), Category: "electricity", ElectricityDay: 110, ElectricityNight: 0},
	}
	result := aggregateElectricity(entries)
	points := result[2024]
	if len(points) != 2 {
		t.Fatalf("expected 2 graph points for single rate month, got %d", len(points))
	}
	for _, p := range points {
		if p.Day != 10 && p.Day != 20 {
			t.Errorf("unexpected day %d for single-rate month", p.Day)
		}
		if p.Rate != 1.0 {
			t.Errorf("expected rate 1.0, got %f", p.Rate)
		}
	}
}

func TestAggregateWaterSameYearOnly(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "water", Water: 100},
		{Date: date(2025, 1, 1), Category: "water", Water: 200},
	}
	result := aggregateWater(entries)
	if len(result) != 0 {
		t.Errorf("expected no points when readings span different years, got %v", result)
	}
}

func TestAggregateFuelSkipsRefuel(t *testing.T) {
	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "fuel", Gasoline: 400},
		{Date: date(2024, 1, 11), Category: "fuel", Gasoline: 500},
		{Date: date(2024, 1, 21), Category: "fuel", Gasoline: 480},
	}
	result := aggregateFuel(entries)
	if len(result) == 0 {
		t.Fatal("expected fuel aggregation data")
	}
	// Only consumption interval (500->480) should produce points
	found := false
	for _, points := range result {
		for _, p := range points {
			if p.Rate == 2.0 {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected rate 2.0 from consumption interval")
	}
}

func TestSubmitHandlerWater(t *testing.T) {
	setupTestStore(t)

	form := url.Values{
		"date":     {"2024-06-15"},
		"category": {"water"},
		"water":    {"123.45"},
	}
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	SubmitHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, "tab=form") || !strings.Contains(loc, "category=water") {
		t.Errorf("unexpected redirect: %s", loc)
	}

	entries, _ := storage.LoadData()
	if len(entries) != 1 || entries[0].Water != 123.45 {
		t.Fatalf("unexpected saved entries: %+v", entries)
	}
}

func TestSubmitHandlerElectricity(t *testing.T) {
	setupTestStore(t)

	form := url.Values{
		"date":             {"2024-06-15"},
		"category":         {"electricity"},
		"electricity_day":  {"100"},
		"electricity_night": {"50"},
	}
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	SubmitHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	entries, _ := storage.LoadData()
	if len(entries) != 1 || entries[0].ElectricityDay != 100 {
		t.Fatalf("unexpected saved entries: %+v", entries)
	}
}

func TestSubmitHandlerFuel(t *testing.T) {
	setupTestStore(t)

	form := url.Values{
		"date":     {"2024-06-15"},
		"category": {"fuel"},
		"gasoline": {"500"},
	}
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	SubmitHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	entries, _ := storage.LoadData()
	if len(entries) != 1 || entries[0].Gasoline != 500 {
		t.Fatalf("unexpected saved entries: %+v", entries)
	}
}

func TestSubmitHandlerInvalidDate(t *testing.T) {
	setupTestStore(t)

	form := url.Values{
		"date":     {"not-a-date"},
		"category": {"water"},
		"water":    {"100"},
	}
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	SubmitHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestSubmitHandlerUnknownCategory(t *testing.T) {
	setupTestStore(t)

	form := url.Values{
		"date":     {"2024-06-15"},
		"category": {"gas"},
	}
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	SubmitHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestDeleteHandlerSuccess(t *testing.T) {
	setupTestStore(t)

	entries := []models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "water", Water: 100},
		{Date: date(2024, 2, 1), Category: "water", Water: 110},
	}
	if err := storage.SaveData(entries); err != nil {
		t.Fatal(err)
	}

	form := url.Values{"index": {"0"}} // most recent (Feb) is index 0
	req := httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	DeleteHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	remaining, _ := storage.LoadData()
	if len(remaining) != 1 || remaining[0].Water != 100 {
		t.Fatalf("unexpected remaining entries: %+v", remaining)
	}
}

func TestDeleteHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/delete", nil)
	rr := httptest.NewRecorder()

	DeleteHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rr.Code)
	}
}

func TestDeleteHandlerInvalidIndex(t *testing.T) {
	setupTestStore(t)

	form := url.Values{"index": {"abc"}}
	req := httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	DeleteHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestDeleteHandlerIndexOutOfRange(t *testing.T) {
	setupTestStore(t)

	if err := storage.SaveData([]models.ConsumptionEntry{
		{Date: date(2024, 1, 1), Category: "water", Water: 100},
	}); err != nil {
		t.Fatal(err)
	}

	form := url.Values{"index": {"5"}}
	req := httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	DeleteHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestHomeHandler(t *testing.T) {
	setupTestStore(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	HomeHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Konsumo") {
		t.Error("expected HTML to contain Konsumo")
	}
	if !strings.Contains(body, "chartData") {
		t.Error("expected HTML to contain chartData")
	}
}
