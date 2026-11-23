package analytics

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/vanam-san/fittrack/internal/calendar"
	"github.com/vanam-san/fittrack/internal/database/dbquery"
	"github.com/vanam-san/fittrack/internal/nutrition"
	"github.com/vanam-san/fittrack/internal/weight"
)

type Period struct {
	From       string              `json:"from"`
	To         string              `json:"to"`
	Nutrition  nutrition.Nutrients `json:"nutrition"`
	LoggedDays int                 `json:"loggedDays"`
	Workouts   int                 `json:"workouts"`
	ElapsedMS  int64               `json:"elapsedMs"`
	VolumeKg   float64             `json:"volumeKg"`
}

type ExerciseDay struct {
	Date            string  `json:"date"`
	WeightKg        float64 `json:"weightKg"`
	Reps            int64   `json:"reps"`
	DurationSeconds int64   `json:"durationSeconds"`
	DistanceM       float64 `json:"distanceM"`
	VolumeKg        float64 `json:"volumeKg"`
}

type Record struct {
	Metric string  `json:"metric"`
	Date   string  `json:"date"`
	Value  float64 `json:"value"`
}

type Exercise struct {
	ID       int64         `json:"id"`
	Name     string        `json:"name"`
	Tracking string        `json:"tracking"`
	Days     []ExerciseDay `json:"days"`
	Records  []Record      `json:"records"`
}

type Report struct {
	Total        Period               `json:"total"`
	Periods      []Period             `json:"periods"`
	Measurements []weight.Measurement `json:"measurements"`
	Exercises    []Exercise           `json:"exercises"`
}

type inputError string

func (e inputError) Error() string { return string(e) }

type Service struct{ db *sql.DB }

func NewService(db *sql.DB) *Service { return &Service{db: db} }

func (s *Service) Report(ctx context.Context, from, to, grouping string) (Report, error) {
	start, end, err := calendar.Range(from, to)
	if err != nil {
		return Report{}, err
	}
	periods, err := splitPeriods(start, end, grouping)
	if err != nil {
		return Report{}, err
	}

	// Read every source from the same database snapshot.
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Report{}, err
	}
	defer tx.Rollback()
	q := dbquery.New(tx)
	food, err := q.ReportFoodLogs(ctx, dbquery.ReportFoodLogsParams{DateFrom: from, DateTo: to})
	if err != nil {
		return Report{}, err
	}
	workouts, err := q.ReportCompletedWorkouts(ctx, dbquery.ReportCompletedWorkoutsParams{DateFrom: from, DateTo: to})
	if err != nil {
		return Report{}, err
	}
	sets, err := q.ReportCompletedSets(ctx, dbquery.ReportCompletedSetsParams{DateFrom: from, DateTo: to})
	if err != nil {
		return Report{}, err
	}
	measurements, err := weight.Measurements(ctx, q, from, to)
	if err != nil {
		return Report{}, err
	}
	if err := tx.Commit(); err != nil {
		return Report{}, err
	}

	result := Report{Total: Period{From: from, To: to}, Periods: periods, Measurements: measurements}
	result.addNutrition(food)
	for _, workout := range workouts {
		period := result.periodFor(workout.SessionDate)
		period.Workouts++
		period.ElapsedMS += workout.ElapsedMs
	}
	result.Exercises = result.addExercises(sets)
	for _, period := range periods {
		result.Total.Nutrition.Add(period.Nutrition)
		result.Total.LoggedDays += period.LoggedDays
		result.Total.Workouts += period.Workouts
		result.Total.ElapsedMS += period.ElapsedMS
		result.Total.VolumeKg += period.VolumeKg
	}
	return result, nil
}

