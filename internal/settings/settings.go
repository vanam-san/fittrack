package settings

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/vanam-san/fittrack/internal/database/dbquery"
)

type Preferences struct {
	DisplayName      string `json:"displayName"`
	WeightUnit       string `json:"weightUnit"`
	Theme            string `json:"theme"`
	CalorieGoal      int64  `json:"calorieGoal"`
	ProteinGoal      int64  `json:"proteinGoal"`
	CarbohydrateGoal int64  `json:"carbohydrateGoal"`
	FatGoal          int64  `json:"fatGoal"`
}

type ValidationError struct{ message string }

func (e ValidationError) Error() string { return e.message }

type Service struct{ queries *dbquery.Queries }

func NewService(db *sql.DB) *Service { return &Service{queries: dbquery.New(db)} }

func (s *Service) Load(ctx context.Context) (Preferences, bool, error) {
	row, err := s.queries.GetSettings(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return Preferences{WeightUnit: "kg", Theme: "system"}, false, nil
	}
	if err != nil {
		return Preferences{}, false, err
	}
	return Preferences{
		DisplayName: row.DisplayName, WeightUnit: row.WeightUnit, Theme: row.Theme,
		CalorieGoal: row.CalorieGoal, ProteinGoal: row.ProteinGoal,
		CarbohydrateGoal: row.CarbohydrateGoal, FatGoal: row.FatGoal,
	}, true, nil
}

func (s *Service) Save(ctx context.Context, preferences Preferences) (Preferences, error) {
	preferences.DisplayName = strings.TrimSpace(preferences.DisplayName)
	if utf8.RuneCountInString(preferences.DisplayName) > 60 || strings.ContainsRune(preferences.DisplayName, 0) {
		return Preferences{}, ValidationError{"Enter a name with up to 60 characters."}
	}
	if preferences.WeightUnit != "kg" && preferences.WeightUnit != "lb" {
		return Preferences{}, ValidationError{"Choose kilograms or pounds."}
	}
	switch preferences.Theme {
	case "system", "light", "dark":
	default:
		return Preferences{}, ValidationError{"Choose system, light, or dark appearance."}
	}
	goals := []struct {
		value int64
		max   int64
	}{
		{preferences.CalorieGoal, 20000},
		{preferences.ProteinGoal, 5000},
		{preferences.CarbohydrateGoal, 5000},
		{preferences.FatGoal, 5000},
	}
	for _, goal := range goals {
		if goal.value < 0 || goal.value > goal.max {
			return Preferences{}, ValidationError{"Use 0–20,000 kcal or 0–5,000 g for each macro goal."}
		}
	}
	err := s.queries.SaveSettings(ctx, dbquery.SaveSettingsParams{
		DisplayName: preferences.DisplayName, WeightUnit: preferences.WeightUnit, Theme: preferences.Theme,
		CalorieGoal: preferences.CalorieGoal, ProteinGoal: preferences.ProteinGoal,
		CarbohydrateGoal: preferences.CarbohydrateGoal, FatGoal: preferences.FatGoal,
	})
	return preferences, err
}
