-- name: GetSettings :one
SELECT * FROM app_settings WHERE id = 1;

-- name: SaveSettings :exec
INSERT INTO app_settings (
    id, display_name, weight_unit, theme, calorie_goal,
    protein_goal, carbohydrate_goal, fat_goal
) VALUES (1, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE SET
    display_name = excluded.display_name,
    weight_unit = excluded.weight_unit,
    theme = excluded.theme,
    calorie_goal = excluded.calorie_goal,
    protein_goal = excluded.protein_goal,
    carbohydrate_goal = excluded.carbohydrate_goal,
    fat_goal = excluded.fat_goal;
