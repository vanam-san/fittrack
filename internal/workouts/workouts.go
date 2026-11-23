package workouts

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/vanam-san/fittrack/internal/database/dbquery"
)

type ExerciseInput struct {
	Name             string `json:"name"`
	PrimaryMuscle    string `json:"primaryMuscle"`
	SecondaryMuscles string `json:"secondaryMuscles"`
	Equipment        string `json:"equipment"`
	Tracking         string `json:"tracking"`
	Notes            string `json:"notes"`
}

type Exercise struct {
	ExerciseInput
	ID      int64 `json:"id"`
	BuiltIn bool  `json:"builtIn"`
}

type Plan struct {
	ExerciseID  int64 `json:"exerciseId"`
	PlannedSets int64 `json:"plannedSets"`
	PlannedReps int64 `json:"plannedReps"`
	RestSeconds int64 `json:"restSeconds"`
}

type Routine struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Exercises []Plan `json:"exercises"`
}

type SetInput struct {
	Weight          float64 `json:"weight"`
	WeightUnit      string  `json:"weightUnit"`
	Reps            int64   `json:"reps"`
	DurationSeconds int64   `json:"durationSeconds"`
	DistanceM       float64 `json:"distanceM"`
	Done            bool    `json:"done"`
	Notes           string  `json:"notes"`
}

type Set struct {
	SetInput
	ID int64 `json:"id"`
}

type SessionExercise struct {
	Plan
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Tracking string `json:"tracking"`
	Sets     []Set  `json:"sets"`
	Previous []Set  `json:"previous"`
}

type SessionSummary struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Date        string `json:"date"`
	Status      string `json:"status"`
	StartedAt   int64  `json:"startedAt"`
	CompletedAt int64  `json:"completedAt"`
	ElapsedMS   int64  `json:"elapsedMs"`
}

type Session struct {
	SessionSummary
	RestUntil    int64             `json:"restUntil"`
	PausedRestMS int64             `json:"pausedRestMs"`
	Exercises    []SessionExercise `json:"exercises"`
}

type Overview struct {
	Exercises []Exercise       `json:"exercises"`
	Routines  []Routine        `json:"routines"`
	History   []SessionSummary `json:"history"`
	Active    *Session         `json:"active"`
}

type inputError string

func (e inputError) Error() string { return string(e) }

type Service struct {
	db  *sql.DB
	q   *dbquery.Queries
	now func() time.Time
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db, q: dbquery.New(db), now: time.Now}
}

