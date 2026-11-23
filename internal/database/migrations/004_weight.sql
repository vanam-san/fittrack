CREATE TABLE weight_measurements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    measured_on TEXT NOT NULL,
    weight_kg REAL NOT NULL CHECK (weight_kg > 0 AND weight_kg <= 1500)
);
CREATE INDEX weight_measurement_date ON weight_measurements (measured_on, id);

CREATE TABLE weight_goal (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    weight_kg REAL NOT NULL CHECK (weight_kg > 0 AND weight_kg <= 1500)
);

CREATE INDEX completed_workout_date ON workout_sessions (session_date, id) WHERE status = 'completed';
