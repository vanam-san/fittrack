<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import DateRangeFields from '@/components/DateRangeFields.vue'
import MetricChart from '@/components/MetricChart.vue'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import NutritionTotals from '@/features/nutrition/NutritionTotals.vue'
import WeightTrend from '@/features/weight/WeightTrend.vue'
import {
  errorMessage,
  formatDate,
  formatNumber,
  recentDateRange,
} from '@/lib/format'
import { displayWeight } from '@/lib/weight'
import {
  loadAnalytics,
  type ExerciseDay,
  type Report,
} from '@/services/analytics'

const props = defineProps<{ weightUnit: string }>()
const filters = ref({ ...recentDateRange(30), grouping: 'weekly' })
const report = ref<Report>()
const loading = ref(true)
const error = ref('')
const selectedExerciseId = ref<number>()
const selectedMetricKey = ref('')

const exercise = computed(() =>
  report.value?.exercises?.find(({ id }) => id === selectedExerciseId.value),
)
const metrics = computed(() => ({
  weightKg: {
    name: 'Heaviest set',
    unit: props.weightUnit,
    display: (value: number) => displayWeight(value, props.weightUnit),
  },
  reps: { name: 'Most reps in a set', unit: 'reps', display: Number },
  volumeKg: {
    name: 'Daily training volume',
    unit: `${props.weightUnit} × reps`,
    display: (value: number) => displayWeight(value, props.weightUnit),
  },
  durationSeconds: { name: 'Longest set', unit: 's', display: Number },
  distanceM: { name: 'Farthest set', unit: 'm', display: Number },
}))
type MetricKey = keyof typeof metrics.value & keyof ExerciseDay

const trackingMetrics: Record<string, MetricKey[]> = {
  weight_reps: ['weightKg', 'reps', 'volumeKg'],
  reps: ['reps'],
  duration: ['durationSeconds'],
  distance_duration: ['distanceM', 'durationSeconds'],
}
const metricKeys = computed(
  () => trackingMetrics[exercise.value?.tracking ?? ''] ?? [],
)
const metricKey = computed(
  () =>
    metricKeys.value.find((key) => key === selectedMetricKey.value) ??
    metricKeys.value[0],
)
const metric = computed(() =>
  metricKey.value ? metrics.value[metricKey.value] : undefined,
)
const exerciseSeries = computed(() => {
  const key = metricKey.value
  const selected = metric.value
  return key && selected
    ? [
        {
          name: selected.name,
          values:
            exercise.value?.days?.map((day) => selected.display(day[key])) ??
            [],
        },
      ]
    : []
})

const periodStarts = computed(
  () => report.value?.periods?.map(({ from }) => from) ?? [],
)
const periodEnds = computed(
  () => report.value?.periods?.map(({ to }) => to) ?? [],
)
const calories = computed(() => [
  {
    name: 'Calories',
    values:
      report.value?.periods?.map(({ nutrition }) => nutrition.calories) ?? [],
  },
])
const macros = computed(() => [
  {
    name: 'Protein',
    values:
      report.value?.periods?.map(({ nutrition }) => nutrition.protein) ?? [],
  },
  {
    name: 'Carbs',
    values:
      report.value?.periods?.map(({ nutrition }) => nutrition.carbohydrates) ??
      [],
  },
  {
    name: 'Fat',
    values: report.value?.periods?.map(({ nutrition }) => nutrition.fat) ?? [],
  },
])
const workoutCharts = computed(() => {
  const periods = report.value?.periods ?? []
  return [
    {
      title: 'Workout frequency',
      unit: 'sessions',
      series: [
        {
          name: 'Completed workouts',
          values: periods.map(({ workouts }) => workouts),
        },
      ],
    },
    {
      title: 'Workout duration',
      unit: 'min',
      series: [
        {
          name: 'Active time',
          values: periods.map(({ elapsedMs }) => elapsedMs / 60_000),
        },
      ],
    },
    {
      title: 'Training volume',
      unit: `${props.weightUnit} × reps`,
      series: [
        {
          name: 'Training volume',
          values: periods.map(({ volumeKg }) =>
            displayWeight(volumeKg, props.weightUnit),
          ),
        },
      ],
    },
  ]
})

async function refresh() {
  const { from, to, grouping } = filters.value
  loading.value = true
  error.value = ''
  report.value = undefined
  try {
    const next = await loadAnalytics(from, to, grouping)
    if (!next.exercises?.some(({ id }) => id === selectedExerciseId.value)) {
      selectedExerciseId.value = next.exercises?.[0]?.id
      selectedMetricKey.value = ''
    }
    report.value = next
  } catch (cause) {
    error.value = errorMessage(cause)
  } finally {
    loading.value = false
  }
}

onMounted(refresh)
</script>

