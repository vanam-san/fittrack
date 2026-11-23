<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Plus } from 'lucide-vue-next'
import ConfirmAction from '@/components/ConfirmAction.vue'
import { Button } from '@/components/ui/button'
import { errorMessage, formatDate } from '@/lib/format'
import ExerciseLibrary from './ExerciseLibrary.vue'
import RoutineEditor from './RoutineEditor.vue'
import WorkoutSession from './WorkoutSession.vue'
import WorkoutCalendar from './WorkoutCalendar.vue'
import { formatDuration } from './tracking'
import {
  addSet,
  changeWorkoutStatus,
  createExercise,
  deleteRoutine,
  duplicateRoutine,
  loadWorkouts,
  loadWorkoutSession,
  saveRoutine,
  saveSet,
  skipRest,
  startWorkout,
  type ExerciseInput,
  type Overview,
  type Routine,
  type Session,
  type SetInput,
} from '@/services/workouts'

const props = defineProps<{ weightUnit: string }>()
const sections = [
  { key: 'routines', label: 'Routines' },
  { key: 'library', label: 'Exercise library' },
  { key: 'history', label: 'History' },
] as const
const section = ref<(typeof sections)[number]['key'] | 'live'>('routines')
const overview = ref<Overview>()
const historyDetail = ref<Session>()
const showCalendar = ref(false)
const historyDate = ref('')
const history = computed(() => overview.value?.history ?? [])
const workoutDates = computed(
  () => new Set(history.value.map((session) => session.date)),
)
const visibleHistory = computed(() =>
  historyDate.value
    ? history.value.filter((session) => session.date === historyDate.value)
    : history.value,
)
const historyIndex = computed(() =>
  visibleHistory.value.findIndex(
    (session) => session.id === historyDetail.value?.id,
  ),
)
const newerWorkout = computed(
  () => visibleHistory.value[historyIndex.value - 1],
)
const olderWorkout = computed(() =>
  historyIndex.value >= 0
    ? visibleHistory.value[historyIndex.value + 1]
    : undefined,
)

function toggleCalendar() {
  showCalendar.value = !showCalendar.value
  if (!showCalendar.value) historyDate.value = ''
}
const edit = ref<Routine>()
const library = ref<InstanceType<typeof ExerciseLibrary>>()
const busy = ref(false)
const error = ref('')
const notice = ref('')
const confirmation = ref<{
  title: string
  description: string
  label: string
  destructive?: boolean
  apply: () => Promise<void>
} | null>(null)
const exerciseNames = computed(
  () =>
    new Map(
      overview.value?.exercises?.map((exercise) => [
        exercise.id,
        exercise.name,
      ]),
    ),
)

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

async function refresh() {
  overview.value = await loadWorkouts(props.weightUnit)
}

function load() {
  void run(async () => {
    await refresh()
    if (overview.value?.active) section.value = 'live'
  })
}

function saveExercise(input: ExerciseInput, addAnother: boolean) {
  void run(async () => {
    await createExercise(input)
    library.value?.saved(addAnother)
    await refresh()
    notice.value = 'Exercise added to the library.'
  })
}

function submitRoutine(routine: Routine) {
  void run(async () => {
    await saveRoutine(routine)
    edit.value = undefined
    await refresh()
    notice.value = 'Routine saved.'
  })
}

function copyRoutine(routine: Routine) {
  void run(async () => {
    const id = await duplicateRoutine(routine.id)
    await refresh()
    edit.value = overview.value?.routines?.find((row) => row.id === id)
    notice.value = 'Routine duplicated.'
  })
}

function confirmDelete(routine: Routine) {
  confirmation.value = {
    title: 'Delete routine?',
    description: `“${routine.name}” will be removed from your routines. Saved workouts keep their original plan.`,
    label: 'Delete routine',
    destructive: true,
    apply: async () => {
      await deleteRoutine(routine.id)
      if (edit.value?.id === routine.id) edit.value = undefined
      await refresh()
      notice.value = 'Routine deleted.'
    },
  }
}

