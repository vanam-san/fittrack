-- name: ListExercises :many
SELECT * FROM exercises ORDER BY name;

-- name: GetExercise :one
SELECT * FROM exercises WHERE id = ?;

-- name: FindExerciseByName :one
SELECT id FROM exercises WHERE name = ?;

-- name: CreateExercise :one
INSERT INTO exercises (name, primary_muscle, secondary_muscles, equipment, tracking, notes)
VALUES (?, ?, ?, ?, ?, ?) RETURNING id;

-- name: ListRoutines :many
SELECT * FROM workout_routines ORDER BY name COLLATE NOCASE, id;

-- name: GetRoutine :one
SELECT * FROM workout_routines WHERE id = ?;

-- name: ListRoutineExercises :many
SELECT * FROM routine_exercises WHERE routine_id = ? ORDER BY position;

-- name: CreateRoutine :one
INSERT INTO workout_routines (name) VALUES (?) RETURNING id;

-- name: RenameRoutine :execrows
UPDATE workout_routines SET name = ? WHERE id = ?;

-- name: DeleteRoutine :execrows
DELETE FROM workout_routines WHERE id = ?;

-- name: ClearRoutineExercises :exec
DELETE FROM routine_exercises WHERE routine_id = ?;

-- name: AddRoutineExercise :exec
INSERT INTO routine_exercises (routine_id, position, exercise_id, planned_sets, planned_reps, rest_seconds)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetOpenSession :one
SELECT * FROM workout_sessions WHERE status <> 'completed';

-- name: GetSession :one
SELECT * FROM workout_sessions WHERE id = ?;

-- name: ListCompletedSessions :many
SELECT * FROM workout_sessions WHERE status = 'completed' ORDER BY completed_at DESC, id DESC;

-- name: CreateSession :one
INSERT INTO workout_sessions (routine_name, session_date, started_at, status, resumed_at)
VALUES (?, ?, ?, 'active', ?) RETURNING id;

-- name: UpdateSessionState :exec
UPDATE workout_sessions SET status = ?, completed_at = ?, resumed_at = ?, elapsed_ms = ?, rest_until = ?, paused_rest_ms = ?
WHERE id = ?;

-- name: UpdateRestTimer :exec
UPDATE workout_sessions SET rest_until = ? WHERE id = ?;

-- name: SnapshotSessionExercises :exec
INSERT INTO session_exercises (session_id, position, exercise_id, exercise_name, tracking, planned_sets, planned_reps, rest_seconds)
SELECT sqlc.arg(session_id), r.position, e.id, e.name, e.tracking, r.planned_sets, r.planned_reps, r.rest_seconds
FROM routine_exercises r JOIN exercises e ON e.id = r.exercise_id
WHERE r.routine_id = sqlc.arg(routine_id);

-- name: ListSessionExercises :many
SELECT * FROM session_exercises WHERE session_id = ? ORDER BY position;

-- name: GetSessionExercise :one
SELECT * FROM session_exercises WHERE id = ? AND session_id = ?;

-- name: ListWorkoutSets :many
SELECT * FROM workout_sets WHERE session_exercise_id = ? ORDER BY position;

-- name: GetWorkoutSet :one
SELECT * FROM workout_sets WHERE id = ? AND session_exercise_id = ?;

-- name: AddWorkoutSet :exec
INSERT INTO workout_sets (session_exercise_id, position, reps) VALUES (?, ?, ?);

-- name: UpdateWorkoutSet :exec
UPDATE workout_sets SET weight_kg = ?, reps = ?, duration_seconds = ?, distance_m = ?, done = ?, notes = ?
WHERE id = ?;

-- name: PreviousWorkoutSets :many
SELECT w.* FROM workout_sets w
JOIN session_exercises e ON e.id = w.session_exercise_id
JOIN workout_sessions s ON s.id = e.session_id
WHERE w.done = 1 AND e.exercise_id = sqlc.arg(exercise_id) AND s.id = (
    SELECT prior.id FROM workout_sessions prior
    JOIN session_exercises pe ON pe.session_id = prior.id
    JOIN workout_sets pw ON pw.session_exercise_id = pe.id
    WHERE prior.status = 'completed' AND pe.exercise_id = sqlc.arg(exercise_id) AND pw.done = 1
    ORDER BY prior.completed_at DESC, prior.id DESC LIMIT 1
)
ORDER BY e.position, w.position;
