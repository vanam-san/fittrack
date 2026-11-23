-- name: ListFoods :many
SELECT * FROM foods ORDER BY name COLLATE NOCASE, id;

-- name: CreateFood :one
INSERT INTO foods (name, serving, calories, protein, carbohydrates, fat)
VALUES (?, ?, ?, ?, ?, ?) RETURNING id;

-- name: UpdateFood :execrows
UPDATE foods SET name = ?, serving = ?, calories = ?, protein = ?, carbohydrates = ?, fat = ? WHERE id = ?;

-- name: DeleteFood :execrows
DELETE FROM foods WHERE id = ?;

-- name: ListFoodLogDates :many
SELECT DISTINCT log_date FROM food_logs ORDER BY log_date DESC;

-- name: ListFoodLogs :many
SELECT * FROM food_logs WHERE log_date = ? ORDER BY id;

-- name: CreateFoodLog :one
INSERT INTO food_logs (log_date, category, name, serving, quantity, calories, protein, carbohydrates, fat)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id;

-- name: LogFoodPreset :one
INSERT INTO food_logs (food_id, log_date, category, name, serving, quantity, calories, protein, carbohydrates, fat)
SELECT f.id, sqlc.arg(log_date), sqlc.arg(category), f.name, f.serving, CAST(sqlc.arg(quantity) AS REAL), f.calories, f.protein, f.carbohydrates, f.fat
FROM foods f WHERE f.id = sqlc.arg(food_id) RETURNING food_logs.id;

-- name: UpdateFoodLog :execrows
UPDATE food_logs SET log_date = ?, category = ?, name = ?, serving = ?, quantity = ?, calories = ?, protein = ?, carbohydrates = ?, fat = ?
WHERE id = ?;

-- name: DeleteFoodLog :execrows
DELETE FROM food_logs WHERE id = ?;
