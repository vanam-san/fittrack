import type { Food, Nutrients } from '@/services/nutrition'

export const nutrientFields = [
  { key: 'calories', label: 'Calories', unit: 'kcal', max: 20000 },
  { key: 'protein', label: 'Protein', unit: 'g', max: 5000 },
  { key: 'carbohydrates', label: 'Carbs', unit: 'g', max: 5000 },
  { key: 'fat', label: 'Fat', unit: 'g', max: 5000 },
] as const

export const categoryLabels: Record<string, string> = {
  meal: 'Meals',
  snack: 'Snacks',
  drink: 'Drinks',
}

export function zeroNutrients(): Nutrients {
  return { calories: 0, protein: 0, carbohydrates: 0, fat: 0 }
}

export function newFood(): Food {
  return { id: 0, name: '', serving: '', ...zeroNutrients() }
}
