<script setup lang="ts">
import { computed, ref } from 'vue'
import { Button } from '@/components/ui/button'
import { formatDate, localDate } from '@/lib/format'

defineProps<{ workoutDates: Set<string>; busy: boolean }>()
const selected = defineModel<string>({ required: true })
const today = localDate()
// Build the month in local time to match the date keys stored with sessions.
const initial = new Date(`${selected.value || today}T12:00:00`)
const month = ref(new Date(initial.getFullYear(), initial.getMonth(), 1, 12))
const monthLabel = computed(() =>
  month.value.toLocaleDateString(undefined, { month: 'long', year: 'numeric' }),
)
const weekdays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const days = computed(() => {
  const year = month.value.getFullYear()
  const index = month.value.getMonth()
  const count = new Date(year, index + 1, 0).getDate()
  return Array.from({ length: count }, (_, offset) => ({
    number: offset + 1,
    date: localDate(new Date(year, index, offset + 1, 12)),
  }))
})

function moveMonth(delta: number) {
  month.value = new Date(
    month.value.getFullYear(),
    month.value.getMonth() + delta,
    1,
    12,
  )
}
</script>

<template>
  <section class="fit-panel max-w-lg space-y-3" aria-label="Workout calendar">
    <div class="flex items-center justify-between gap-2">
      <Button
        type="button"
        variant="outline"
        size="sm"
        :disabled="busy"
        aria-label="Previous month"
        @click="moveMonth(-1)"
        >‹</Button
      >
      <h2 class="font-semibold" aria-live="polite">{{ monthLabel }}</h2>
      <Button
        type="button"
        variant="outline"
        size="sm"
        :disabled="busy"
        aria-label="Next month"
        @click="moveMonth(1)"
        >›</Button
      >
    </div>
    <div class="grid grid-cols-7 gap-1 text-center text-sm">
      <span
        v-for="weekday in weekdays"
        :key="weekday"
        class="text-muted-foreground"
      >
        {{ weekday }}
      </span>
      <span v-for="offset in month.getDay()" :key="offset" aria-hidden="true" />
      <Button
        v-for="day in days"
        :key="day.date"
        type="button"
        size="sm"
        :variant="
          selected === day.date
            ? 'default'
            : workoutDates.has(day.date)
              ? 'secondary'
              : 'ghost'
        "
        :aria-pressed="selected === day.date"
        :aria-current="day.date === today ? 'date' : undefined"
        :aria-label="`${formatDate(day.date)}${workoutDates.has(day.date) ? ', workout logged' : ''}`"
        :disabled="busy"
        @click="selected = day.date"
        >{{ day.number }}</Button
      >
    </div>
    <Button
      type="button"
      size="sm"
      variant="outline"
      :disabled="busy || !selected"
      @click="selected = ''"
      >Show all dates</Button
    >
  </section>
</template>
