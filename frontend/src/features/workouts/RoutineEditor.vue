<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowDown, ArrowUp, Plus, Trash2 } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { tracksReps } from './tracking'
import type { Exercise, Plan, Routine } from '@/services/workouts'

const props = defineProps<{
  routine: Routine
  exercises: Exercise[]
  busy: boolean
}>()
const emit = defineEmits<{ save: [routine: Routine]; cancel: [] }>()
const name = ref(props.routine.name)
const plans = ref<Plan[]>(
  (props.routine.exercises ?? []).map((plan) => ({ ...plan })),
)
const exerciseId = ref(props.exercises[0]?.id ?? 0)
const library = computed(
  () => new Map(props.exercises.map((exercise) => [exercise.id, exercise])),
)

function addExercise() {
  if (!exerciseId.value || plans.value.length >= 50) return
  plans.value.push({
    exerciseId: exerciseId.value,
    plannedSets: 3,
    plannedReps: 8,
    restSeconds: 90,
  })
}

function move(index: number, step: number) {
  const plan = plans.value[index]
  const destination = index + step
  if (!plan || destination < 0 || destination >= plans.value.length) return
  plans.value.splice(index, 1)
  plans.value.splice(destination, 0, plan)
}
</script>

<template>
  <form
    class="fit-panel space-y-5"
    @submit.prevent="emit('save', { id: routine.id, name, exercises: plans })"
  >
    <h2 class="text-lg font-semibold">
      {{ routine.id ? 'Edit routine' : 'New routine' }}
    </h2>
    <fieldset :disabled="busy" class="space-y-4">
      <label class="fit-label"
        >Routine name
        <input v-model="name" class="fit-field" required maxlength="100"
      /></label>
      <ol class="space-y-3">
        <li
          v-for="(plan, index) in plans"
          :key="index"
          class="space-y-3 rounded-xl border p-3"
        >
          <div class="flex flex-wrap items-center justify-between gap-2">
            <span class="font-medium"
              >{{ index + 1 }}. {{ library.get(plan.exerciseId)?.name }}</span
            >
            <div class="flex gap-1">
              <Button
                type="button"
                variant="outline"
                size="icon-sm"
                :disabled="index === 0"
                :aria-label="'Move exercise ' + (index + 1) + ' up'"
                @click="move(index, -1)"
                ><ArrowUp
              /></Button>
              <Button
                type="button"
                variant="outline"
                size="icon-sm"
                :disabled="index === plans.length - 1"
                :aria-label="'Move exercise ' + (index + 1) + ' down'"
                @click="move(index, 1)"
                ><ArrowDown
              /></Button>
              <Button
                type="button"
                variant="destructive"
                size="icon-sm"
                :aria-label="'Remove exercise ' + (index + 1)"
                @click="plans.splice(index, 1)"
                ><Trash2
              /></Button>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
            <label class="fit-label"
              >Planned sets
              <input
                v-model.number="plan.plannedSets"
                class="fit-field"
                type="number"
                required
                min="1"
                max="50"
            /></label>
            <label
              v-if="tracksReps(library.get(plan.exerciseId)?.tracking ?? '')"
              class="fit-label"
              >Planned reps
              <input
                v-model.number="plan.plannedReps"
                class="fit-field"
                type="number"
                required
                min="1"
                max="999"
            /></label>
            <label class="fit-label"
              >Rest (sec)
              <input
                v-model.number="plan.restSeconds"
                class="fit-field"
                type="number"
                required
                min="0"
                max="3600"
            /></label>
          </div>
        </li>
      </ol>
      <div class="flex flex-wrap items-end gap-2">
        <label class="fit-label min-w-0 flex-1"
          >Exercise
          <select v-model.number="exerciseId" class="fit-field">
            <option
              v-for="exercise in exercises"
              :key="exercise.id"
              :value="exercise.id"
            >
              {{ exercise.name }}
            </option>
          </select>
        </label>
        <Button
          type="button"
          variant="outline"
          size="lg"
          :disabled="!exerciseId || plans.length >= 50"
          @click="addExercise"
          ><Plus aria-hidden="true" /> Add exercise</Button
        >
      </div>
    </fieldset>
    <div class="flex gap-2">
      <Button type="submit" :disabled="busy || !plans.length"
        >Save routine</Button
      >
      <Button
        type="button"
        variant="outline"
        :disabled="busy"
        @click="emit('cancel')"
        >Cancel</Button
      >
    </div>
  </form>
</template>
