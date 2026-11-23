<script setup lang="ts">
import { Activity, Dumbbell, Flame, Scale } from 'lucide-vue-next'
import type { Preferences } from '@/services/settings'

defineProps<{ preferences: Preferences }>()

const today = new Intl.DateTimeFormat(undefined, {
  weekday: 'long',
  month: 'long',
  day: 'numeric',
}).format(new Date())

const macros = [
  { label: 'Protein', key: 'proteinGoal' },
  { label: 'Carbohydrates', key: 'carbohydrateGoal' },
  { label: 'Fat', key: 'fatGoal' },
] as const
</script>

<template>
  <div class="space-y-7">
    <header class="space-y-2">
      <p class="text-sm text-muted-foreground">{{ today }}</p>
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">
        Dashboard
      </h1>
      <p v-if="preferences.displayName" class="text-muted-foreground">
        Welcome, {{ preferences.displayName }}.
      </p>
    </header>

    <div class="grid gap-4 xl:grid-cols-3">
      <section class="fit-panel space-y-4" aria-labelledby="nutrition-title">
        <div class="flex items-center gap-3">
          <span class="fit-icon"><Flame :size="18" aria-hidden="true" /></span>
          <h2 id="nutrition-title" class="font-semibold">Today’s nutrition</h2>
        </div>
        <p class="text-sm text-muted-foreground">No food logged today.</p>
        <p v-if="preferences.calorieGoal" class="text-sm">
          Daily goal
          <strong class="ml-1"
            >{{ preferences.calorieGoal.toLocaleString() }} kcal</strong
          >
        </p>
        <dl
          v-if="macros.some(({ key }) => preferences[key] > 0)"
          class="space-y-2 border-t border-border pt-3 text-sm"
        >
          <div
            v-for="macro in macros"
            :key="macro.key"
            class="flex justify-between gap-2"
          >
            <dt class="text-muted-foreground">{{ macro.label }}</dt>
            <dd v-if="preferences[macro.key]">
              {{ preferences[macro.key] }} g
            </dd>
            <dd v-else aria-label="No goal">—</dd>
          </div>
        </dl>
      </section>
      <section class="fit-panel space-y-4" aria-labelledby="weight-title">
        <div class="flex items-center gap-3">
          <span class="fit-icon"><Scale :size="18" aria-hidden="true" /></span>
          <h2 id="weight-title" class="font-semibold">Latest weight</h2>
        </div>
        <p class="text-sm text-muted-foreground">
          No measurements recorded yet.
        </p>
      </section>
      <section class="fit-panel space-y-4" aria-labelledby="workout-title">
        <div class="flex items-center gap-3">
          <span class="fit-icon"
            ><Dumbbell :size="18" aria-hidden="true"
          /></span>
          <h2 id="workout-title" class="font-semibold">Workouts</h2>
        </div>
        <p class="text-sm text-muted-foreground">No workouts recorded yet.</p>
      </section>
    </div>

    <section class="fit-panel" aria-labelledby="activity-title">
      <div class="flex items-center gap-3">
        <span class="fit-icon"><Activity :size="18" aria-hidden="true" /></span>
        <h2 id="activity-title" class="font-semibold">Recent activity</h2>
      </div>
      <p class="mt-4 text-sm text-muted-foreground">
        Your saved activity will appear here.
      </p>
    </section>
  </div>
</template>
