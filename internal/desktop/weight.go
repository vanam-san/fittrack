package desktop

import (
	"context"

	"github.com/vanam-san/fittrack/internal/weight"
)

type WeightService struct{ weight *weight.Service }

func NewWeightService(service *weight.Service) *WeightService {
	return &WeightService{weight: service}
}

func (s *WeightService) Overview(ctx context.Context) (weight.Overview, error) {
	result, err := s.weight.Overview(ctx)
	return result, weight.UserError(err)
}

func (s *WeightService) Trend(ctx context.Context, from, to string) ([]weight.Measurement, error) {
	result, err := s.weight.Trend(ctx, from, to)
	return result, weight.UserError(err)
}

func (s *WeightService) Save(ctx context.Context, measurement weight.Input) error {
	return weight.UserError(s.weight.Save(ctx, measurement))
}

func (s *WeightService) Delete(ctx context.Context, id int64) error {
	return weight.UserError(s.weight.Delete(ctx, id))
}

func (s *WeightService) SaveGoal(ctx context.Context, goal weight.Value) error {
	return weight.UserError(s.weight.SaveGoal(ctx, goal))
}

func (s *WeightService) ClearGoal(ctx context.Context) error {
	return weight.UserError(s.weight.ClearGoal(ctx))
}
