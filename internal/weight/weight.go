package weight

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/vanam-san/fittrack/internal/calendar"
	"github.com/vanam-san/fittrack/internal/database/dbquery"
	"github.com/vanam-san/fittrack/internal/units"
)

type Measurement struct {
	ID       int64   `json:"id"`
	Date     string  `json:"date"`
	WeightKg float64 `json:"weightKg"`
}

type Value struct {
	Weight float64 `json:"weight"`
	Unit   string  `json:"unit"`
}

type Input struct {
	ID   int64  `json:"id"`
	Date string `json:"date"`
	Value
}

type Overview struct {
	Latest *Measurement `json:"latest"`
	GoalKg *float64     `json:"goalKg"`
}

type inputError string

func (e inputError) Error() string { return string(e) }

type Service struct{ q *dbquery.Queries }

func NewService(db *sql.DB) *Service { return &Service{q: dbquery.New(db)} }

func (s *Service) Overview(ctx context.Context) (Overview, error) {
	var result Overview
	row, err := s.q.LatestWeightMeasurement(ctx)
	switch {
	case err == nil:
		result.Latest = &Measurement{ID: row.ID, Date: row.MeasuredOn, WeightKg: row.WeightKg}
	case !errors.Is(err, sql.ErrNoRows):
		return result, err
	}
	goal, err := s.q.GetWeightGoal(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return result, nil
	}
	if err == nil {
		result.GoalKg = &goal
	}
	return result, err
}

func (s *Service) Trend(ctx context.Context, from, to string) ([]Measurement, error) {
	if _, _, err := calendar.Range(from, to); err != nil {
		return nil, err
	}
	return Measurements(ctx, s.q, from, to)
}

func Measurements(ctx context.Context, q *dbquery.Queries, from, to string) ([]Measurement, error) {
	rows, err := q.ListWeightMeasurements(ctx, dbquery.ListWeightMeasurementsParams{DateFrom: from, DateTo: to})
	result := make([]Measurement, 0, len(rows))
	for _, row := range rows {
		result = append(result, Measurement{ID: row.ID, Date: row.MeasuredOn, WeightKg: row.WeightKg})
	}
	return result, err
}

func (s *Service) Save(ctx context.Context, input Input) error {
	if _, err := calendar.Parse(input.Date); err != nil {
		return err
	}
	kg, err := input.kilograms()
	if err != nil {
		return err
	}
	if input.ID == 0 {
		return s.q.CreateWeightMeasurement(ctx, dbquery.CreateWeightMeasurementParams{MeasuredOn: input.Date, WeightKg: kg})
	}
	count, err := s.q.UpdateWeightMeasurement(ctx, dbquery.UpdateWeightMeasurementParams{
		ID: input.ID, MeasuredOn: input.Date, WeightKg: kg,
	})
	return requireMeasurement(count, err)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return requireMeasurement(s.q.DeleteWeightMeasurement(ctx, id))
}

func (s *Service) SaveGoal(ctx context.Context, goal Value) error {
	kg, err := goal.kilograms()
	if err != nil {
		return err
	}
	return s.q.SaveWeightGoal(ctx, kg)
}

func (s *Service) ClearGoal(ctx context.Context) error {
	return s.q.ClearWeightGoal(ctx)
}

func (v Value) kilograms() (float64, error) {
	if v.Unit != "kg" && v.Unit != "lb" {
		return 0, inputError("Choose kilograms or pounds.")
	}
	kg := units.WeightInKG(v.Weight, v.Unit)
	if !(kg > 0 && kg <= 1500) {
		return 0, inputError("Enter a weight greater than zero and up to 1,500 kg or its equivalent.")
	}
	return kg, nil
}

func requireMeasurement(count int64, err error) error {
	if err == nil && count == 0 {
		return inputError("That measurement is no longer available.")
	}
	return err
}

func UserError(err error) error {
	var validation inputError
	var invalidDate calendar.DateError
	if err == nil || errors.As(err, &validation) || errors.As(err, &invalidDate) {
		return err
	}
	slog.Error("could not access local weight records")
	return errors.New("FitTrack couldn't access your weight records. Try again.")
}
