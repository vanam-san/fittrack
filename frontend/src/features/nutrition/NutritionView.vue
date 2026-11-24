<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ChevronLeft, ChevronRight, Plus } from 'lucide-vue-next'
import ConfirmAction from '@/components/ConfirmAction.vue'
import { Button } from '@/components/ui/button'
import { errorMessage, formatDate, formatNumber, localDate } from '@/lib/format'
import {
  deleteFoodEntry,
  deleteFoodPreset,
  loadFoodLogDates,
  loadFoodPresets,
  loadNutritionDay,
  logFoodPreset,
  saveFoodEntry,
  saveFoodPreset,
  type Day,
  type Entry,
  type Food,
} from '@/services/nutrition'
import EntryEditor from './EntryEditor.vue'
import FoodFields from './FoodFields.vue'
import NutritionTotals from './NutritionTotals.vue'
import { categoryLabels, newFood, zeroNutrients } from './food'

const section = ref<'journal' | 'presets'>('journal')
const date = ref(localDate())
const day = ref<Day>()
const dates = ref<string[]>([])
const presets = ref<Food[]>([])
const loaded = ref(false)
const busy = ref(false)
const error = ref('')
const notice = ref('')
const entry = ref<Entry>()
const draft = ref<Food>()
const confirmation = ref<{
  title: string
  description: string
  label: string
  destructive: true
  apply: () => Promise<void>
} | null>(null)

async function run(action: () => Promise<unknown>) {
  if (busy.value) return
  busy.value = true
  error.value = notice.value = ''
  try {
    await action()
  } catch (failure) {
    error.value = errorMessage(failure)
  } finally {
    busy.value = false
  }
}

async function refresh(target: string) {
  const [journal, history] = await Promise.all([
    loadNutritionDay(target),
    loadFoodLogDates(),
  ])
  date.value = target
  day.value = journal
  dates.value = history ?? []
}

function load() {
  void run(async () => {
    const [list] = await Promise.all([loadFoodPresets(), refresh(date.value)])
    presets.value = list ?? []
    loaded.value = true
  })
}

function selectDate(target: string) {
  if (!target || target === date.value) return
  void run(() => refresh(target))
}

function shiftDay(step: number) {
  const target = new Date(`${date.value}T12:00:00`)
  target.setDate(target.getDate() + step)
  const next = localDate(target)
  if (next.length === 10 && next >= '0001-01-01') selectDate(next)
}

function addEntry(category = 'meal', food = newFood()) {
  entry.value = {
    id: 0,
    date: date.value,
    category,
    quantity: 1,
    food,
    total: zeroNutrients(),
  }
  section.value = 'journal'
}

function submitEntry(target: string, save: () => Promise<number>) {
  void run(async () => {
    await save()
    entry.value = undefined
    await refresh(target)
    notice.value = 'Food entry saved.'
  })
}

function submitPreset() {
  const food = draft.value
  if (!food) return
  void run(async () => {
    await saveFoodPreset(food)
    draft.value = undefined
    presets.value = (await loadFoodPresets()) ?? []
    notice.value = 'Food preset saved.'
  })
}

function askDelete(item: Entry | Food, kind: 'entry' | 'preset') {
  const isEntry = kind === 'entry'
  const name = 'food' in item ? item.food.name : item.name
  confirmation.value = {
    title: isEntry ? 'Delete food entry?' : 'Delete food preset?',
    description: isEntry
      ? `“${name}” will be removed from this day’s journal and totals.`
      : `“${name}” will be removed from your presets. Existing food logs keep their saved nutrients.`,
    label: isEntry ? 'Delete entry' : 'Delete preset',
    destructive: true,
    apply: async () => {
      if (isEntry) {
        await deleteFoodEntry(item.id)
        if (entry.value?.id === item.id) entry.value = undefined
        await refresh(date.value)
      } else {
        await deleteFoodPreset(item.id)
        if (draft.value?.id === item.id) draft.value = undefined
        if (entry.value?.id === 0 && entry.value.food.id === item.id)
          entry.value = undefined
        presets.value = (await loadFoodPresets()) ?? []
      }
      notice.value = isEntry ? 'Food entry deleted.' : 'Food preset deleted.'
    },
  }
}

