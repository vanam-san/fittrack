package nutrition

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/vanam-san/fittrack/internal/database/dbquery"
)

type Nutrients struct {
	Calories      float64 `json:"calories"`
	Protein       float64 `json:"protein"`
	Carbohydrates float64 `json:"carbohydrates"`
	Fat           float64 `json:"fat"`
}

type Food struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Serving string `json:"serving"`
	Nutrients
}

type LogPlacement struct {
	Date     string  `json:"date"`
	Category string  `json:"category"`
	Quantity float64 `json:"quantity"`
}

type Entry struct {
	ID int64 `json:"id"`
	LogPlacement
	Food  Food      `json:"food"`
	Total Nutrients `json:"total"`
}

type Group struct {
	Category string    `json:"category"`
	Entries  []Entry   `json:"entries"`
	Total    Nutrients `json:"total"`
}

type Day struct {
	Date   string    `json:"date"`
	Groups []Group   `json:"groups"`
	Total  Nutrients `json:"total"`
}

type inputError string

func (e inputError) Error() string { return string(e) }

type Service struct{ q *dbquery.Queries }

func NewService(db *sql.DB) *Service { return &Service{q: dbquery.New(db)} }

func (s *Service) Presets(ctx context.Context) ([]Food, error) {
	rows, err := s.q.ListFoods(ctx)
	foods := make([]Food, 0, len(rows))
	for _, row := range rows {
		foods = append(foods, Food{
			ID: row.ID, Name: row.Name, Serving: row.Serving,
			Nutrients: Nutrients{Calories: row.Calories, Protein: row.Protein, Carbohydrates: row.Carbohydrates, Fat: row.Fat},
		})
	}
	return foods, err
}

func (s *Service) SavePreset(ctx context.Context, food Food) (int64, error) {
	if err := validateFood(&food); err != nil {
		return 0, err
	}
	if food.ID == 0 {
		return s.q.CreateFood(ctx, dbquery.CreateFoodParams{
			Name: food.Name, Serving: food.Serving, Calories: food.Calories, Protein: food.Protein,
			Carbohydrates: food.Carbohydrates, Fat: food.Fat,
		})
	}
	count, err := s.q.UpdateFood(ctx, dbquery.UpdateFoodParams{
		ID: food.ID, Name: food.Name, Serving: food.Serving, Calories: food.Calories, Protein: food.Protein,
		Carbohydrates: food.Carbohydrates, Fat: food.Fat,
	})
	return food.ID, requireRow(count, err, "That food preset is no longer available.")
}

func (s *Service) DeletePreset(ctx context.Context, id int64) error {
	count, err := s.q.DeleteFood(ctx, id)
	return requireRow(count, err, "That food preset is no longer available.")
}

func (s *Service) Dates(ctx context.Context) ([]string, error) {
	dates, err := s.q.ListFoodLogDates(ctx)
	if dates == nil {
		dates = []string{}
	}
	return dates, err
}

func (s *Service) Day(ctx context.Context, date string) (Day, error) {
	if err := validateDate(date); err != nil {
		return Day{}, err
	}
	rows, err := s.q.ListFoodLogs(ctx, date)
	if err != nil {
		return Day{}, err
	}
	day := Day{Date: date}
	for _, category := range []string{"meal", "snack", "drink"} {
		group := Group{Category: category, Entries: []Entry{}}
		for _, row := range rows {
			if row.Category != category {
				continue
			}
			nutrients := Nutrients{Calories: row.Calories, Protein: row.Protein, Carbohydrates: row.Carbohydrates, Fat: row.Fat}
			total := nutrients.scale(row.Quantity)
			group.Entries = append(group.Entries, Entry{
				ID:           row.ID,
				LogPlacement: LogPlacement{Date: row.LogDate, Category: row.Category, Quantity: row.Quantity},
				Food:         Food{Name: row.Name, Serving: row.Serving, Nutrients: nutrients},
				Total:        total,
			})
			group.Total.add(total)
		}
		day.Groups = append(day.Groups, group)
		day.Total.add(group.Total)
	}
	return day, nil
}