func splitPeriods(start, end time.Time, grouping string) ([]Period, error) {
	// Keep calendar boundaries for complete periods and clip the first and last
	// periods to the requested range.
	var nextBoundary func(time.Time) time.Time
	switch grouping {
	case "daily":
		nextBoundary = func(day time.Time) time.Time { return day.AddDate(0, 0, 1) }
	case "weekly":
		nextBoundary = func(day time.Time) time.Time {
			// Monday begins the reporting week.
			return day.AddDate(0, 0, 7-(int(day.Weekday())+6)%7)
		}
	case "monthly":
		nextBoundary = func(day time.Time) time.Time {
			return time.Date(day.Year(), day.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		}
	default:
		return nil, inputError("Choose daily, weekly, or monthly summaries.")
	}
	var periods []Period
	for day := start; !day.After(end); {
		if len(periods) == 3660 {
			return nil, inputError("Choose a shorter date range or a larger summary period.")
		}
		next := nextBoundary(day)
		last := next.AddDate(0, 0, -1)
		if last.After(end) {
			last = end
		}
		periods = append(periods, Period{From: day.Format(time.DateOnly), To: last.Format(time.DateOnly)})
		day = next
	}
	return periods, nil
}

func (r Report) periodFor(date string) *Period {
	index, _ := slices.BinarySearchFunc(r.Periods, date, func(period Period, date string) int {
		switch {
		case period.To < date:
			return -1
		case period.From > date:
			return 1
		default:
			return 0
		}
	})
	return &r.Periods[index]
}

func (r Report) addNutrition(rows []dbquery.ReportFoodLogsRow) {
	// Match the journal's order: sum each meal category, then the day.
	for len(rows) > 0 {
		date := rows[0].LogDate
		var total nutrition.Nutrients
		for len(rows) > 0 && rows[0].LogDate == date {
			category := rows[0].Category
			var subtotal nutrition.Nutrients
			for len(rows) > 0 && rows[0].LogDate == date && rows[0].Category == category {
				row := rows[0]
				snapshot := nutrition.Nutrients{Calories: row.Calories, Protein: row.Protein, Carbohydrates: row.Carbohydrates, Fat: row.Fat}
				subtotal.Add(snapshot.ForServings(row.Quantity))
				rows = rows[1:]
			}
			total.Add(subtotal)
		}
		period := r.periodFor(date)
		period.LoggedDays++
		period.Nutrition.Add(total)
	}
}

func (r Report) addExercises(rows []dbquery.ReportCompletedSetsRow) []Exercise {
	// The query sorts sets by date. Each exercise's final day can collect
	// sets from multiple sessions on that date.
	byID := make(map[int64]*Exercise)
	for _, row := range rows {
		exercise := byID[row.ExerciseID]
		if exercise == nil {
			exercise = &Exercise{ID: row.ExerciseID, Name: row.ExerciseName, Tracking: row.Tracking, Records: []Record{}}
			byID[row.ExerciseID] = exercise
		}
		if len(exercise.Days) == 0 || exercise.Days[len(exercise.Days)-1].Date != row.SessionDate {
			exercise.Days = append(exercise.Days, ExerciseDay{Date: row.SessionDate})
		}
		day := &exercise.Days[len(exercise.Days)-1]
		switch row.Tracking {
		case "weight_reps":
			day.WeightKg = max(day.WeightKg, row.WeightKg)
			volume := row.WeightKg * float64(row.Reps)
			day.VolumeKg += volume
			r.periodFor(row.SessionDate).VolumeKg += volume
			fallthrough
		case "reps":
			day.Reps = max(day.Reps, row.Reps)
		case "distance_duration":
			day.DistanceM = max(day.DistanceM, row.DistanceM)
			fallthrough
		case "duration":
			day.DurationSeconds = max(day.DurationSeconds, row.DurationSeconds)
		}
	}
	exercises := make([]Exercise, 0, len(byID))
	for _, exercise := range byID {
		for _, day := range exercise.Days {
			switch exercise.Tracking {
			case "weight_reps":
				exercise.updateRecord("weightKg", day.Date, day.WeightKg)
				exercise.updateRecord("volumeKg", day.Date, day.VolumeKg)
				fallthrough
			case "reps":
				exercise.updateRecord("reps", day.Date, float64(day.Reps))
			case "distance_duration":
				exercise.updateRecord("distanceM", day.Date, day.DistanceM)
				fallthrough
			case "duration":
				exercise.updateRecord("durationSeconds", day.Date, float64(day.DurationSeconds))
			}
		}
		exercises = append(exercises, *exercise)
	}
	slices.SortFunc(exercises, func(a, b Exercise) int { return strings.Compare(a.Name, b.Name) })
	return exercises
}

func (e *Exercise) updateRecord(metric, date string, value float64) {
	for i, record := range e.Records {
		if record.Metric == metric {
			if value > record.Value {
				e.Records[i] = Record{Metric: metric, Date: date, Value: value}
			}
			return
		}
	}
	e.Records = append(e.Records, Record{Metric: metric, Date: date, Value: value})
}

func UserError(err error) error {
	var validation inputError
	var invalidDate calendar.DateError
	if err == nil || errors.As(err, &validation) || errors.As(err, &invalidDate) {
		return err
	}
	slog.Error("could not calculate local analytics")
	return errors.New("FitTrack couldn't load analytics. Try again.")
}
