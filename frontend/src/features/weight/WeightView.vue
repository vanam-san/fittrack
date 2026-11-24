<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import ConfirmAction from '@/components/ConfirmAction.vue'
import DateRangeFields from '@/components/DateRangeFields.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  errorMessage,
  formatDate,
  localDate,
  recentDateRange,
} from '@/lib/format'
import { displayWeight, formatWeight } from '@/lib/weight'
import {
  clearWeightGoal,
  deleteMeasurement,
  loadWeightOverview,
  loadWeightTrend,
  saveMeasurement,
  saveWeightGoal,
  type Measurement,
  type Overview,
} from '@/services/weight'
import WeightTrend from './WeightTrend.vue'

const props = defineProps<{ weightUnit: string }>()
const filters = ref(recentDateRange(30))
const range = ref({ ...filters.value })
const trend = ref<Measurement[]>()
const overview = ref<Overview>()
const date = ref(localDate())
const amount = ref('')
const target = ref('')
const editing = ref<Measurement>()
const removal = ref<Measurement>()
const loading = ref(true)
const saving = ref(false)
const overviewError = ref('')
const trendError = ref('')
const saveError = ref('')
const status = ref('')
const busy = computed(() => loading.value || saving.value)
const gap = computed(() => {
  const latest = overview.value?.latest
  const goal = overview.value?.goalKg
  return latest && goal != null ? latest.weightKg - goal : null
})

function inputWeight(kg: number) {
  return Number(displayWeight(kg, props.weightUnit).toFixed(4)).toString()
}

async function reloadOverview(syncTarget = false) {
  overviewError.value = ''
  try {
    overview.value = await loadWeightOverview()
    if (syncTarget) {
      const goal = overview.value.goalKg
      target.value = goal == null ? '' : inputWeight(goal)
    }
  } catch (error) {
    overviewError.value = errorMessage(error)
  }
}

async function reloadTrend() {
  trendError.value = ''
  trend.value = undefined
  try {
    trend.value =
      (await loadWeightTrend(range.value.from, range.value.to)) ?? []
  } catch (error) {
    trendError.value = errorMessage(error)
  }
}

async function applyRange() {
  if (busy.value) return
  range.value = { ...filters.value }
  loading.value = true
  await reloadTrend()
  loading.value = false
}

async function save(
  run: () => PromiseLike<void>,
  message: string,
  refreshTrend = false,
) {
  if (busy.value) return
  saving.value = true
  saveError.value = ''
  status.value = ''
  try {
    await run()
    status.value = message
    await Promise.all([
      reloadOverview(!refreshTrend),
      refreshTrend ? reloadTrend() : undefined,
    ])
  } catch (error) {
    saveError.value = errorMessage(error)
  } finally {
    saving.value = false
  }
}

function edit(measurement?: Measurement) {
  editing.value = measurement
  date.value = measurement?.date ?? localDate()
  amount.value = measurement ? inputWeight(measurement.weightKg) : ''
  saveError.value = ''
  if (measurement) document.getElementById('measurement-date')?.focus()
}

function recordValue(value: string, currentKg?: number | null) {
  // An unchanged display value should keep its original kilogram precision.
  if (currentKg != null && Number(value) === Number(inputWeight(currentKg))) {
    return { weight: currentKg, unit: 'kg' }
  }
  return { weight: Number(value), unit: props.weightUnit }
}

function submitMeasurement() {
  const input = {
    id: editing.value?.id ?? 0,
    date: date.value,
    ...recordValue(amount.value, editing.value?.weightKg),
  }
  const inRange = input.date >= range.value.from && input.date <= range.value.to
  void save(
    async () => {
      await saveMeasurement(input)
      edit()
    },
    inRange
      ? 'Measurement saved.'
      : 'Measurement saved outside the displayed date range.',
    true,
  )
}

function confirmRemoval() {
  const measurement = removal.value
  if (!measurement) return
  removal.value = undefined
  void save(
    async () => {
      await deleteMeasurement(measurement.id)
      if (editing.value?.id === measurement.id) edit()
    },
    'Measurement deleted.',
    true,
  )
}

onMounted(async () => {
  await Promise.all([reloadOverview(true), reloadTrend()])
  loading.value = false
})
</script>

