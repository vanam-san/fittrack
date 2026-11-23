-- name: ReportFoodLogs :many
SELECT log_date, category, quantity, calories, protein, carbohydrates, fat FROM food_logs
WHERE log_date >= sqlc.arg(date_from) AND log_date <= sqlc.arg(date_to)
ORDER BY log_date, CASE category WHEN 'meal' THEN 0 WHEN 'snack' THEN 1 ELSE 2 END, id;

-- name: ReportCompletedWorkouts :many
SELECT session_date, elapsed_ms FROM workout_sessions
WHERE status = 'completed' AND session_date >= sqlc.arg(date_from) AND session_date <= sqlc.arg(date_to)
ORDER BY session_date, id;

-- name: ReportCompletedSets :many
SELECT s.session_date, e.exercise_id, e.exercise_name, e.tracking,
       w.weight_kg, w.reps, w.duration_seconds, w.distance_m
FROM workout_sessions AS s
JOIN session_exercises AS e ON e.session_id = s.id
JOIN workout_sets AS w ON w.session_exercise_id = e.id
WHERE s.status = 'completed' AND w.done = 1
  AND s.session_date >= sqlc.arg(date_from) AND s.session_date <= sqlc.arg(date_to)
ORDER BY s.session_date, s.id, e.position, w.position;
