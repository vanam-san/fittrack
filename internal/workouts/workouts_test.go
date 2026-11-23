package workouts

import (
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/vanam-san/fittrack/internal/database"
)

func require(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorkoutSurvivesRestartAndRoutineChanges(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "fittrack.sqlite3")
	db, err := database.Open(ctx, path)
	require(t, err)
	defer db.Close()

	zone := time.FixedZone("test", 5*3600+30*60)
	now := time.Date(2026, 11, 24, 0, 5, 0, 0, zone)
	service := NewService(db)
	service.now = func() time.Time { return now }
	overview, err := service.Overview(ctx, "kg")
	require(t, err)
	if len(overview.Exercises) == 0 || !overview.Exercises[0].BuiltIn {
		t.Fatal("expected a seeded exercise library")
	}
	press, err := service.CreateExercise(ctx, ExerciseInput{
		Name: "Landmine Press", PrimaryMuscle: "Shoulders", Equipment: "Barbell", Tracking: "weight_reps",
	})
	require(t, err)
	carry, err := service.CreateExercise(ctx, ExerciseInput{
		Name: "Farmer Carry", PrimaryMuscle: "Forearms", Equipment: "Dumbbells", Tracking: "distance_duration",
	})
	require(t, err)
	plans := []Plan{{ExerciseID: press, PlannedSets: 2, PlannedReps: 8, RestSeconds: 90},
		{ExerciseID: carry, PlannedSets: 1, RestSeconds: 45}}
	routineID, err := service.SaveRoutine(ctx, Routine{Name: "Original plan", Exercises: plans})
	require(t, err)
	copyID, err := service.DuplicateRoutine(ctx, routineID)
	require(t, err)
	overview, err = service.Overview(ctx, "kg")
	require(t, err)
	if len(overview.Routines) != 2 || overview.Routines[1].ID != copyID || overview.Routines[1].Exercises[1] != plans[1] {
		t.Fatalf("duplicate lost the ordered plan: %+v", overview.Routines)
	}

	id, err := service.Start(ctx, routineID)
	require(t, err)
	if _, err := service.Start(ctx, copyID); err == nil {
		t.Fatal("expected only one unfinished workout")
	}
	session, err := service.Session(ctx, id, "lb")
	require(t, err)
	if session.Date != "2026-11-24" || len(session.Exercises) != 2 ||
		session.Exercises[0].ExerciseID != press || session.Exercises[1].ExerciseID != carry {
		t.Fatalf("session did not snapshot the local start date or routine order: %+v", session)
	}
	exercise := session.Exercises[0]
	setID := exercise.Sets[0].ID
	input := SetInput{Weight: 100, WeightUnit: "lb", Reps: 9, Done: true, Notes: "First working set"}
	require(t, service.SaveSet(ctx, id, exercise.ID, setID, input))
	require(t, service.AddSet(ctx, id, exercise.ID))
	timed := session.Exercises[1]
	require(t, service.SaveSet(ctx, id, timed.ID, timed.Sets[0].ID,
		SetInput{DurationSeconds: 35, DistanceM: 60, Done: true}))
	now = now.Add(5 * time.Second)
	require(t, service.ChangeStatus(ctx, id, "paused"))
	if err := service.SaveSet(ctx, id, exercise.ID, setID, input); err == nil {
		t.Fatal("paused workout accepted a set edit")
	}
	require(t, db.Close())

	reopened, err := database.Open(ctx, path)
	require(t, err)
	defer reopened.Close()
	service = NewService(reopened)
	now = now.Add(time.Hour)
	service.now = func() time.Time { return now }
	overview, err = service.Overview(ctx, "lb")
	require(t, err)
	recovered := overview.Active
	if recovered == nil || recovered.ID != id || recovered.Status != "paused" ||
		recovered.ElapsedMS != 5000 || recovered.PausedRestMS != 40000 ||
		len(recovered.Exercises[0].Sets) != 3 || recovered.Exercises[0].Sets[0].SetInput != input {
		t.Fatalf("incorrect recovered session: %+v", recovered)
	}
	require(t, service.ChangeStatus(ctx, id, "active"))
	session, err = service.Session(ctx, id, "kg")
	require(t, err)
	if session.RestUntil != now.UnixMilli()+40000 ||
		math.Abs(session.Exercises[0].Sets[0].Weight-45.359237) > 1e-8 {
		t.Fatalf("timer or saved weight did not survive resume: %+v", session)
	}
	now = now.Add(10 * time.Second)
	require(t, service.ChangeStatus(ctx, id, "completed"))
	_, err = service.SaveRoutine(ctx, Routine{ID: routineID, Name: "Updated plan", Exercises: []Plan{
		{ExerciseID: carry, PlannedSets: 4, RestSeconds: 10},
	}})
	require(t, err)
	require(t, service.DeleteRoutine(ctx, routineID))

	history, err := service.Session(ctx, id, "lb")
	require(t, err)
	if history.Name != "Original plan" || history.Status != "completed" || history.ElapsedMS != 15000 ||
		len(history.Exercises) != 2 || history.Exercises[0].Plan != plans[0] ||
		history.Exercises[0].Sets[0].SetInput != input {
		t.Fatalf("routine changes affected workout history: %+v", history)
	}
	if err := service.SaveSet(ctx, id, exercise.ID, setID, input); err == nil {
		t.Fatal("completed workout accepted a set edit")
	}
	_, err = service.Start(ctx, copyID)
	require(t, err)
	overview, err = service.Overview(ctx, "lb")
	require(t, err)
	previous := overview.Active.Exercises[0].Previous
	if len(previous) != 1 || previous[0].SetInput != input {
		t.Fatalf("previous performance did not use completed sets: %+v", previous)
	}
}

func TestValidateSet(t *testing.T) {
	tests := []struct {
		tracking string
		input    SetInput
	}{
		{"weight_reps", SetInput{Weight: math.Inf(1), WeightUnit: "kg", Reps: 5, Done: true}},
		{"reps", SetInput{Reps: 0, Done: true}},
		{"duration", SetInput{DurationSeconds: -1}},
		{"distance_duration", SetInput{DurationSeconds: 120, DistanceM: math.NaN(), Done: true}},
	}
	for _, test := range tests {
		if _, err := validateSet(1, test.tracking, test.input); err == nil {
			t.Errorf("invalid %s set passed validation", test.tracking)
		}
	}
}