<template>
  <div class="fit-page">
    <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">Weight</h1>
    <p v-if="status" role="status" class="text-sm text-muted-foreground">
      {{ status }}
    </p>
    <p v-if="saveError" role="alert" class="text-sm text-destructive">
      {{ saveError }}
    </p>
    <div class="grid gap-4 @min-[44rem]/main:grid-cols-2">
      <section
        class="fit-panel space-y-4"
        aria-labelledby="measurement-heading"
      >
        <h2 id="measurement-heading" class="font-semibold">
          {{ editing ? 'Edit measurement' : 'Add measurement' }}
        </h2>
        <form class="space-y-4" @submit.prevent="submitMeasurement">
          <div class="grid grid-cols-2 gap-3">
            <div class="min-w-0 space-y-1.5">
              <Label for="measurement-date">Date</Label>
              <Input
                id="measurement-date"
                v-model="date"
                type="date"
                required
                :disabled="busy"
              />
            </div>
            <div class="min-w-0 space-y-1.5">
              <Label for="measurement-weight">Weight ({{ weightUnit }})</Label>
              <Input
                id="measurement-weight"
                v-model="amount"
                type="number"
                step="any"
                required
                :disabled="busy"
              />
            </div>
          </div>
          <div class="flex flex-wrap gap-2">
            <Button type="submit" :disabled="busy">{{
              editing ? 'Save changes' : 'Save measurement'
            }}</Button>
            <Button
              v-if="editing"
              type="button"
              variant="outline"
              :disabled="busy"
              @click="edit()"
              >Cancel edit</Button
            >
          </div>
        </form>
      </section>

      <section class="fit-panel space-y-4" aria-labelledby="target-heading">
        <div class="space-y-1">
          <h2 id="target-heading" class="font-semibold">Target weight</h2>
          <p v-if="overviewError" role="alert" class="text-sm text-destructive">
            {{ overviewError }}
          </p>
          <p v-else-if="overview?.latest" class="text-sm text-muted-foreground">
            Latest measurement:
            {{ formatWeight(overview.latest.weightKg, weightUnit) }} on
            {{ formatDate(overview.latest.date) }}.
          </p>
          <p v-else-if="overview" class="text-sm text-muted-foreground">
            No measurements recorded yet.
          </p>
          <p v-if="gap !== null" class="text-sm text-muted-foreground">
            {{
              Math.abs(displayWeight(gap, weightUnit)) < 0.005
                ? 'At target weight'
                : `${formatWeight(Math.abs(gap), weightUnit)} ${gap > 0 ? 'above' : 'below'} target`
            }}
          </p>
        </div>
        <form
          class="space-y-4"
          @submit.prevent="
            save(
              () => saveWeightGoal(recordValue(target, overview?.goalKg)),
              'Target weight saved.',
            )
          "
        >
          <div class="space-y-1.5">
            <Label for="target-weight">Target ({{ weightUnit }})</Label>
            <Input
              id="target-weight"
              v-model="target"
              type="number"
              step="any"
              required
              :disabled="busy"
            />
          </div>
          <div class="flex flex-wrap gap-2">
            <Button type="submit" :disabled="busy">Save target</Button>
            <Button
              v-if="overview?.goalKg != null"
              type="button"
              variant="outline"
              :disabled="busy"
              @click="save(clearWeightGoal, 'Target weight cleared.')"
              >Clear target</Button
            >
          </div>
        </form>
      </section>
    </div>

    <section class="fit-panel space-y-4" aria-labelledby="weight-trend-heading">
      <h2 id="weight-trend-heading" class="font-semibold">Trend and history</h2>
      <form class="flex flex-wrap items-end gap-3" @submit.prevent="applyRange">
        <DateRangeFields v-model:from="filters.from" v-model:to="filters.to" />
        <Button type="submit" variant="outline" :disabled="busy"
          >Apply range</Button
        >
      </form>
      <p v-if="trendError" role="alert" class="text-sm text-destructive">
        {{ trendError }}
      </p>
      <p v-else-if="!trend" role="status" class="text-sm text-muted-foreground">
        Loading measurements…
      </p>
      <template v-else>
        <WeightTrend
          :measurements="trend"
          :unit="weightUnit"
          :from="range.from"
          :to="range.to"
        />
        <ul
          v-if="trend.length"
          class="divide-y"
          aria-label="Measurement history"
        >
          <li
            v-for="measurement in [...trend].reverse()"
            :key="measurement.id"
            class="flex flex-wrap items-center justify-between gap-2 py-2 text-sm"
          >
            <div class="flex flex-wrap gap-x-3">
              <span class="font-medium">{{
                formatWeight(measurement.weightKg, weightUnit)
              }}</span>
              <span class="text-muted-foreground">{{
                formatDate(measurement.date)
              }}</span>
            </div>
            <div class="flex gap-1">
              <Button
                type="button"
                variant="outline"
                size="sm"
                :disabled="busy"
                :aria-label="`Edit measurement from ${formatDate(measurement.date)}`"
                @click="edit(measurement)"
                >Edit</Button
              >
              <Button
                type="button"
                variant="ghost"
                size="sm"
                :disabled="busy"
                :aria-label="`Delete measurement from ${formatDate(measurement.date)}`"
                @click="removal = measurement"
                >Delete</Button
              >
            </div>
          </li>
        </ul>
      </template>
    </section>
    <ConfirmAction
      :action="
        removal
          ? {
              title: 'Delete measurement?',
              description: `Remove the measurement from ${formatDate(removal.date)} from your history.`,
              label: 'Delete measurement',
              destructive: true,
            }
          : null
      "
      @close="removal = undefined"
      @confirm="confirmRemoval"
    />
  </div>
</template>