function start(routine: Routine) {
  void run(async () => {
    await startWorkout(routine.id)
    edit.value = undefined
    await refresh()
    section.value = 'live'
  })
}

function changeStatus(target: 'paused' | 'active' | 'completed') {
  const session = overview.value?.active
  if (!session) return
  if (target !== 'completed') {
    updateLive((id) => changeWorkoutStatus(id, target))
    return
  }
  const hasIncomplete = session.exercises?.some((exercise) =>
    exercise.sets?.some((set) => !set.done),
  )
  confirmation.value = {
    title: 'Finish workout?',
    description: hasIncomplete
      ? 'Any incomplete sets will be marked as skipped in your workout history.'
      : 'This workout will be saved to your session history.',
    label: 'Finish workout',
    apply: async () => {
      await changeWorkoutStatus(session.id, target)
      await refresh()
      section.value = 'history'
      historyDate.value = ''
      historyDetail.value = await loadWorkoutSession(
        session.id,
        props.weightUnit,
      )
      notice.value = 'Workout completed.'
    },
  }
}

function confirm() {
  const action = confirmation.value
  confirmation.value = null
  if (action) void run(action.apply)
}

function updateLive(action: (id: number) => Promise<void>) {
  const current = overview.value
  if (!current?.active) return
  const id = current.active.id
  void run(async () => {
    await action(id)
    current.active = await loadWorkoutSession(id, props.weightUnit)
  })
}

function logSet(exerciseId: number, setId: number, input: SetInput) {
  updateLive((id) => saveSet(id, exerciseId, setId, input))
}

function showHistory(id: number) {
  void run(async () => {
    historyDetail.value = await loadWorkoutSession(id, props.weightUnit)
  })
}

onMounted(load)
</script>

