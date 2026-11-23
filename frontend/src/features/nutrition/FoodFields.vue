<script setup lang="ts">
import { nutrientFields } from './food'
import type { Food } from '@/services/nutrition'

const food = defineModel<Food>({ required: true })
</script>

<template>
  <div class="space-y-3">
    <div class="grid gap-3 sm:grid-cols-2">
      <label class="fit-label"
        >Food or drink name
        <input v-model="food.name" class="fit-field" required maxlength="100" />
      </label>
      <label class="fit-label"
        >One serving
        <input
          v-model="food.serving"
          class="fit-field"
          required
          maxlength="100"
          placeholder="e.g. 100 g or 1 cup"
        />
      </label>
    </div>
    <fieldset class="grid grid-cols-2 gap-3 sm:grid-cols-4">
      <legend class="mb-2 text-sm font-semibold">Nutrients per serving</legend>
      <label v-for="field in nutrientFields" :key="field.key" class="fit-label">
        {{ field.label }} ({{ field.unit }})
        <input
          v-model.number="food[field.key]"
          class="fit-field"
          type="number"
          min="0"
          :max="field.max"
          step="any"
          required
        />
      </label>
    </fieldset>
  </div>
</template>
