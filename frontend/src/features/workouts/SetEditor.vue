<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { tracksReps } from './tracking'
import type { SetInput, WorkoutSet } from '@/services/workouts'

const props = defineProps<{
  set: WorkoutSet
  index: number
  tracking: string
  busy: boolean
}>()
const emit = defineEmits<{ save: [input: SetInput] }>()
const draft = reactive({ ...props.set })
const dirty = computed(() =>
  Object.keys(props.set).some((key) => {
    const field = key as keyof WorkoutSet
    return draft[field] !== props.set[field]
  }),
)

watch(
  () => props.set,
  (saved, previous) => {
    if (
      Object.keys(saved).some((key) => {
        const field = key as keyof WorkoutSet
        return saved[field] !== previous[field]
      })
    )
      Object.assign(draft, saved)
  },
)

function submit(event: SubmitEvent) {
  const action = (event.submitter as HTMLButtonElement | null)?.value
  const done = action === 'draft' ? props.set.done : action !== 'undo'
  const { weight, weightUnit, reps, durationSeconds, distanceM, notes } = draft
  emit('save', {
    weight,
    weightUnit,
    reps,
    durationSeconds,
    distanceM,
    notes,
    done,
  })
}
</script>

<template>
  <form class="rounded-xl border px-3 py-2.5" @submit.prevent="submit">
    <fieldset :disabled="busy" class="space-y-2">
      <legend class="px-0 text-sm font-semibold">
        Set {{ index + 1 }}
        <span v-if="set.done" class="ml-2 font-normal text-primary"
          >Logged</span
        >
      </legend>
      <div class="flex flex-wrap items-end gap-2">
        <label v-if="tracking === 'weight_reps'" class="fit-label w-22"
          >Weight ({{ set.weightUnit }})
          <input
            v-model.number="draft.weight"
            class="fit-field"
            type="number"
            min="0"
            step="any"
            required
          />
        </label>
        <label v-if="tracksReps(tracking)" class="fit-label w-18"
          >Reps
          <input
            v-model.number="draft.reps"
            class="fit-field"
            type="number"
            min="0"
            max="999"
            required
          />
        </label>
        <label v-else class="fit-label w-26"
          >Duration (sec)
          <input
            v-model.number="draft.durationSeconds"
            class="fit-field"
            type="number"
            min="0"
            max="86400"
            required
          />
        </label>
        <label v-if="tracking === 'distance_duration'" class="fit-label w-24"
          >Distance (m)
          <input
            v-model.number="draft.distanceM"
            class="fit-field"
            type="number"
            min="0"
            max="1000000"
            step="any"
            required
          />
        </label>
        <label class="fit-label min-w-32 flex-1"
          >Set notes
          <input v-model="draft.notes" class="fit-field" maxlength="1000"
        /></label>
      </div>
      <div class="flex flex-wrap gap-2">
        <Button v-if="!set.done" type="submit" value="log" :disabled="busy"
          >Log set</Button
        >
        <Button
          v-if="dirty"
          type="submit"
          value="draft"
          variant="outline"
          :disabled="busy"
          >{{ set.done ? 'Save changes' : 'Save draft' }}</Button
        >
        <Button
          v-if="set.done"
          type="submit"
          value="undo"
          variant="ghost"
          :disabled="busy"
          >Mark incomplete</Button
        >
      </div>
    </fieldset>
  </form>
</template>
