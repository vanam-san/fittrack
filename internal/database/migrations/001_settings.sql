CREATE TABLE app_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    display_name TEXT NOT NULL CHECK (length(display_name) <= 60),
    weight_unit TEXT NOT NULL CHECK (weight_unit IN ('kg', 'lb')),
    theme TEXT NOT NULL CHECK (theme IN ('system', 'light', 'dark')),
    calorie_goal INTEGER NOT NULL CHECK (calorie_goal BETWEEN 0 AND 20000),
    protein_goal INTEGER NOT NULL CHECK (protein_goal BETWEEN 0 AND 5000),
    carbohydrate_goal INTEGER NOT NULL CHECK (carbohydrate_goal BETWEEN 0 AND 5000),
    fat_goal INTEGER NOT NULL CHECK (fat_goal BETWEEN 0 AND 5000)
);
