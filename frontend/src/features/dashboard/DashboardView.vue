<script setup lang="ts">
import { Activity, Dumbbell, Flame, Scale } from 'lucide-vue-next'
import { onMounted, ref } from 'vue'
import { Button } from '@/components/ui/button'
import NutritionTotals from '@/features/nutrition/NutritionTotals.vue'
import { formatDate, localDate } from '@/lib/format'
import { loadNutritionDay, type Day } from '@/services/nutrition'
import type { Preferences } from '@/services/settings'
import { loadWorkouts, type Overview } from '@/services/workouts'

const props = defineProps<{ preferences: Preferences }>()
const emit = defineEmits<{ navigate: [page: 'nutrition' | 'workouts'] }>()
const nutrition = ref<Day>()
const workouts = ref<Overview>()
const nutritionError = ref(false)
const workoutError = ref(false)

const now = new Date()
const today = new Intl.DateTimeFormat(undefined, {
  weekday: 'long',
  month: 'long',
  day: 'numeric',
}).format(now)

onMounted(async () => {
  const [journal, training] = await Promise.allSettled([
    Promise.resolve().then(() => loadNutritionDay(localDate(now))),
    Promise.resolve().then(() => loadWorkouts(props.preferences.weightUnit)),
  ])
  if (journal.status === 'fulfilled') nutrition.value = journal.value
  else nutritionError.value = true
  if (training.status === 'fulfilled') workouts.value = training.value
  else workoutError.value = true
})

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
        <NutritionTotals
          v-if="nutrition?.groups?.some((group) => group.entries?.length)"
          :total="nutrition.total"
        />
        <p v-else class="text-sm text-muted-foreground">
          {{
            nutritionError
              ? 'Could not load today’s nutrition.'
              : nutrition
                ? 'No food logged today.'
                : 'Loading nutrition…'
          }}
        </p>
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
        <Button
          type="button"
          variant="outline"
          @click="emit('navigate', 'nutrition')"
          >Open journal</Button
        >
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
        <div v-if="workouts?.active" class="text-sm">
          <p class="font-medium">{{ workouts.active.name }}</p>
          <p class="text-muted-foreground">
            {{
              workouts.active.status === 'paused'
                ? 'Paused workout'
                : 'Workout in progress'
            }}
          </p>
        </div>
        <div v-else-if="workouts?.history?.[0]" class="text-sm">
          <p class="font-medium">{{ workouts.history[0].name }}</p>
          <p class="text-muted-foreground">
            Completed {{ formatDate(workouts.history[0].date) }}
          </p>
        </div>
        <p v-else class="text-sm text-muted-foreground">
          {{
            workoutError
              ? 'Could not load workouts.'
              : workouts
                ? 'No workouts recorded yet.'
                : 'Loading workouts…'
          }}
        </p>
        <Button
          type="button"
          variant="outline"
          @click="emit('navigate', 'workouts')"
        >
          {{ workouts?.active ? 'Open current workout' : 'Open workouts' }}
        </Button>
      </section>
    </div>

    <section class="fit-panel" aria-labelledby="activity-title">
      <div class="flex items-center gap-3">
        <span class="fit-icon"><Activity :size="18" aria-hidden="true" /></span>
        <h2 id="activity-title" class="font-semibold">Recent activity</h2>
      </div>
      <ul v-if="workouts?.history?.length" class="mt-4 divide-y text-sm">
        <li
          v-for="session in workouts.history.slice(0, 3)"
          :key="session.id"
          class="flex flex-wrap justify-between gap-2 py-2"
        >
          <span>{{ session.name }}</span>
          <span class="text-muted-foreground">{{
            formatDate(session.date)
          }}</span>
        </li>
      </ul>
      <p v-else-if="workouts" class="mt-4 text-sm text-muted-foreground">
        Completed workouts will appear here.
      </p>
      <p v-else class="mt-4 text-sm text-muted-foreground">
        {{
          workoutError
            ? 'Could not load recent workouts.'
            : 'Loading recent workouts…'
        }}
      </p>
    </section>
  </div>
</template>