func (s *Service) transaction(ctx context.Context, run func(*dbquery.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := run(s.q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) Overview(ctx context.Context, unit string) (Overview, error) {
	result := Overview{Exercises: []Exercise{}, Routines: []Routine{}, History: []SessionSummary{}}
	err := s.transaction(ctx, func(q *dbquery.Queries) error {
		exercises, err := q.ListExercises(ctx)
		if err != nil {
			return err
		}
		for _, exercise := range exercises {
			result.Exercises = append(result.Exercises, Exercise{
				ID: exercise.ID, BuiltIn: exercise.BuiltIn != 0,
				ExerciseInput: ExerciseInput{
					Name: exercise.Name, PrimaryMuscle: exercise.PrimaryMuscle,
					SecondaryMuscles: exercise.SecondaryMuscles, Equipment: exercise.Equipment,
					Tracking: exercise.Tracking, Notes: exercise.Notes,
				},
			})
		}
		routines, err := q.ListRoutines(ctx)
		if err != nil {
			return err
		}
		for _, row := range routines {
			routine, err := readRoutine(ctx, q, row)
			if err != nil {
				return err
			}
			result.Routines = append(result.Routines, routine)
		}
		history, err := q.ListCompletedSessions(ctx)
		if err != nil {
			return err
		}
		now := s.now().UnixMilli()
		for _, row := range history {
			result.History = append(result.History, summarize(row, now))
		}
		open, err := q.GetOpenSession(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		session, err := readSession(ctx, q, open, unit, now)
		result.Active = &session
		return err
	})
	return result, err
}

func (s *Service) CreateExercise(ctx context.Context, input ExerciseInput) (int64, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.PrimaryMuscle = strings.TrimSpace(input.PrimaryMuscle)
	input.SecondaryMuscles = strings.TrimSpace(input.SecondaryMuscles)
	input.Equipment = strings.TrimSpace(input.Equipment)
	input.Notes = strings.TrimSpace(input.Notes)
	if !validText(input.Name, 100, true) || !validText(input.PrimaryMuscle, 80, true) ||
		!validText(input.SecondaryMuscles, 160, false) || !validText(input.Equipment, 80, true) ||
		!validText(input.Notes, 2000, false) {
		return 0, inputError("Enter a name, primary muscle and equipment, and keep optional fields within their limits.")
	}
	switch input.Tracking {
	case "weight_reps", "reps", "duration", "distance_duration":
	default:
		return 0, inputError("Choose an exercise tracking type.")
	}
	var id int64
	err := s.transaction(ctx, func(q *dbquery.Queries) error {
		_, err := q.FindExerciseByName(ctx, input.Name)
		if err == nil {
			return inputError("An exercise already uses that name.")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		id, err = q.CreateExercise(ctx, dbquery.CreateExerciseParams{
			Name: input.Name, PrimaryMuscle: input.PrimaryMuscle, SecondaryMuscles: input.SecondaryMuscles,
			Equipment: input.Equipment, Tracking: input.Tracking, Notes: input.Notes,
		})
		return err
	})
	return id, err
}

func readRoutine(ctx context.Context, q *dbquery.Queries, row dbquery.WorkoutRoutine) (Routine, error) {
	exercises, err := q.ListRoutineExercises(ctx, row.ID)
	routine := Routine{ID: row.ID, Name: row.Name, Exercises: make([]Plan, 0, len(exercises))}
	for _, exercise := range exercises {
		routine.Exercises = append(routine.Exercises, Plan{
			ExerciseID: exercise.ExerciseID, PlannedSets: exercise.PlannedSets,
			PlannedReps: exercise.PlannedReps, RestSeconds: exercise.RestSeconds,
		})
	}
	return routine, err
}

func (s *Service) SaveRoutine(ctx context.Context, routine Routine) (int64, error) {
	routine.Name = strings.TrimSpace(routine.Name)
	if routine.ID < 0 || !validText(routine.Name, 100, true) || len(routine.Exercises) < 1 || len(routine.Exercises) > 50 {
		return 0, inputError("Enter a routine name and add 1–50 exercises.")
	}
	id := routine.ID
	err := s.transaction(ctx, func(q *dbquery.Queries) error {
		for _, plan := range routine.Exercises {
			exercise, err := q.GetExercise(ctx, plan.ExerciseID)
			if errors.Is(err, sql.ErrNoRows) {
				return inputError("Choose an exercise from the library.")
			}
			if err != nil {
				return err
			}
			if plan.PlannedSets < 1 || plan.PlannedSets > 50 || plan.RestSeconds < 0 || plan.RestSeconds > 3600 {
				return inputError("Use 1–50 planned sets and 0–3,600 rest seconds per exercise.")
			}
			if tracksReps(exercise.Tracking) && (plan.PlannedReps < 1 || plan.PlannedReps > 999) {
				return inputError("Use 1–999 planned reps for rep-based exercises.")
			}
		}
		if id == 0 {
			var err error
			id, err = q.CreateRoutine(ctx, routine.Name)
			if err != nil {
				return err
			}
		} else {
			count, err := q.RenameRoutine(ctx, dbquery.RenameRoutineParams{ID: id, Name: routine.Name})
			if err := requireRow(count, err, "That routine is no longer available."); err != nil {
				return err
			}
			if err := q.ClearRoutineExercises(ctx, id); err != nil {
				return err
			}
		}
		return writePlans(ctx, q, id, routine.Exercises)
	})
	return id, err
}

func writePlans(ctx context.Context, q *dbquery.Queries, id int64, plans []Plan) error {
	for index, plan := range plans {
		exercise, err := q.GetExercise(ctx, plan.ExerciseID)
		if err != nil {
			return err
		}
		if !tracksReps(exercise.Tracking) {
			plan.PlannedReps = 0
		}
		if err := q.AddRoutineExercise(ctx, dbquery.AddRoutineExerciseParams{
			RoutineID: id, Position: int64(index), ExerciseID: plan.ExerciseID,
			PlannedSets: plan.PlannedSets, PlannedReps: plan.PlannedReps, RestSeconds: plan.RestSeconds,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) DuplicateRoutine(ctx context.Context, sourceID int64) (int64, error) {
	var id int64
	err := s.transaction(ctx, func(q *dbquery.Queries) error {
		row, err := q.GetRoutine(ctx, sourceID)
		if errors.Is(err, sql.ErrNoRows) {
			return inputError("That routine is no longer available.")
		}
		if err != nil {
			return err
		}
		routine, err := readRoutine(ctx, q, row)
		if err != nil {
			return err
		}
		name := []rune(row.Name)
		id, err = q.CreateRoutine(ctx, string(name[:min(len(name), 93)])+" (copy)")
		if err != nil {
			return err
		}
		return writePlans(ctx, q, id, routine.Exercises)
	})
	return id, err
}

func (s *Service) DeleteRoutine(ctx context.Context, id int64) error {
	count, err := s.q.DeleteRoutine(ctx, id)
	return requireRow(count, err, "That routine is no longer available.")
}

func tracksReps(tracking string) bool {
	return tracking == "weight_reps" || tracking == "reps"
}

func validText(value string, limit int, required bool) bool {
	return (!required || value != "") && !strings.ContainsRune(value, 0) && utf8.RuneCountInString(value) <= limit
}

func validNumber(value, maxValue float64) bool {
	return value >= 0 && value <= maxValue && !math.IsNaN(value)
}

func requireRow(count int64, err error, missing inputError) error {
	if err == nil && count == 0 {
		return missing
	}
	return err
}

// UserError translates storage errors at the desktop boundary.
func UserError(err error) error {
	var validation inputError
	if errors.As(err, &validation) || err == nil {
		return err
	}
	return errors.New("FitTrack couldn't access your workout data. Try again.")
}
