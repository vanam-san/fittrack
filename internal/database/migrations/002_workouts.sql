CREATE TABLE exercises (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL COLLATE NOCASE UNIQUE,
    primary_muscle TEXT NOT NULL,
    secondary_muscles TEXT NOT NULL DEFAULT '',
    equipment TEXT NOT NULL,
    tracking TEXT NOT NULL CHECK (tracking IN ('weight_reps', 'reps', 'duration', 'distance_duration')),
    notes TEXT NOT NULL DEFAULT '',
    built_in INTEGER NOT NULL DEFAULT 0 CHECK (built_in IN (0, 1))
);

INSERT INTO exercises (name, primary_muscle, secondary_muscles, equipment, tracking, notes, built_in) VALUES
    ('Barbell Squat', 'Quadriceps', 'Glutes, core', 'Barbell', 'weight_reps', 'Brace before descending and keep your feet planted.', 1),
    ('Barbell Bench Press', 'Chest', 'Triceps, shoulders', 'Barbell', 'weight_reps', 'Set your shoulder blades and lower the bar under control.', 1),
    ('Deadlift', 'Hamstrings', 'Glutes, back', 'Barbell', 'weight_reps', 'Hinge at the hips and keep the bar close.', 1),
    ('Dumbbell Row', 'Back', 'Biceps, rear deltoids', 'Dumbbell', 'weight_reps', 'Pull your elbow toward your hip.', 1),
    ('Overhead Press', 'Shoulders', 'Triceps, core', 'Barbell', 'weight_reps', 'Brace your torso as you press overhead.', 1),
    ('Dumbbell Curl', 'Biceps', 'Forearms', 'Dumbbell', 'weight_reps', 'Keep your upper arms steady through each rep.', 1),
    ('Lat Pulldown', 'Back', 'Biceps', 'Cable machine', 'weight_reps', 'Pull the bar toward your upper chest.', 1),
    ('Leg Press', 'Quadriceps', 'Glutes', 'Machine', 'weight_reps', 'Keep your hips supported against the pad.', 1),
    ('Push-Up', 'Chest', 'Triceps, core', 'Bodyweight', 'reps', 'Keep your body in a straight line.', 1),
    ('Pull-Up', 'Back', 'Biceps', 'Pull-up bar', 'reps', 'Lift without swinging and lower under control.', 1),
    ('Plank', 'Core', 'Shoulders', 'Bodyweight', 'duration', 'Keep your ribs and hips aligned.', 1),
    ('Running', 'Legs', 'Core', 'Treadmill or outdoors', 'distance_duration', 'Record elapsed time and distance.', 1);

CREATE TABLE workout_routines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);
CREATE TABLE routine_exercises (
    routine_id INTEGER NOT NULL REFERENCES workout_routines(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id),
    planned_sets INTEGER NOT NULL CHECK (planned_sets BETWEEN 1 AND 50),
    planned_reps INTEGER NOT NULL CHECK (planned_reps BETWEEN 0 AND 999),
    rest_seconds INTEGER NOT NULL CHECK (rest_seconds BETWEEN 0 AND 3600),
    PRIMARY KEY (routine_id, position)
);

CREATE TABLE workout_sessions (
    id INTEGER PRIMARY KEY,
    routine_name TEXT NOT NULL,
    session_date TEXT NOT NULL,
    started_at INTEGER NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'paused', 'completed')),
    completed_at INTEGER NOT NULL DEFAULT 0,
    resumed_at INTEGER NOT NULL,
    elapsed_ms INTEGER NOT NULL DEFAULT 0,
    rest_until INTEGER NOT NULL DEFAULT 0,
    paused_rest_ms INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX one_unfinished_workout ON workout_sessions ((1)) WHERE status <> 'completed';
CREATE INDEX workout_session_history ON workout_sessions (completed_at DESC, id DESC) WHERE status = 'completed';

CREATE TABLE session_exercises (
    id INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id),
    exercise_name TEXT NOT NULL,
    tracking TEXT NOT NULL,
    planned_sets INTEGER NOT NULL,
    planned_reps INTEGER NOT NULL,
    rest_seconds INTEGER NOT NULL,
    UNIQUE (session_id, position)
);
CREATE INDEX session_exercise_history ON session_exercises (exercise_id, session_id);

CREATE TABLE workout_sets (
    id INTEGER PRIMARY KEY,
    session_exercise_id INTEGER NOT NULL REFERENCES session_exercises(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    weight_kg REAL NOT NULL DEFAULT 0,
    reps INTEGER NOT NULL DEFAULT 0,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    distance_m REAL NOT NULL DEFAULT 0,
    done INTEGER NOT NULL DEFAULT 0 CHECK (done IN (0, 1)),
    notes TEXT NOT NULL DEFAULT '',
    UNIQUE (session_exercise_id, position)
);
