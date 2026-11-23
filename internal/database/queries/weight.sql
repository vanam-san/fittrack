-- name: LatestWeightMeasurement :one
SELECT * FROM weight_measurements ORDER BY measured_on DESC, id DESC LIMIT 1;

-- name: ListWeightMeasurements :many
SELECT * FROM weight_measurements
WHERE measured_on BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
ORDER BY measured_on, id;

-- name: CreateWeightMeasurement :exec
INSERT INTO weight_measurements (measured_on, weight_kg) VALUES (?, ?);

-- name: UpdateWeightMeasurement :execrows
UPDATE weight_measurements SET measured_on = ?, weight_kg = ? WHERE id = ?;

-- name: DeleteWeightMeasurement :execrows
DELETE FROM weight_measurements WHERE id = ?;

-- name: GetWeightGoal :one
SELECT weight_kg FROM weight_goal WHERE id = 1;

-- name: SaveWeightGoal :exec
INSERT INTO weight_goal (id, weight_kg) VALUES (1, ?)
ON CONFLICT (id) DO UPDATE SET weight_kg = excluded.weight_kg;

-- name: ClearWeightGoal :exec
DELETE FROM weight_goal WHERE id = 1;
