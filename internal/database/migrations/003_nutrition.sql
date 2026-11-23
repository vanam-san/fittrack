CREATE TABLE foods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    serving TEXT NOT NULL,
    calories REAL NOT NULL CHECK (calories BETWEEN 0 AND 20000),
    protein REAL NOT NULL CHECK (protein BETWEEN 0 AND 5000),
    carbohydrates REAL NOT NULL CHECK (carbohydrates BETWEEN 0 AND 5000),
    fat REAL NOT NULL CHECK (fat BETWEEN 0 AND 5000)
);

CREATE TABLE food_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    food_id INTEGER REFERENCES foods(id) ON DELETE SET NULL,
    log_date TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('meal', 'snack', 'drink')),
    name TEXT NOT NULL,
    serving TEXT NOT NULL,
    quantity REAL NOT NULL CHECK (quantity > 0 AND quantity <= 1000),
    calories REAL NOT NULL CHECK (calories BETWEEN 0 AND 20000),
    protein REAL NOT NULL CHECK (protein BETWEEN 0 AND 5000),
    carbohydrates REAL NOT NULL CHECK (carbohydrates BETWEEN 0 AND 5000),
    fat REAL NOT NULL CHECK (fat BETWEEN 0 AND 5000)
);
CREATE INDEX food_log_date ON food_logs (log_date DESC, id);
CREATE INDEX food_log_preset ON food_logs (food_id);
