package analytics

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vanam-san/fittrack/internal/database"
	"github.com/vanam-san/fittrack/internal/nutrition"
	"github.com/vanam-san/fittrack/internal/weight"
)

func TestReportReconcilesRecordsAcrossPeriodsAndAfterCorrections(t *testing.T) {
	ctx := t.Context()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "fittrack.sqlite3"))
	check(t, err)
	defer db.Close()
	journal := nutrition.NewService(db)
	entry := nutrition.Entry{
		LogPlacement: nutrition.LogPlacement{Date: "2026-11-29", Category: "meal", Quantity: 1.5},
		Food: nutrition.Food{Name: "Yogurt", Serving: "100 g",
			Nutrients: nutrition.Nutrients{Calories: 120, Protein: 9, Carbohydrates: 8, Fat: 4}},
	}
	mealID, err := journal.SaveEntry(ctx, entry)
	check(t, err)
	entry.Food = nutrition.Food{Name: "Milk", Serving: "1 glass", Nutrients: nutrition.Nutrients{Calories: 100, Protein: 7}}
	entry.Date, entry.Category, entry.Quantity = "2026-11-30", "drink", 2
	_, err = journal.SaveEntry(ctx, entry)
	check(t, err)
	entry.Date, entry.Category, entry.Quantity = "2026-12-01", "snack", 1
	_, err = journal.SaveEntry(ctx, entry)
	check(t, err)
	check(t, weight.NewService(db).Save(ctx, weight.Input{Date: "2026-11-30", Value: weight.Value{Weight: 81, Unit: "kg"}}))

	// Include a skipped set, an unfinished workout and a workout outside the range.
	_, err = db.ExecContext(ctx, `
		INSERT INTO workout_sessions (id, routine_name, session_date, started_at, status, resumed_at, elapsed_ms) VALUES
			(1, 'Strength', '2026-11-29', 0, 'completed', 0, 120000),
			(2, 'Strength', '2026-11-30', 0, 'completed', 0, 180000),
			(3, 'Unfinished', '2026-12-01', 0, 'active', 0, 90000),
			(4, 'Outside range', '2026-11-28', 0, 'completed', 0, 240000);
		INSERT INTO session_exercises (id, session_id, position, exercise_id, exercise_name, tracking, planned_sets, planned_reps, rest_seconds) VALUES
			(1, 1, 1, 1, 'Barbell Squat', 'weight_reps', 3, 5, 60),
			(2, 2, 1, 1, 'Barbell Squat', 'weight_reps', 1, 5, 60),
			(3, 3, 1, 1, 'Barbell Squat', 'weight_reps', 1, 5, 60),
			(4, 2, 2, 12, 'Running', 'distance_duration', 1, 0, 0),
			(5, 2, 3, 9, 'Push-Up', 'reps', 1, 10, 0),
			(6, 2, 4, 11, 'Plank', 'duration', 1, 0, 0);
		INSERT INTO workout_sets (session_exercise_id, position, weight_kg, reps, duration_seconds, distance_m, done) VALUES
			(1, 1, 50, 6, 0, 0, 1), (1, 2, 55, 4, 0, 0, 1), (1, 3, 500, 10, 0, 0, 0),
			(2, 1, 60, 5, 0, 0, 1), (3, 1, 90, 10, 0, 0, 1),
			(4, 1, 0, 0, 120, 400, 1), (5, 1, 0, 15, 0, 0, 1), (6, 1, 0, 0, 60, 0, 1);
	`)
	check(t, err)

	service := NewService(db)
	for _, test := range []struct {
		grouping string
		starts   []string
		ends     []string
		calories []float64
	}{
		{"daily", []string{"2026-11-29", "2026-11-30", "2026-12-01", "2026-12-02"}, []string{"2026-11-29", "2026-11-30", "2026-12-01", "2026-12-02"}, []float64{180, 200, 100, 0}},
		{"weekly", []string{"2026-11-29", "2026-11-30"}, []string{"2026-11-29", "2026-12-02"}, []float64{180, 300}},
		{"monthly", []string{"2026-11-29", "2026-12-01"}, []string{"2026-11-30", "2026-12-02"}, []float64{380, 100}},
	} {
		t.Run(test.grouping, func(t *testing.T) {
			report, err := service.Report(t.Context(), "2026-11-29", "2026-12-02", test.grouping)
			check(t, err)
			if len(report.Periods) != len(test.starts) {
				t.Fatalf("unexpected periods: %+v", report.Periods)
			}
			for index, period := range report.Periods {
				if period.From != test.starts[index] || period.To != test.ends[index] || period.Nutrition.Calories != test.calories[index] {
					t.Fatalf("unexpected period %d: %+v", index, period)
				}
				if test.grouping == "daily" {
					day, err := journal.Day(t.Context(), period.From)
					check(t, err)
					if day.Total != period.Nutrition {
						t.Fatalf("analytics differed from journal on %s: %+v vs %+v", period.From, period.Nutrition, day.Total)
					}
				}
			}
			want := nutrition.Nutrients{Calories: 480, Protein: 34.5, Carbohydrates: 12, Fat: 6}
			if report.Total.Nutrition != want || report.Total.LoggedDays != 3 ||
				report.Total.Workouts != 2 || report.Total.ElapsedMS != 300000 || report.Total.VolumeKg != 820 ||
				len(report.Measurements) != 1 || report.Measurements[0].WeightKg != 81 || len(report.Exercises) != 4 {
				t.Fatalf("unexpected report: %+v", report)
			}
			squat := report.Exercises[0]
			expectedDays := []ExerciseDay{
				{Date: "2026-11-29", WeightKg: 55, Reps: 6, VolumeKg: 520},
				{Date: "2026-11-30", WeightKg: 60, Reps: 5, VolumeKg: 300},
			}
			expectedRecords := []Record{
				{Metric: "weightKg", Date: "2026-11-30", Value: 60},
				{Metric: "volumeKg", Date: "2026-11-29", Value: 520},
				{Metric: "reps", Date: "2026-11-29", Value: 6},
			}
			if squat.Name != "Barbell Squat" || !reflect.DeepEqual(squat.Days, expectedDays) || !reflect.DeepEqual(squat.Records, expectedRecords) {
				t.Fatalf("unexpected squat progression: %+v", squat)
			}
			for i, records := range [][]Record{
				{{Metric: "durationSeconds", Date: "2026-11-30", Value: 60}},
				{{Metric: "reps", Date: "2026-11-30", Value: 15}},
				{{Metric: "distanceM", Date: "2026-11-30", Value: 400}, {Metric: "durationSeconds", Date: "2026-11-30", Value: 120}},
			} {
				if !reflect.DeepEqual(report.Exercises[i+1].Records, records) {
					t.Fatalf("unexpected records for %s: %+v", report.Exercises[i+1].Name, report.Exercises[i+1].Records)
				}
			}
		})
	}

	entry.ID = mealID
	_, err = journal.SaveEntry(ctx, entry)
	check(t, err)
	report, err := service.Report(ctx, "2026-11-29", "2026-12-02", "daily")
	check(t, err)
	if report.Total.Nutrition.Calories != 400 || report.Periods[0].LoggedDays != 0 || report.Periods[2].Nutrition.Calories != 200 {
		t.Fatalf("edited food was not moved and recalculated: %+v", report.Periods)
	}
	check(t, journal.DeleteEntry(ctx, mealID))
	check(t, weight.NewService(db).Delete(ctx, report.Measurements[0].ID))
	_, err = db.ExecContext(ctx, "UPDATE workout_sets SET reps = 3 WHERE session_exercise_id = 1 AND position = 1; DELETE FROM workout_sessions WHERE id = 2")
	check(t, err)
	report, err = service.Report(ctx, "2026-11-29", "2026-12-02", "weekly")
	check(t, err)
	if report.Total.Nutrition.Calories != 300 || report.Total.Workouts != 1 || report.Total.ElapsedMS != 120000 ||
		report.Total.VolumeKg != 370 || len(report.Measurements) != 0 || len(report.Exercises) != 1 ||
		report.Exercises[0].Records[0].Value != 55 || report.Exercises[0].Records[2].Value != 4 {
		t.Fatalf("analytics retained outdated values: %+v", report)
	}
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
