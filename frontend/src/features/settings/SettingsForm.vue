<script setup lang="ts">
import { reactive, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { Preferences } from '@/services/settings'

const props = defineProps<{
  preferences: Preferences
  firstRun: boolean
  saving: boolean
  error: string
  saved: boolean
}>()
const emit = defineEmits<{ save: [preferences: Preferences] }>()

const form = reactive({ ...props.preferences })
watch(
  () => props.preferences,
  (preferences) => Object.assign(form, preferences),
)

const goals = [
  { field: 'calorieGoal', label: 'Calories', unit: 'kcal', max: 20000 },
  { field: 'proteinGoal', label: 'Protein', unit: 'g', max: 5000 },
  { field: 'carbohydrateGoal', label: 'Carbohydrates', unit: 'g', max: 5000 },
  { field: 'fatGoal', label: 'Fat', unit: 'g', max: 5000 },
] as const

function submit() {
  emit('save', {
    ...form,
    calorieGoal: Number(form.calorieGoal),
    proteinGoal: Number(form.proteinGoal),
    carbohydrateGoal: Number(form.carbohydrateGoal),
    fatGoal: Number(form.fatGoal),
  })
}
</script>

<template>
  <form class="space-y-6" @submit.prevent="submit">
    <fieldset class="space-y-4" :disabled="saving">
      <legend class="mb-3 text-sm font-semibold">
        Profile and preferences
      </legend>
      <div class="space-y-2">
        <Label for="display-name"
          >Display name
          <span class="text-muted-foreground">(optional)</span></Label
        >
        <Input
          id="display-name"
          v-model="form.displayName"
          autocomplete="off"
        />
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label for="weight-unit">Weight unit</Label>
          <select id="weight-unit" v-model="form.weightUnit" class="fit-select">
            <option value="kg">Kilograms (kg)</option>
            <option value="lb">Pounds (lb)</option>
          </select>
        </div>
        <div class="space-y-2">
          <Label for="theme">Appearance</Label>
          <select id="theme" v-model="form.theme" class="fit-select">
            <option value="system">Match system</option>
            <option value="light">Light</option>
            <option value="dark">Dark</option>
          </select>
        </div>
      </div>
    </fieldset>

    <fieldset :disabled="saving">
      <legend class="mb-1 text-sm font-semibold">Daily nutrition goals</legend>
      <p class="mb-4 text-sm text-muted-foreground">
        Leave a field blank or at zero to skip that goal.
      </p>
      <div class="grid grid-cols-2 gap-4">
        <div v-for="goal in goals" :key="goal.field" class="space-y-2">
          <Label :for="goal.field">{{ goal.label }} ({{ goal.unit }})</Label>
          <Input
            :id="goal.field"
            v-model="form[goal.field]"
            type="number"
            min="0"
            :max="goal.max"
            step="1"
          />
        </div>
      </div>
    </fieldset>

    <p v-if="error" role="alert" class="text-sm text-destructive">
      {{ error }}
    </p>
    <div class="flex flex-wrap items-center gap-3">
      <Button type="submit" :disabled="saving">
        {{
          saving
            ? 'Saving…'
            : firstRun
              ? 'Start using FitTrack'
              : 'Save changes'
        }}
      </Button>
      <span v-if="saved" role="status" class="text-sm text-muted-foreground"
        >Settings saved.</span
      >
    </div>
  </form>
</template>
