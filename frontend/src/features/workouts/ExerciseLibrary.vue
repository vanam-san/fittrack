<script setup lang="ts">
import { computed, ref } from 'vue'
import { Plus, Search } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { trackingLabels } from './tracking'
import type { Exercise, ExerciseInput } from '@/services/workouts'

const props = defineProps<{ exercises: Exercise[]; busy: boolean }>()
const emit = defineEmits<{ save: [input: ExerciseInput] }>()
const search = ref('')
const draft = ref<ExerciseInput>()
const matching = computed(() => {
  const query = search.value.trim().toLocaleLowerCase()
  return props.exercises.filter((exercise) =>
    [
      exercise.name,
      exercise.primaryMuscle,
      exercise.secondaryMuscles,
      exercise.equipment,
    ].some((value) => value.toLocaleLowerCase().includes(query)),
  )
})

function create() {
  draft.value = {
    name: '',
    primaryMuscle: '',
    secondaryMuscles: '',
    equipment: '',
    tracking: 'weight_reps',
    notes: '',
  }
}

function saved() {
  draft.value = undefined
  search.value = ''
}

defineExpose({ saved })
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <label class="relative block w-full max-w-sm">
        <Search
          class="absolute top-2.5 left-3 size-4 text-muted-foreground"
          aria-hidden="true"
        />
        <input
          v-model="search"
          class="fit-field pl-9"
          type="search"
          aria-label="Search exercises"
          placeholder="Search by exercise, muscle, or equipment"
        />
      </label>
      <Button type="button" :disabled="busy" @click="create"
        ><Plus aria-hidden="true" /> Custom exercise</Button
      >
    </div>
    <form
      v-if="draft"
      class="fit-panel space-y-4"
      @submit.prevent="emit('save', { ...draft })"
    >
      <h2 class="text-lg font-semibold">New exercise</h2>
      <fieldset :disabled="busy" class="grid gap-3 sm:grid-cols-2">
        <label class="fit-label"
          >Name
          <input
            v-model="draft.name"
            class="fit-field"
            required
            maxlength="100"
        /></label>
        <label class="fit-label"
          >Equipment
          <input
            v-model="draft.equipment"
            class="fit-field"
            required
            maxlength="80"
        /></label>
        <label class="fit-label"
          >Primary muscle
          <input
            v-model="draft.primaryMuscle"
            class="fit-field"
            required
            maxlength="80"
        /></label>
        <label class="fit-label"
          >Secondary muscles
          <input
            v-model="draft.secondaryMuscles"
            class="fit-field"
            maxlength="160"
        /></label>
        <label class="fit-label"
          >Tracking type
          <select v-model="draft.tracking" class="fit-field">
            <option
              v-for="(label, value) in trackingLabels"
              :key="value"
              :value="value"
            >
              {{ label }}
            </option>
          </select>
        </label>
        <label class="fit-label sm:col-span-2"
          >Instructions or notes
          <textarea
            v-model="draft.notes"
            class="fit-field min-h-20"
            maxlength="2000"
          />
        </label>
      </fieldset>
      <div class="flex gap-2">
        <Button type="submit" :disabled="busy">Save exercise</Button>
        <Button
          type="button"
          variant="outline"
          :disabled="busy"
          @click="draft = undefined"
          >Cancel</Button
        >
      </div>
    </form>
    <p v-if="!matching.length" class="fit-panel text-sm text-muted-foreground">
      No exercises match your search.
    </p>
    <ul v-else class="grid gap-3 lg:grid-cols-2">
      <li
        v-for="exercise in matching"
        :key="exercise.id"
        class="fit-panel space-y-2"
      >
        <h2 class="font-semibold">{{ exercise.name }}</h2>
        <p class="text-sm text-muted-foreground">
          {{ exercise.primaryMuscle }} · {{ exercise.equipment }} ·
          {{ trackingLabels[exercise.tracking] }}
        </p>
        <p v-if="exercise.secondaryMuscles" class="text-sm">
          Secondary muscles · {{ exercise.secondaryMuscles }}
        </p>
        <p v-if="exercise.notes" class="text-sm whitespace-pre-wrap">
          {{ exercise.notes }}
        </p>
      </li>
    </ul>
  </div>
</template>
