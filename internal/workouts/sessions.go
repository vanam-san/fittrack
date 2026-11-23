package workouts

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/vanam-san/fittrack/internal/database/dbquery"
)

const kgPerPound = 0.45359237

func (s *Service) Start(ctx context.Context, routineID int64) (int64, error) {
	var id int64
	err := s.transaction(ctx, func(q *dbquery.Queries) error {
		if _, err := q.GetOpenSession(ctx); err == nil {
			return inputError("Finish the current workout before starting another.")
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		routine, err := q.GetRoutine(ctx, routineID)
		if errors.Is(err, sql.ErrNoRows) {
			return inputError("That routine is no longer available.")
		}
		if err != nil {
			return err
		}
		now := s.now()
		id, err = q.CreateSession(ctx, dbquery.CreateSessionParams{
			RoutineName: routine.Name, SessionDate: now.Format(time.DateOnly),
			StartedAt: now.UnixMilli(), ResumedAt: now.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if err := q.SnapshotSessionExercises(ctx, dbquery.SnapshotSessionExercisesParams{
			SessionID: id, RoutineID: routineID,
		}); err != nil {
			return err
		}
		exercises, err := q.ListSessionExercises(ctx, id)
		if err != nil {
			return err
		}
		for _, exercise := range exercises {
			for position := range exercise.PlannedSets {
				if err := q.AddWorkoutSet(ctx, dbquery.AddWorkoutSetParams{
					SessionExerciseID: exercise.ID, Position: position, Reps: exercise.PlannedReps,
				}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return id, err
}

func (s *Service) Session(ctx context.Context, id int64, unit string) (Session, error) {
	var session Session
	err := s.transaction(ctx, func(q *dbquery.Queries) error {
		row, err := q.GetSession(ctx, id)
		if errors.Is(err, sql.ErrNoRows) {
			return inputError("That workout is no longer available.")
		}
		if err != nil {
			return err
		}
		session, err = readSession(ctx, q, row, unit, s.now().UnixMilli())
		return err
	})
	return session, err
}

func summarize(row dbquery.WorkoutSession, now int64) SessionSummary {
	elapsed := row.ElapsedMs
	if row.Status == "active" {
		elapsed += max(0, now-row.ResumedAt)
	}
	return SessionSummary{
		ID: row.ID, Name: row.RoutineName, Date: row.SessionDate, Status: row.Status,
		StartedAt: row.StartedAt, CompletedAt: row.CompletedAt, ElapsedMS: elapsed,
	}
}

func readSession(ctx context.Context, q *dbquery.Queries, row dbquery.WorkoutSession, unit string, now int64) (Session, error) {
	session := Session{
		SessionSummary: summarize(row, now), RestUntil: row.RestUntil, PausedRestMS: row.PausedRestMs,
		Exercises: []SessionExercise{},
	}
	exercises, err := q.ListSessionExercises(ctx, row.ID)
	if err != nil {
		return Session{}, err
	}
	for _, exercise := range exercises {
		sets, err := q.ListWorkoutSets(ctx, exercise.ID)
		if err != nil {
			return Session{}, err
		}
		var previous []dbquery.WorkoutSet
		if row.Status != "completed" {
			previous, err = q.PreviousWorkoutSets(ctx, exercise.ExerciseID)
			if err != nil {
				return Session{}, err
			}
		}
		session.Exercises = append(session.Exercises, SessionExercise{
			ID: exercise.ID, Name: exercise.ExerciseName, Tracking: exercise.Tracking,
			Plan: Plan{
				ExerciseID: exercise.ExerciseID, PlannedSets: exercise.PlannedSets,
				PlannedReps: exercise.PlannedReps, RestSeconds: exercise.RestSeconds,
			},
			Sets: displaySets(sets, unit), Previous: displaySets(previous, unit),
		})
	}
	return session, nil
}

func displaySets(rows []dbquery.WorkoutSet, unit string) []Set {
	if unit != "lb" {
		unit = "kg"
	}
	sets := make([]Set, 0, len(rows))
	for _, row := range rows {
		weight := row.WeightKg
		if unit == "lb" {
			weight /= kgPerPound
		}
		sets = append(sets, Set{
			ID: row.ID,
			SetInput: SetInput{
				Weight: weight, WeightUnit: unit, Reps: row.Reps, DurationSeconds: row.DurationSeconds,
				DistanceM: row.DistanceM, Done: row.Done != 0, Notes: row.Notes,
			},
		})
	}
	return sets
}

func (s *Service) ChangeStatus(ctx context.Context, id int64, target string) error {
	return s.transaction(ctx, func(q *dbquery.Queries) error {
		row, err := q.GetSession(ctx, id)
		if errors.Is(err, sql.ErrNoRows) {
			return inputError("That workout is no longer available.")
		}
		if err != nil {
			return err
		}
		now := s.now().UnixMilli()
		update := dbquery.UpdateSessionStateParams{ID: id, Status: target, ElapsedMs: summarize(row, now).ElapsedMS}
		switch {
		case target == row.Status:
			return nil
		case target == "paused" && row.Status == "active":
			update.PausedRestMs = max(0, row.RestUntil-now)
		case target == "active" && row.Status == "paused":
			update.ResumedAt = now
			if row.PausedRestMs > 0 {
				update.RestUntil = now + row.PausedRestMs
			}
		case target == "completed" && row.Status != "completed":
			update.CompletedAt = now
		default:
			return inputError("That action is unavailable for this workout.")
		}
		return q.UpdateSessionState(ctx, update)
	})
}

func activeExercise(ctx context.Context, q *dbquery.Queries, sessionID, exerciseID int64) (dbquery.SessionExercise, error) {
	row, err := q.GetSession(ctx, sessionID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && row.Status != "active") {
		return dbquery.SessionExercise{}, inputError("Resume the workout before logging sets.")
	}
	if err != nil {
		return dbquery.SessionExercise{}, err
	}
	exercise, err := q.GetSessionExercise(ctx, dbquery.GetSessionExerciseParams{ID: exerciseID, SessionID: sessionID})
	if errors.Is(err, sql.ErrNoRows) {
		return dbquery.SessionExercise{}, inputError("That exercise is no longer available in this workout.")
	}
	return exercise, err
}

func (s *Service) AddSet(ctx context.Context, sessionID, exerciseID int64) error {
	return s.transaction(ctx, func(q *dbquery.Queries) error {
		exercise, err := activeExercise(ctx, q, sessionID, exerciseID)
		if err != nil {
			return err
		}
		sets, err := q.ListWorkoutSets(ctx, exercise.ID)
		if err != nil {
			return err
		}
		if len(sets) >= 100 {
			return inputError("A workout exercise can have up to 100 sets.")
		}
		return q.AddWorkoutSet(ctx, dbquery.AddWorkoutSetParams{
			SessionExerciseID: exercise.ID, Position: int64(len(sets)), Reps: exercise.PlannedReps,
		})
	})
}

func (s *Service) SaveSet(ctx context.Context, sessionID, exerciseID, setID int64, input SetInput) error {
	return s.transaction(ctx, func(q *dbquery.Queries) error {
		exercise, err := activeExercise(ctx, q, sessionID, exerciseID)
		if err != nil {
			return err
		}
		set, err := q.GetWorkoutSet(ctx, dbquery.GetWorkoutSetParams{ID: setID, SessionExerciseID: exercise.ID})
		if errors.Is(err, sql.ErrNoRows) {
			return inputError("That set is no longer available.")
		}
		if err != nil {
			return err
		}
		update, err := validateSet(set.ID, exercise.Tracking, input)
		if err != nil {
			return err
		}
		if err := q.UpdateWorkoutSet(ctx, update); err != nil {
			return err
		}
		if input.Done && set.Done == 0 && exercise.RestSeconds > 0 {
			return q.UpdateRestTimer(ctx, dbquery.UpdateRestTimerParams{
				ID: sessionID, RestUntil: s.now().UnixMilli() + exercise.RestSeconds*1000,
			})
		}
		return nil
	})
}

func validateSet(id int64, tracking string, input SetInput) (dbquery.UpdateWorkoutSetParams, error) {
	update := dbquery.UpdateWorkoutSetParams{ID: id, Notes: strings.TrimSpace(input.Notes)}
	if !validText(update.Notes, 1000, false) {
		return update, inputError("Use up to 1,000 characters for set notes.")
	}
	if tracking == "weight_reps" {
		update.WeightKg = input.Weight
		switch input.WeightUnit {
		case "lb":
			update.WeightKg *= kgPerPound
		case "kg":
		default:
			return update, inputError("Choose kilograms or pounds for set weight.")
		}
		if !validNumber(update.WeightKg, 5000) {
			return update, inputError("Enter a nonnegative set weight up to 5,000 kg or its equivalent.")
		}
	}
	if tracksReps(tracking) {
		if input.Reps < 0 || input.Reps > 999 || (input.Done && input.Reps == 0) {
			return update, inputError("Enter 1–999 reps for a completed set.")
		}
		update.Reps = input.Reps
	} else {
		if input.DurationSeconds < 0 || input.DurationSeconds > 86400 || (input.Done && input.DurationSeconds == 0) {
			return update, inputError("Enter 1–86,400 seconds for a completed set.")
		}
		update.DurationSeconds = input.DurationSeconds
	}
	if tracking == "distance_duration" {
		if !validNumber(input.DistanceM, 1000000) || (input.Done && input.DistanceM == 0) {
			return update, inputError("Enter a positive distance up to 1,000,000 metres for a completed set.")
		}
		update.DistanceM = input.DistanceM
	}
	if input.Done {
		update.Done = 1
	}
	return update, nil
}

func (s *Service) SkipRest(ctx context.Context, id int64) error {
	return s.transaction(ctx, func(q *dbquery.Queries) error {
		row, err := q.GetSession(ctx, id)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && row.Status != "active") {
			return inputError("Resume the workout before changing the rest timer.")
		}
		if err != nil {
			return err
		}
		return q.UpdateRestTimer(ctx, dbquery.UpdateRestTimerParams{ID: id})
	})
}
