package desktop

import (
	"context"

	"github.com/vanam-san/fittrack/internal/workouts"
)

type WorkoutService struct{ workouts *workouts.Service }

func NewWorkoutService(service *workouts.Service) *WorkoutService {
	return &WorkoutService{workouts: service}
}

func (s *WorkoutService) Overview(ctx context.Context, unit string) (workouts.Overview, error) {
	result, err := s.workouts.Overview(ctx, unit)
	return result, workouts.UserError(err)
}

func (s *WorkoutService) Session(ctx context.Context, id int64, unit string) (workouts.Session, error) {
	result, err := s.workouts.Session(ctx, id, unit)
	return result, workouts.UserError(err)
}

func (s *WorkoutService) CreateExercise(ctx context.Context, input workouts.ExerciseInput) (int64, error) {
	id, err := s.workouts.CreateExercise(ctx, input)
	return id, workouts.UserError(err)
}

func (s *WorkoutService) SaveRoutine(ctx context.Context, routine workouts.Routine) (int64, error) {
	id, err := s.workouts.SaveRoutine(ctx, routine)
	return id, workouts.UserError(err)
}

func (s *WorkoutService) DuplicateRoutine(ctx context.Context, id int64) (int64, error) {
	copyID, err := s.workouts.DuplicateRoutine(ctx, id)
	return copyID, workouts.UserError(err)
}

func (s *WorkoutService) DeleteRoutine(ctx context.Context, id int64) error {
	return workouts.UserError(s.workouts.DeleteRoutine(ctx, id))
}

func (s *WorkoutService) Start(ctx context.Context, routineID int64) (int64, error) {
	id, err := s.workouts.Start(ctx, routineID)
	return id, workouts.UserError(err)
}

func (s *WorkoutService) ChangeStatus(ctx context.Context, id int64, target string) error {
	return workouts.UserError(s.workouts.ChangeStatus(ctx, id, target))
}

func (s *WorkoutService) AddSet(ctx context.Context, sessionID, exerciseID int64) error {
	return workouts.UserError(s.workouts.AddSet(ctx, sessionID, exerciseID))
}

func (s *WorkoutService) SaveSet(ctx context.Context, sessionID, exerciseID, setID int64, input workouts.SetInput) error {
	return workouts.UserError(s.workouts.SaveSet(ctx, sessionID, exerciseID, setID, input))
}

func (s *WorkoutService) SkipRest(ctx context.Context, id int64) error {
	return workouts.UserError(s.workouts.SkipRest(ctx, id))
}