func (s *Service) SaveEntry(ctx context.Context, entry Entry) (int64, error) {
	if err := validatePlacement(entry.LogPlacement); err != nil {
		return 0, err
	}
	food := entry.Food
	if err := validateFood(&food); err != nil {
		return 0, err
	}
	if entry.ID == 0 {
		return s.q.CreateFoodLog(ctx, dbquery.CreateFoodLogParams{
			LogDate: entry.Date, Category: entry.Category, Quantity: entry.Quantity, Name: food.Name, Serving: food.Serving,
			Calories: food.Calories, Protein: food.Protein, Carbohydrates: food.Carbohydrates, Fat: food.Fat,
		})
	}
	count, err := s.q.UpdateFoodLog(ctx, dbquery.UpdateFoodLogParams{
		ID: entry.ID, LogDate: entry.Date, Category: entry.Category, Quantity: entry.Quantity, Name: food.Name, Serving: food.Serving,
		Calories: food.Calories, Protein: food.Protein, Carbohydrates: food.Carbohydrates, Fat: food.Fat,
	})
	return entry.ID, requireRow(count, err, "That food entry is no longer available.")
}

func (s *Service) LogPreset(ctx context.Context, id int64, placement LogPlacement) (int64, error) {
	if err := validatePlacement(placement); err != nil {
		return 0, err
	}
	logID, err := s.q.LogFoodPreset(ctx, dbquery.LogFoodPresetParams{
		FoodID: id, LogDate: placement.Date, Category: placement.Category, Quantity: placement.Quantity,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, inputError("That food preset is no longer available.")
	}
	return logID, err
}

func (s *Service) DeleteEntry(ctx context.Context, id int64) error {
	count, err := s.q.DeleteFoodLog(ctx, id)
	return requireRow(count, err, "That food entry is no longer available.")
}

func (n Nutrients) scale(quantity float64) Nutrients {
	return Nutrients{
		Calories: n.Calories * quantity, Protein: n.Protein * quantity,
		Carbohydrates: n.Carbohydrates * quantity, Fat: n.Fat * quantity,
	}
}

func (n *Nutrients) add(other Nutrients) {
	n.Calories += other.Calories
	n.Protein += other.Protein
	n.Carbohydrates += other.Carbohydrates
	n.Fat += other.Fat
}

func validateFood(food *Food) error {
	food.Name = strings.TrimSpace(food.Name)
	food.Serving = strings.TrimSpace(food.Serving)
	for _, value := range []string{food.Name, food.Serving} {
		if value == "" || strings.ContainsRune(value, 0) || utf8.RuneCountInString(value) > 100 {
			return inputError("Enter a food name and a serving description of up to 100 characters each.")
		}
	}
	if !inRange(food.Calories, 20000) || !inRange(food.Protein, 5000) || !inRange(food.Carbohydrates, 5000) || !inRange(food.Fat, 5000) {
		return inputError("Use 0–20,000 kcal and 0–5,000 g for each macro per serving.")
	}
	return nil
}

func inRange(value, max float64) bool {
	return value >= 0 && value <= max
}

func validateDate(date string) error {
	parsed, err := time.Parse(time.DateOnly, date)
	if err != nil || parsed.Year() < 1 {
		return inputError("Choose a valid calendar date.")
	}
	return nil
}

func validatePlacement(placement LogPlacement) error {
	if err := validateDate(placement.Date); err != nil {
		return err
	}
	switch placement.Category {
	case "meal", "snack", "drink":
	default:
		return inputError("Choose meal, snack, or drink.")
	}
	if !(placement.Quantity > 0 && placement.Quantity <= 1000) {
		return inputError("Use a serving quantity greater than zero and up to 1,000.")
	}
	return nil
}

func requireRow(count int64, err error, missing inputError) error {
	if err == nil && count == 0 {
		return missing
	}
	return err
}

func UserError(err error) error {
	var validation inputError
	if errors.As(err, &validation) || err == nil {
		return err
	}
	return errors.New("FitTrack couldn't access your nutrition data. Try again.")
}
