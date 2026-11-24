<script setup lang="ts">
import { computed, ref } from 'vue'
import { Plus, Search } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { trackingLabels } from './tracking'
import type { Exercise, ExerciseInput } from '@/services/workouts'

const props = defineProps<{ exercises: Exercise[]; busy: boolean }>()
const emit = defineEmits<{
  save: [input: ExerciseInput, addAnother: boolean]
}>()
const search = ref('')
const category = ref('')
const draft = ref<ExerciseInput>()
const collator = new Intl.Collator(undefined, {
  sensitivity: 'base',
  numeric: true,
})
const conditioning = new Set(['Burpee', 'Jump Rope', 'Stair Climber'])

function exerciseCategory(exercise: Exercise) {
  if (
    exercise.builtIn &&
    (exercise.tracking === 'distance_duration' ||
      conditioning.has(exercise.name))
  ) {
    return 'Cardio and conditioning'
  }
  return exercise.primaryMuscle.trim() || 'Uncategorized'
}

const categories = computed(() => {
  const groups = new Map<string, { label: string; exercises: Exercise[] }>()
  for (const exercise of props.exercises) {
    const label = exerciseCategory(exercise)
    const key = label.toLocaleLowerCase()
    const group = groups.get(key) ?? { label, exercises: [] }
    group.exercises.push(exercise)
    groups.set(key, group)
  }
  return [...groups].sort(([, a], [, b]) => collator.compare(a.label, b.label))
})

const matching = computed(() => {
  const query = search.value.trim().toLocaleLowerCase()
  return categories.value.flatMap(([key, { label, exercises }]) => {
    if (category.value && category.value !== key) return []
    const matches = exercises.filter((exercise) =>
      [
        exercise.name,
        label,
        exercise.primaryMuscle,
        exercise.secondaryMuscles,
        exercise.equipment,
      ].some((value) => value.toLocaleLowerCase().includes(query)),
    )
    matches.sort((a, b) => collator.compare(a.name, b.name))
    return matches.length ? [{ key, label, exercises: matches }] : []
  })
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

function saved(addAnother: boolean) {
  if (addAnother) {
    create()
  } else {
    draft.value = undefined
  }
  search.value = ''
  category.value = ''
}

function submit(event: SubmitEvent) {
  if (!draft.value || props.busy) return
  emit(
    'save',
    { ...draft.value },
    event.submitter?.getAttribute('value') === 'another',
  )
}

defineExpose({ saved })
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-end gap-3">
      <label class="relative block w-full @min-[32rem]/main:max-w-xs">
        <Search
          class="pointer-events-none absolute top-1/2 start-3 size-4 -translate-y-1/2 text-muted-foreground"
          aria-hidden="true"
        />
        <input
          v-model="search"
          class="fit-field exercise-search"
          type="search"
          aria-label="Search exercises"
          placeholder="Search exercises"
        />
      </label>
      <label class="fit-label w-full @min-[32rem]/main:w-auto">
        Category
        <select v-model="category" class="fit-select">
          <option value="">All categories</option>
          <option v-for="[key, group] in categories" :key="key" :value="key">
            {{ group.label }} ({{ group.exercises.length }})
          </option>
        </select>
      </label>
      <Button type="button" :disabled="busy" @click="create"
        ><Plus aria-hidden="true" /> Custom exercise</Button
      >
    </div>
    <form v-if="draft" class="fit-panel space-y-4" @submit.prevent="submit">
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
      <div class="flex flex-wrap gap-2">
        <Button type="submit" :disabled="busy">Save exercise</Button>
        <Button type="submit" value="another" variant="outline" :disabled="busy"
          >Save and add another</Button
        >
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
      No exercises match your filters.
    </p>
    <section v-for="group in matching" :key="group.key" class="space-y-3">
      <h2 class="text-lg font-semibold">
        {{ group.label }}
        <span class="ml-1 text-sm font-normal text-muted-foreground">
          ({{ group.exercises.length }})
        </span>
      </h2>
      <ul class="grid gap-3 @min-[44rem]/main:grid-cols-2">
        <li
          v-for="exercise in group.exercises"
          :key="exercise.id"
          class="fit-panel space-y-2"
        >
          <h3 class="font-semibold">{{ exercise.name }}</h3>
          <p class="text-sm text-muted-foreground">
            {{ exercise.primaryMuscle }} · {{ exercise.equipment }} ·
            {{ trackingLabels[exercise.tracking] }}
          </p>
          <p
            v-if="exercise.secondaryMuscles || exercise.builtIn"
            class="text-sm"
          >
            Secondary muscles · {{ exercise.secondaryMuscles || 'None listed' }}
          </p>
          <p v-if="exercise.notes" class="text-sm whitespace-pre-wrap">
            {{ exercise.notes }}
          </p>
        </li>
      </ul>
    </section>
  </div>
</template>

<style scoped>
.exercise-search {
  padding-inline-start: 2.25rem;
}
</style>
