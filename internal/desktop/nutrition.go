package desktop

import (
	"context"

	"github.com/vanam-san/fittrack/internal/nutrition"
)

type NutritionService struct{ nutrition *nutrition.Service }

func NewNutritionService(service *nutrition.Service) *NutritionService {
	return &NutritionService{nutrition: service}
}

func (s *NutritionService) Day(ctx context.Context, date string) (nutrition.Day, error) {
	day, err := s.nutrition.Day(ctx, date)
	return day, nutrition.UserError(err)
}

func (s *NutritionService) Dates(ctx context.Context) ([]string, error) {
	dates, err := s.nutrition.Dates(ctx)
	return dates, nutrition.UserError(err)
}

func (s *NutritionService) Presets(ctx context.Context) ([]nutrition.Food, error) {
	foods, err := s.nutrition.Presets(ctx)
	return foods, nutrition.UserError(err)
}

func (s *NutritionService) SavePreset(ctx context.Context, food nutrition.Food) (int64, error) {
	id, err := s.nutrition.SavePreset(ctx, food)
	return id, nutrition.UserError(err)
}

func (s *NutritionService) DeletePreset(ctx context.Context, id int64) error {
	return nutrition.UserError(s.nutrition.DeletePreset(ctx, id))
}

func (s *NutritionService) SaveEntry(ctx context.Context, entry nutrition.Entry) (int64, error) {
	id, err := s.nutrition.SaveEntry(ctx, entry)
	return id, nutrition.UserError(err)
}

func (s *NutritionService) LogPreset(ctx context.Context, id int64, placement nutrition.LogPlacement) (int64, error) {
	logID, err := s.nutrition.LogPreset(ctx, id, placement)
	return logID, nutrition.UserError(err)
}

func (s *NutritionService) DeleteEntry(ctx context.Context, id int64) error {
	return nutrition.UserError(s.nutrition.DeleteEntry(ctx, id))
}