<template>
  <div class="space-y-5">
    <header class="flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">
        Workouts
      </h1>
      <Button
        v-if="overview && section === 'routines' && !edit"
        type="button"
        :disabled="busy"
        @click="edit = { id: 0, name: '', exercises: [] }"
      >
        <Plus aria-hidden="true" /> New routine
      </Button>
    </header>
    <p v-if="error" role="alert" class="fit-panel text-sm text-destructive">
      {{ error }}
    </p>
    <p v-if="notice" role="status" class="text-sm text-primary">{{ notice }}</p>
    <div v-if="!overview" class="fit-panel text-sm text-muted-foreground">
      <p v-if="busy" role="status">Loading workouts…</p>
      <Button v-else type="button" variant="outline" @click="load"
        >Retry</Button
      >
    </div>
    <template v-else>
      <nav
        aria-label="Workout sections"
        class="flex flex-wrap gap-1 border-b pb-2"
      >
        <Button
          v-for="item in sections"
          :key="item.key"
          type="button"
          :variant="section === item.key ? 'secondary' : 'ghost'"
          :aria-current="section === item.key ? 'page' : undefined"
          :disabled="busy"
          @click="section = item.key"
          >{{ item.label }}</Button
        >
        <Button
          v-if="overview.active"
          type="button"
          :variant="section === 'live' ? 'secondary' : 'ghost'"
          :aria-current="section === 'live' ? 'page' : undefined"
          :disabled="busy"
          @click="section = 'live'"
          >Current workout</Button
        >
      </nav>

      <div v-if="section === 'routines'" class="space-y-4">
        <RoutineEditor
          v-if="edit"
          :key="edit.id"
          :routine="edit"
          :exercises="overview.exercises ?? []"
          :busy="busy"
          @cancel="edit = undefined"
          @save="submitRoutine"
        />
        <p
          v-if="!overview.routines?.length && !edit"
          class="fit-panel text-sm text-muted-foreground"
        >
          Create a routine to plan exercises, sets, and rest before you start
          training.
        </p>
        <ul v-else class="grid gap-3 xl:grid-cols-2">
          <li
            v-for="routine in overview.routines"
            :key="routine.id"
            class="fit-panel space-y-3"
          >
            <h2 class="font-semibold">{{ routine.name }}</h2>
            <ol class="space-y-1 text-sm text-muted-foreground">
              <li v-for="(plan, index) in routine.exercises" :key="index">
                {{ index + 1 }}. {{ exerciseNames.get(plan.exerciseId) }} ·
                {{ plan.plannedSets }} sets<span v-if="plan.plannedReps">
                  × {{ plan.plannedReps }}</span
                >
              </li>
            </ol>
            <div class="flex flex-wrap gap-2">
              <Button
                type="button"
                :disabled="busy || !!overview.active"
                @click="start(routine)"
                >Start workout</Button
              >
              <Button
                type="button"
                variant="outline"
                :disabled="busy"
                @click="edit = routine"
                >Edit</Button
              >
              <Button
                type="button"
                variant="outline"
                :disabled="busy"
                @click="copyRoutine(routine)"
                >Duplicate</Button
              >
              <Button
                type="button"
                variant="ghost"
                :disabled="busy"
                @click="confirmDelete(routine)"
                >Delete</Button
              >
            </div>
          </li>
        </ul>
      </div>

      <ExerciseLibrary
        v-else-if="section === 'library'"
        ref="library"
        :exercises="overview.exercises ?? []"
        :busy="busy"
        @save="saveExercise"
      />

      <div v-else-if="section === 'history'" class="space-y-4">
        <template v-if="historyDetail">
          <div class="flex flex-wrap gap-2">
            <Button
              type="button"
              variant="outline"
              :disabled="busy"
              @click="historyDetail = undefined"
              >Back to history</Button
            >
            <Button
              type="button"
              variant="outline"
              :disabled="busy || !newerWorkout"
              @click="newerWorkout && showHistory(newerWorkout.id)"
              >Newer workout</Button
            >
            <Button
              type="button"
              variant="outline"
              :disabled="busy || !olderWorkout"
              @click="olderWorkout && showHistory(olderWorkout.id)"
              >Older workout</Button
            >
          </div>
          <WorkoutSession :session="historyDetail" :busy="busy" />
        </template>
        <div v-show="!historyDetail" class="space-y-4">
          <Button
            type="button"
            variant="outline"
            :aria-expanded="showCalendar"
            :disabled="busy"
            @click="toggleCalendar"
            >{{ showCalendar ? 'Hide calendar' : 'Browse calendar' }}</Button
          >
          <WorkoutCalendar
            v-if="showCalendar"
            v-model="historyDate"
            :workout-dates="workoutDates"
            :busy="busy"
          />
          <h2 v-if="historyDate" class="font-semibold">
            {{ formatDate(historyDate) }}
          </h2>
          <p
            v-if="!visibleHistory.length"
            class="fit-panel text-sm text-muted-foreground"
          >
            {{
              historyDate
                ? 'No completed workouts on this date.'
                : 'Completed workouts will appear here.'
            }}
          </p>
          <ul v-else class="space-y-2">
            <li
              v-for="session in visibleHistory"
              :key="session.id"
              class="fit-panel flex flex-wrap items-center justify-between gap-3"
            >
              <div>
                <h2 class="font-semibold">{{ session.name }}</h2>
                <p class="text-sm text-muted-foreground">
                  {{ formatDate(session.date) }} ·
                  {{ formatDuration(Math.floor(session.elapsedMs / 1000)) }}
                </p>
              </div>
              <Button
                type="button"
                variant="outline"
                :disabled="busy"
                @click="showHistory(session.id)"
                >View workout</Button
              >
            </li>
          </ul>
        </div>
      </div>

      <WorkoutSession
        v-else-if="overview.active"
        :key="overview.active.id"
        :session="overview.active"
        :busy="busy"
        @status="changeStatus"
        @add="(id) => updateLive((sessionId) => addSet(sessionId, id))"
        @save="logSet"
        @skip="updateLive(skipRest)"
      />
    </template>
    <ConfirmAction
      :action="confirmation"
      @close="confirmation = null"
      @confirm="confirm"
    />
  </div>
</template>
