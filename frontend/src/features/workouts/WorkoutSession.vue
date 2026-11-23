<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { Timer } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { formatDate } from '@/lib/format'
import SetEditor from './SetEditor.vue'
import { formatDuration, formatSet } from './tracking'
import type { Session, SetInput } from '@/services/workouts'

const props = defineProps<{ session: Session; busy: boolean }>()
const emit = defineEmits<{
  status: [target: 'paused' | 'active' | 'completed']
  add: [exerciseId: number]
  save: [exerciseId: number, setId: number, input: SetInput]
  skip: []
}>()
const now = ref(Date.now())
const loadedAt = ref(now.value)
// The service owns persisted timestamps. This clock refreshes the display between responses.
let interval: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  interval = setInterval(() => {
    now.value = Date.now()
  }, 250)
})
onUnmounted(() => clearInterval(interval))
watch(
  () => props.session,
  () => {
    loadedAt.value = now.value = Date.now()
  },
)
const active = computed(() => props.session.status === 'active')
const completed = computed(() => props.session.status === 'completed')
const elapsed = computed(() =>
  Math.floor(
    (props.session.elapsedMs +
      (active.value ? Math.max(0, now.value - loadedAt.value) : 0)) /
      1000,
  ),
)
const remaining = computed(() => {
  const ms = active.value
    ? props.session.restUntil - now.value
    : props.session.pausedRestMs
  return Math.ceil(Math.max(0, ms) / 1000)
})
const hasTimer = computed(
  () => props.session.restUntil > 0 || props.session.pausedRestMs > 0,
)
</script>

<template>
  <div class="space-y-4">
    <section class="fit-panel space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-xl font-semibold">{{ session.name }}</h2>
          <p class="mt-1 text-sm text-muted-foreground">
            {{ formatDate(session.date) }} ·
            {{ completed ? 'Completed' : active ? 'In progress' : 'Paused' }} ·
            {{ formatDuration(elapsed) }}
          </p>
        </div>
        <div v-if="!completed" class="flex flex-wrap gap-2">
          <Button
            v-if="active"
            type="button"
            variant="outline"
            :disabled="busy"
            @click="emit('status', 'paused')"
            >Pause workout</Button
          >
          <Button
            v-else
            type="button"
            :disabled="busy"
            @click="emit('status', 'active')"
            >Resume workout</Button
          >
          <Button
            type="button"
            :variant="active ? 'default' : 'outline'"
            :disabled="busy"
            @click="emit('status', 'completed')"
            >Finish workout</Button
          >
        </div>
      </div>
      <div
        v-if="!completed"
        class="flex flex-wrap items-center gap-3 rounded-xl bg-accent px-3 py-2 text-sm text-accent-foreground"
      >
        <Timer :size="17" aria-hidden="true" />
        <span class="font-medium">Rest timer</span>
        <span v-if="remaining" role="timer" class="font-mono tabular-nums"
          >{{ formatDuration(remaining) }}{{ active ? '' : ' paused' }}</span
        >
        <span v-else role="status">{{
          hasTimer
            ? 'Rest complete'
            : 'Starts when you log a set with planned rest'
        }}</span>
        <Button
          v-if="active && remaining"
          type="button"
          size="sm"
          variant="outline"
          :disabled="busy"
          @click="emit('skip')"
          >Skip rest</Button
        >
      </div>
    </section>

    <section
      v-for="(exercise, index) in session.exercises"
      :key="exercise.id"
      class="fit-panel space-y-3"
    >
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h3 class="font-semibold">{{ index + 1 }}. {{ exercise.name }}</h3>
          <p class="text-sm text-muted-foreground">
            {{ exercise.plannedSets }} planned sets<span
              v-if="exercise.plannedReps"
            >
              × {{ exercise.plannedReps }} reps</span
            >
            · {{ exercise.restSeconds }} sec rest
          </p>
        </div>
        <Button
          v-if="active"
          type="button"
          variant="outline"
          :disabled="busy"
          @click="emit('add', exercise.id)"
          >Add set</Button
        >
      </div>
      <p v-if="!completed" class="text-sm">
        <span class="font-medium">Previous session</span> ·
        {{
          exercise.previous?.length
            ? exercise.previous
                .map((set) => formatSet(set, exercise.tracking))
                .join(' / ')
            : 'No logged sets yet'
        }}
      </p>
      <template v-if="active">
        <SetEditor
          v-for="(set, setIndex) in exercise.sets"
          :key="set.id"
          :set="set"
          :index="setIndex"
          :tracking="exercise.tracking"
          :busy="busy"
          @save="emit('save', exercise.id, set.id, $event)"
        />
      </template>
      <ol v-else class="divide-y rounded-xl border px-3 text-sm">
        <li v-for="(set, setIndex) in exercise.sets" :key="set.id" class="py-2">
          <p class="flex flex-wrap justify-between gap-2">
            <span>Set {{ setIndex + 1 }}</span>
            <span :class="{ 'text-muted-foreground': !set.done }">
              {{
                set.done
                  ? formatSet(set, exercise.tracking)
                  : completed
                    ? 'Skipped'
                    : 'Incomplete · ' + formatSet(set, exercise.tracking)
              }}
            </span>
          </p>
          <p v-if="set.notes" class="mt-1 text-muted-foreground">
            {{ set.notes }}
          </p>
        </li>
      </ol>
    </section>
  </div>
</template>
