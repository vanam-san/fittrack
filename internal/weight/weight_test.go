package weight

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/vanam-san/fittrack/internal/database"
	"github.com/vanam-san/fittrack/internal/settings"
)

func TestMeasurementsAndGoalSurviveRestartAndUnitChanges(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "fittrack.sqlite3")
	db, err := database.Open(ctx, path)
	check(t, err)
	defer db.Close()
	service := NewService(db)
	empty, err := service.Overview(ctx)
	check(t, err)
	if empty.Latest != nil || empty.GoalKg != nil {
		t.Fatalf("expected an empty overview, got %+v", empty)
	}

	check(t, service.Save(ctx, Input{Date: "2026-11-19", Value: Value{Weight: 180, Unit: "lb"}}))
	check(t, service.Save(ctx, Input{Date: "2026-11-19", Value: Value{Weight: 82.5, Unit: "kg"}}))
	check(t, service.Save(ctx, Input{Date: "2026-11-03", Value: Value{Weight: 86, Unit: "kg"}}))
	check(t, service.SaveGoal(ctx, Value{Weight: 170, Unit: "lb"}))
	preferences := settings.NewService(db)
	for _, unit := range []string{"lb", "kg", "lb"} {
		_, err := preferences.Save(ctx, settings.Preferences{WeightUnit: unit, Theme: "system"})
		check(t, err)
	}
	check(t, db.Close())

	db, err = database.Open(ctx, path)
	check(t, err)
	defer db.Close()
	service = NewService(db)
	overview, err := service.Overview(ctx)
	check(t, err)
	measurements, err := service.Trend(ctx, "2026-11-19", "2026-11-19")
	check(t, err)
	if len(measurements) != 2 || measurements[0].ID >= measurements[1].ID ||
		math.Abs(measurements[0].WeightKg-81.6466266) > 1e-9 ||
		overview.Latest == nil || *overview.Latest != measurements[1] || overview.Latest.WeightKg != 82.5 ||
		overview.GoalKg == nil || math.Abs(*overview.GoalKg-77.1107029) > 1e-9 {
		t.Fatalf("unexpected persisted weight data: overview=%+v measurements=%+v", overview, measurements)
	}

	check(t, service.Save(ctx, Input{ID: measurements[0].ID, Date: "2026-11-20", Value: Value{Weight: 80, Unit: "kg"}}))
	check(t, service.Delete(ctx, measurements[1].ID))
	check(t, service.ClearGoal(ctx))
	overview, err = service.Overview(ctx)
	check(t, err)
	if overview.Latest == nil || overview.Latest.Date != "2026-11-20" || overview.Latest.WeightKg != 80 || overview.GoalKg != nil {
		t.Fatalf("unexpected weight after corrections: %+v", overview)
	}
	measurements, err = service.Trend(ctx, "2026-11-19", "2026-11-19")
	check(t, err)
	if len(measurements) != 0 {
		t.Fatalf("old date still has measurements: %+v", measurements)
	}
}

func TestRejectsInvalidMeasurements(t *testing.T) {
	db, err := database.Open(t.Context(), filepath.Join(t.TempDir(), "fittrack.sqlite3"))
	check(t, err)
	defer db.Close()
	service := NewService(db)
	for _, input := range []Input{
		{Date: "2026-02-29", Value: Value{Weight: 78, Unit: "kg"}},
		{Date: "2026-02-02", Value: Value{Weight: 78, Unit: "stone"}},
		{Date: "2026-02-02", Value: Value{Weight: 0, Unit: "kg"}},
		{Date: "2026-02-02", Value: Value{Weight: math.Inf(1), Unit: "kg"}},
		{ID: 42, Date: "2026-02-02", Value: Value{Weight: 78, Unit: "kg"}},
	} {
		if err := service.Save(t.Context(), input); err == nil {
			t.Fatalf("accepted invalid measurement: %+v", input)
		}
	}
	if _, err := service.Trend(t.Context(), "2026-02-03", "2026-02-02"); err == nil {
		t.Fatal("accepted an inverted date range")
	}
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