<template>
  <div class="fit-page">
    <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">Analytics</h1>
    <form
      class="fit-panel flex flex-wrap items-end gap-3"
      @submit.prevent="refresh"
    >
      <DateRangeFields v-model:from="filters.from" v-model:to="filters.to" />
      <div class="space-y-1.5">
        <Label for="report-grouping">Group by</Label>
        <select
          id="report-grouping"
          v-model="filters.grouping"
          class="fit-select"
        >
          <option value="daily">Day</option>
          <option value="weekly">Week</option>
          <option value="monthly">Month</option>
        </select>
      </div>
      <Button type="submit" :disabled="loading">Apply filters</Button>
    </form>
    <p v-if="loading" role="status" class="text-sm text-muted-foreground">
      Calculating analytics…
    </p>
    <p v-else-if="error" role="alert" class="text-sm text-destructive">
      {{ error }}
    </p>

    <template v-else-if="report">
      <section
        class="fit-panel space-y-4"
        aria-labelledby="analytics-nutrition"
      >
        <h2 id="analytics-nutrition" class="font-semibold">Nutrition</h2>
        <NutritionTotals
          v-if="report.total.loggedDays"
          :total="report.total.nutrition"
        />
        <p class="text-sm text-muted-foreground">
          {{
            report.total.loggedDays
              ? `Food logged on ${report.total.loggedDays} ${report.total.loggedDays === 1 ? 'day' : 'days'} in this range.`
              : 'No food logged in this date range.'
          }}
        </p>
        <div
          v-if="report.total.loggedDays"
          class="grid gap-4 @min-[44rem]/main:grid-cols-2"
        >
          <MetricChart
            title="Calories"
            unit="kcal"
            type="bar"
            :from="report.total.from"
            :to="report.total.to"
            :dates="periodStarts"
            :period-ends="periodEnds"
            :series="calories"
            table
          />
          <MetricChart
            title="Macronutrients"
            unit="g"
            type="bar"
            :from="report.total.from"
            :to="report.total.to"
            :dates="periodStarts"
            :period-ends="periodEnds"
            :series="macros"
            table
          />
        </div>
      </section>

      <section class="fit-panel space-y-4" aria-labelledby="analytics-weight">
        <h2 id="analytics-weight" class="font-semibold">Weight</h2>
        <WeightTrend
          :measurements="report.measurements ?? []"
          :unit="weightUnit"
          :from="report.total.from"
          :to="report.total.to"
          table
        />
      </section>

      <section class="fit-panel space-y-4" aria-labelledby="analytics-workouts">
        <h2 id="analytics-workouts" class="font-semibold">Workouts</h2>
        <dl class="grid grid-cols-2 gap-3 text-sm sm:grid-cols-3">
          <div>
            <dt class="text-muted-foreground">Completed workouts</dt>
            <dd class="font-medium">
              {{ formatNumber(report.total.workouts) }}
            </dd>
          </div>
          <div>
            <dt class="text-muted-foreground">Total active time</dt>
            <dd class="font-medium">
              {{ formatNumber(report.total.elapsedMs / 60_000) }} min
            </dd>
          </div>
          <div>
            <dt class="text-muted-foreground">Total training volume</dt>
            <dd class="font-medium">
              {{
                formatNumber(displayWeight(report.total.volumeKg, weightUnit))
              }}
              {{ weightUnit }} × reps
            </dd>
          </div>
        </dl>
        <p v-if="!report.total.workouts" class="text-sm text-muted-foreground">
          No completed workouts in this date range.
        </p>
        <template v-else>
          <p class="text-sm text-muted-foreground">
            Training volume is the sum of load × reps across logged weighted
            sets.
          </p>
          <div class="grid gap-4 @min-[44rem]/main:grid-cols-2">
            <MetricChart
              v-for="chart in workoutCharts"
              :key="chart.title"
              v-bind="chart"
              type="bar"
              :from="report.total.from"
              :to="report.total.to"
              :dates="periodStarts"
              :period-ends="periodEnds"
              table
            />
          </div>
        </template>
      </section>

      <section
        class="fit-panel space-y-4"
        aria-labelledby="analytics-exercises"
      >
        <h2 id="analytics-exercises" class="font-semibold">
          Exercise progression
        </h2>
        <div
          v-if="report.exercises?.length"
          class="flex flex-wrap items-end gap-3"
        >
          <div class="min-w-0 space-y-1.5">
            <Label for="analytics-exercise">Exercise</Label>
            <select
              id="analytics-exercise"
              v-model="selectedExerciseId"
              class="fit-select max-w-full"
              @change="selectedMetricKey = ''"
            >
              <option
                v-for="item in report.exercises"
                :key="item.id"
                :value="item.id"
              >
                {{ item.name }}
              </option>
            </select>
          </div>
          <div v-if="metricKeys.length > 1" class="min-w-0 space-y-1.5">
            <Label for="analytics-metric">Metric</Label>
            <select
              id="analytics-metric"
              :value="metricKey"
              class="fit-select"
              @change="
                selectedMetricKey = ($event.target as HTMLSelectElement).value
              "
            >
              <option v-for="key in metricKeys" :key="key" :value="key">
                {{ metrics[key].name }}
              </option>
            </select>
          </div>
        </div>
        <template v-if="exercise && metric">
          <MetricChart
            :title="metric.name"
            :unit="metric.unit"
            type="line"
            :from="report.total.from"
            :to="report.total.to"
            :dates="exercise.days?.map(({ date }) => date) ?? []"
            :series="exerciseSeries"
            table
          />
          <h3 class="font-medium">Personal records in this range</h3>
          <dl class="grid gap-3 text-sm sm:grid-cols-2 xl:grid-cols-3">
            <div
              v-for="record in exercise.records"
              :key="record.metric"
              class="rounded-lg border border-border p-3"
            >
              <dt class="text-muted-foreground">
                {{ metrics[record.metric as MetricKey].name }}
              </dt>
              <dd class="font-medium">
                {{
                  formatNumber(
                    metrics[record.metric as MetricKey].display(record.value),
                  )
                }}
                {{ metrics[record.metric as MetricKey].unit }}
              </dd>
              <dd class="text-muted-foreground">
                {{ formatDate(record.date) }}
              </dd>
            </div>
          </dl>
        </template>
        <p v-else class="text-sm text-muted-foreground">
          No logged sets from completed workouts in this date range.
        </p>
      </section>
    </template>
  </div>
</template>