function confirm() {
  const action = confirmation.value
  confirmation.value = null
  if (action) void run(action.apply)
}

onMounted(load)
</script>

<template>
  <div class="fit-page">
    <header class="flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">
        Nutrition
      </h1>
      <Button
        v-if="loaded && section === 'journal' && !entry"
        type="button"
        :disabled="busy"
        @click="addEntry()"
      >
        <Plus aria-hidden="true" /> Add entry
      </Button>
      <Button
        v-if="loaded && section === 'presets' && !draft"
        type="button"
        :disabled="busy"
        @click="draft = newFood()"
      >
        <Plus aria-hidden="true" /> New preset
      </Button>
    </header>
    <p v-if="error" role="alert" class="fit-panel text-sm text-destructive">
      {{ error }}
    </p>
    <p v-if="notice" role="status" class="text-sm text-primary">{{ notice }}</p>
    <div v-if="!loaded" class="fit-panel text-sm text-muted-foreground">
      <p v-if="busy" role="status">Loading nutrition…</p>
      <Button v-else type="button" variant="outline" @click="load"
        >Retry</Button
      >
    </div>
    <template v-else>
      <nav aria-label="Nutrition sections" class="fit-sections">
        <Button
          type="button"
          :variant="section === 'journal' ? 'secondary' : 'ghost'"
          :aria-current="section === 'journal' ? 'page' : undefined"
          :disabled="busy"
          @click="section = 'journal'"
          >Daily journal</Button
        >
        <Button
          type="button"
          :variant="section === 'presets' ? 'secondary' : 'ghost'"
          :aria-current="section === 'presets' ? 'page' : undefined"
          :disabled="busy"
          @click="section = 'presets'"
          >Food presets</Button
        >
      </nav>
      <div v-if="section === 'journal' && day" class="space-y-4">
        <section
          class="fit-panel space-y-4"
          aria-label="Daily nutrition totals"
        >
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p class="text-sm text-muted-foreground">Daily totals</p>
              <h2 class="text-lg font-semibold">{{ formatDate(date) }}</h2>
            </div>
            <div
              class="grid max-w-full grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-2"
            >
              <Button
                type="button"
                size="icon"
                variant="outline"
                aria-label="Previous day"
                :disabled="busy || !!entry || date === '0001-01-01'"
                @click="shiftDay(-1)"
                ><ChevronLeft
              /></Button>
              <input
                :value="date"
                class="fit-field"
                type="date"
                min="0001-01-01"
                max="9999-12-31"
                aria-label="Journal date"
                :disabled="busy || !!entry"
                @change="
                  (event) => {
                    const field = event.target as HTMLInputElement
                    if (field.value && field.validity.valid)
                      selectDate(field.value)
                    field.value = date
                  }
                "
              />
              <Button
                type="button"
                size="icon"
                variant="outline"
                aria-label="Next day"
                :disabled="busy || !!entry || date === '9999-12-31'"
                @click="shiftDay(1)"
                ><ChevronRight
              /></Button>
              <Button
                v-if="date !== localDate()"
                type="button"
                variant="outline"
                :disabled="busy || !!entry"
                @click="selectDate(localDate())"
                >Today</Button
              >
            </div>
          </div>
          <NutritionTotals :total="day.total" prominent />
          <label v-if="dates.length" class="fit-label max-w-xs"
            >Logged dates
            <select
              class="fit-field"
              :value="dates.includes(date) ? date : ''"
              :disabled="busy || !!entry"
              @change="selectDate(($event.target as HTMLSelectElement).value)"
            >
              <option value="" disabled>Choose a date</option>
              <option
                v-for="loggedDate in dates"
                :key="loggedDate"
                :value="loggedDate"
              >
                {{ formatDate(loggedDate) }}
              </option>
            </select>
          </label>
        </section>
        <EntryEditor
          v-if="entry"
          :key="`${entry.id}-${entry.food.id}-${entry.category}`"
          :entry="entry"
          :presets="presets"
          :busy="busy"
          @save="(item) => submitEntry(item.date, () => saveFoodEntry(item))"
          @log-preset="
            (id, placement) =>
              submitEntry(placement.date, () => logFoodPreset(id, placement))
          "
          @cancel="entry = undefined"
        />
        <section
          v-for="group in day.groups"
          :key="group.category"
          class="fit-panel space-y-4"
        >
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <h2 class="font-semibold">
                {{ categoryLabels[group.category] }}
              </h2>
              <NutritionTotals
                v-if="group.entries?.length"
                :total="group.total"
                class="mt-1"
              />
            </div>
            <Button
              type="button"
              variant="outline"
              :disabled="busy || !!entry"
              @click="addEntry(group.category)"
              >Add {{ group.category }}</Button
            >
          </div>
          <p
            v-if="!group.entries?.length"
            class="text-sm text-muted-foreground"
          >
            Nothing logged in
            {{ categoryLabels[group.category]?.toLowerCase() }} on this date.
          </p>
          <ul v-else class="divide-y border-t">
            <li
              v-for="item in group.entries"
              :key="item.id"
              class="flex flex-wrap items-center justify-between gap-2 py-3"
            >
              <div>
                <h3 class="font-medium">{{ item.food.name }}</h3>
                <p class="text-sm">
                  {{ formatNumber(item.quantity) }} × {{ item.food.serving }}
                </p>
                <NutritionTotals :total="item.total" class="mt-1" />
              </div>
              <div class="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  :disabled="busy || !!entry"
                  @click="entry = item"
                  >Edit</Button
                >
                <Button
                  type="button"
                  variant="ghost"
                  :disabled="busy || !!entry"
                  @click="askDelete(item, 'entry')"
                  >Delete</Button
                >
              </div>
            </li>
          </ul>
        </section>
      </div>
      <div v-else-if="section === 'presets'" class="space-y-4">
        <form
          v-if="draft"
          class="fit-panel space-y-4"
          @submit.prevent="submitPreset"
        >
          <h2 class="text-lg font-semibold">
            {{ draft.id ? 'Edit food preset' : 'New food preset' }}
          </h2>
          <fieldset :disabled="busy"><FoodFields v-model="draft" /></fieldset>
          <div class="flex gap-2">
            <Button type="submit" :disabled="busy">Save preset</Button>
            <Button
              type="button"
              variant="outline"
              :disabled="busy"
              @click="draft = undefined"
              >Cancel</Button
            >
          </div>
        </form>
        <p
          v-if="!presets.length && !draft"
          class="fit-panel text-sm text-muted-foreground"
        >
          Save frequently logged foods to add them to any date.
        </p>
        <ul v-else class="grid gap-3 @min-[44rem]/main:grid-cols-2">
          <li
            v-for="preset in presets"
            :key="preset.id"
            class="fit-panel space-y-3"
          >
            <div>
              <h2 class="font-semibold">{{ preset.name }}</h2>
              <p class="text-sm">One serving · {{ preset.serving }}</p>
              <NutritionTotals :total="preset" class="mt-1" />
            </div>
            <div class="flex flex-wrap gap-2">
              <Button
                type="button"
                :disabled="busy || !!entry"
                @click="addEntry('meal', preset)"
                >Log preset</Button
              >
              <Button
                type="button"
                variant="outline"
                :disabled="busy || !!draft"
                @click="draft = { ...preset }"
                >Edit</Button
              >
              <Button
                type="button"
                variant="ghost"
                :disabled="busy || !!draft"
                @click="askDelete(preset, 'preset')"
                >Delete</Button
              >
            </div>
          </li>
        </ul>
      </div>
    </template>
    <ConfirmAction
      :action="confirmation"
      @close="confirmation = null"
      @confirm="confirm"
    />
  </div>
</template>
