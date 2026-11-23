package nutrition

import (
	"math"
	"path/filepath"
	"slices"
	"testing"

	"github.com/vanam-san/fittrack/internal/database"
)

func require(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPresetSnapshotsAndJournalTotals(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "fittrack.sqlite3")
	db, err := database.Open(ctx, path)
	require(t, err)
	defer db.Close()
	service := NewService(db)
	food := Food{Name: "Greek yogurt", Serving: "100 g", Nutrients: Nutrients{100, 8, 7, 4}}
	food.ID, err = service.SavePreset(ctx, food)
	require(t, err)

	first := LogPlacement{Date: "2026-12-31", Category: "snack", Quantity: 1.5}
	second := LogPlacement{Date: "2027-01-01", Category: "meal", Quantity: 2}
	firstID, err := service.LogPreset(ctx, food.ID, first)
	require(t, err)
	_, err = service.LogPreset(ctx, food.ID, second)
	require(t, err)
	waterID, err := service.SaveEntry(ctx, Entry{
		LogPlacement: LogPlacement{Date: first.Date, Category: "drink", Quantity: 1},
		Food:         Food{Name: "Juice", Serving: "1 cup", Nutrients: Nutrients{120, 0, 30, 0}},
	})
	require(t, err)

	food.Nutrients = Nutrients{200, 18, 10, 8}
	food.Serving = "200 g"
	_, err = service.SavePreset(ctx, food)
	require(t, err)
	newID, err := service.LogPreset(ctx, food.ID, first)
	require(t, err)
	require(t, db.Close())

	reopened, err := database.Open(ctx, path)
	require(t, err)
	defer reopened.Close()
	service = NewService(reopened)
	day, err := service.Day(ctx, first.Date)
	require(t, err)
	if day.Total != (Nutrients{570, 39, 55.5, 18}) ||
		day.Groups[1].Total != (Nutrients{450, 39, 25.5, 18}) ||
		day.Groups[1].Entries[0].ID != firstID ||
		day.Groups[1].Entries[0].Food.Serving != "100 g" ||
		day.Groups[1].Entries[1].Food.Serving != "200 g" {
		t.Fatalf("incorrect first day or preset snapshots: %+v", day)
	}
	other, err := service.Day(ctx, second.Date)
	require(t, err)
	if other.Total != (Nutrients{200, 16, 14, 8}) {
		t.Fatalf("preset edit affected second date: %+v", other)
	}
	require(t, service.DeletePreset(ctx, food.ID))

	entry := day.Groups[1].Entries[0]
	entry.Date, entry.Category, entry.Quantity = second.Date, "drink", 0.5
	_, err = service.SaveEntry(ctx, entry)
	require(t, err)
	require(t, service.DeleteEntry(ctx, waterID))
	require(t, service.DeleteEntry(ctx, newID))
	day, err = service.Day(ctx, first.Date)
	require(t, err)
	other, err = service.Day(ctx, second.Date)
	require(t, err)
	dates, err := service.Dates(ctx)
	require(t, err)
	if day.Total != (Nutrients{}) || other.Total != (Nutrients{250, 20, 17.5, 10}) ||
		!slices.Equal(dates, []string{second.Date}) {
		t.Fatalf("history edits did not reconcile totals and dates: %+v, %+v, %v", day, other, dates)
	}
}

func TestValidation(t *testing.T) {
	for _, date := range []string{"2025-02-29", "0000-01-01", "2026-11-02T01:00:00Z"} {
		if err := validateDate(date); err == nil {
			t.Errorf("accepted invalid calendar date %q", date)
		}
	}
	if err := validateDate("2024-02-29"); err != nil {
		t.Fatal(err)
	}
	if err := validatePlacement(LogPlacement{Date: "2026-11-23", Category: "meal", Quantity: math.Inf(1)}); err == nil {
		t.Fatal("accepted a non-finite serving quantity")
	}
	if err := validateFood(&Food{Name: "Meal", Serving: "1 bowl", Nutrients: Nutrients{Protein: math.NaN()}}); err == nil {
		t.Fatal("accepted a non-finite nutrient value")
	}
}
