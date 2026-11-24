package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/vanam-san/fittrack/internal/calendar"
)

const portableVersion = 1

type records map[string][]map[string]any

type archive struct {
	Version int     `json:"version"`
	Tables  records `json:"tables"`
}

type table struct {
	name   string
	fields string
}

// Parent tables precede their dependents. Keep version 1 fields stable when
// adding database migrations. The format stores canonical units and IDs.
var portableTables = []table{
	{"app_settings", "id display_name weight_unit theme calorie_goal protein_goal carbohydrate_goal fat_goal"},
	{"exercises", "id name primary_muscle secondary_muscles equipment tracking notes built_in"},
	{"workout_routines", "id name"},
	{"routine_exercises", "routine_id position exercise_id planned_sets planned_reps rest_seconds"},
	{"workout_sessions", "id routine_name session_date started_at status completed_at resumed_at elapsed_ms rest_until paused_rest_ms"},
	{"session_exercises", "id session_id position exercise_id exercise_name tracking planned_sets planned_reps rest_seconds"},
	{"workout_sets", "id session_exercise_id position weight_kg reps duration_seconds distance_m done notes"},
	{"foods", "id name serving calories protein carbohydrates fat"},
	{"food_logs", "id food_id log_date category name serving quantity calories protein carbohydrates fat"},
	{"weight_measurements", "id measured_on weight_kg"},
	{"weight_goal", "id weight_kg"},
}

func fieldKind(field string) byte {
	switch field {
	case "display_name", "weight_unit", "theme", "name", "primary_muscle", "secondary_muscles", "equipment", "tracking", "notes", "routine_name", "status", "exercise_name", "serving", "category":
		return 's'
	case "session_date", "log_date", "measured_on":
		return 'd'
	case "weight_kg", "distance_m", "calories", "protein", "carbohydrates", "fat", "quantity":
		return 'f'
	default:
		return 'i'
	}
}

func parseField(field, text string) (any, error) {
	switch fieldKind(field) {
	case 'd':
		if _, err := calendar.Parse(text); err != nil {
			return nil, err
		}
		return text, nil
	case 's':
		if !utf8.ValidString(text) {
			return nil, fmt.Errorf("invalid UTF-8 text")
		}
		if field == "tracking" {
			switch text {
			case "weight_reps", "reps", "duration", "distance_duration":
			default:
				return nil, fmt.Errorf("unknown exercise tracking type")
			}
		}
		return text, nil
	case 'f':
		value, err := strconv.ParseFloat(text, 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return nil, fmt.Errorf("expected a finite nonnegative number")
		}
		return value, nil
	default:
		value, err := strconv.ParseInt(text, 10, 64)
		if err != nil || value < 0 {
			return nil, fmt.Errorf("expected a nonnegative integer")
		}
		if (field == "id" || strings.HasSuffix(field, "_id")) && (value == 0 || value == math.MaxInt64) {
			return nil, fmt.Errorf("record ID is outside the supported range")
		}
		return value, nil
	}
}

func normalizeRecord(columns []string, row map[string]any) ([]any, error) {
	if len(row) != len(columns) {
		return nil, fmt.Errorf("unexpected number of fields")
	}
	result := make([]any, len(columns))
	for i, field := range columns {
		value, exists := row[field]
		if !exists {
			return nil, fmt.Errorf("missing field %s", field)
		}
		if value == nil && field == "food_id" {
			continue
		}
		var text string
		switch v := value.(type) {
		case string:
			if kind := fieldKind(field); kind == 's' || kind == 'd' {
				text = v
				break
			}
			return nil, fmt.Errorf("%s must be a number", field)
		case json.Number, int64, float64:
			if kind := fieldKind(field); kind == 'i' || kind == 'f' {
				text = fmt.Sprint(v)
				break
			}
			return nil, fmt.Errorf("%s must be text", field)
		default:
			return nil, fmt.Errorf("invalid value for %s", field)
		}
		parsed, err := parseField(field, text)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", field, err)
		}
		result[i] = parsed
	}
	return result, nil
}

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func readRecords(ctx context.Context, db queryer) (records, error) {
	data := make(records, len(portableTables))
	for _, table := range portableTables {
		columns := strings.Fields(table.fields)
		// Every table begins with its primary key, except routine_exercises,
		// which begins with both columns of its composite key.
		order := "1"
		if table.name == "routine_exercises" {
			order = "1, 2"
		}
		rows, err := db.QueryContext(ctx, "SELECT "+strings.Join(columns, ", ")+" FROM "+table.name+" ORDER BY "+order)
		if err != nil {
			return nil, err
		}
		data[table.name], err = scanRecords(rows, columns)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", table.name, err)
		}
	}
	return data, nil
}

func scanRecords(rows *sql.Rows, columns []string) ([]map[string]any, error) {
	defer rows.Close()
	result := make([]map[string]any, 0)
	values := make([]any, len(columns))
	ptrs := make([]any, len(columns))
	for i := range values {
		ptrs[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(columns))
		for i, column := range columns {
			row[column] = values[i]
		}
		if _, err := normalizeRecord(columns, row); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func replaceRecords(ctx context.Context, tx *sql.Tx, data records) error {
	if len(data) != len(portableTables) {
		return fmt.Errorf("unexpected set of tables")
	}
	for i := len(portableTables) - 1; i >= 0; i-- {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+portableTables[i].name); err != nil {
			return err
		}
	}
	for _, table := range portableTables {
		if err := insertRecords(ctx, tx, table, data[table.name]); err != nil {
			return fmt.Errorf("%s: %w", table.name, err)
		}
	}
	// Older portable exports can contain built-in exercises without instructions.
	instructions, err := migrations.ReadFile("migrations/006_exercise_instructions.sql")
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, string(instructions)); err != nil {
		return fmt.Errorf("complete built-in exercise details: %w", err)
	}
	return nil
}

func insertRecords(ctx context.Context, tx *sql.Tx, table table, rows []map[string]any) error {
	if rows == nil {
		return fmt.Errorf("missing table or null row list")
	}
	columns := strings.Fields(table.fields)
	values := strings.TrimSuffix(strings.Repeat("?, ", len(columns)), ", ")
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO "+table.name+" ("+strings.Join(columns, ", ")+") VALUES ("+values+")")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for i, row := range rows {
		args, err := normalizeRecord(columns, row)
		if err == nil {
			_, err = stmt.ExecContext(ctx, args...)
		}
		if err != nil {
			return fmt.Errorf("row %d: %w", i+1, err)
		}
	}
	return nil
}
