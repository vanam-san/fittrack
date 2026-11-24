<script setup lang="ts">
import {
  ChartNoAxesCombined,
  Dumbbell,
  LayoutDashboard,
  Scale,
  Settings2,
  Utensils,
} from 'lucide-vue-next'
import {
  computed,
  defineAsyncComponent,
  nextTick,
  onMounted,
  onUnmounted,
  ref,
  watch,
} from 'vue'
import DashboardView from '@/features/dashboard/DashboardView.vue'
import NutritionView from '@/features/nutrition/NutritionView.vue'
import DataFiles from '@/features/settings/DataFiles.vue'
import SettingsForm from '@/features/settings/SettingsForm.vue'
import WorkoutsView from '@/features/workouts/WorkoutsView.vue'
import brandIcon from '../../build/appicon.png'
import {
  loadSettings,
  saveSettings,
  type Preferences,
} from '@/services/settings'

const state = ref<Awaited<ReturnType<typeof loadSettings>>>()
const WeightView = defineAsyncComponent(
  () => import('@/features/weight/WeightView.vue'),
)
const AnalyticsView = defineAsyncComponent(
  () => import('@/features/analytics/AnalyticsView.vue'),
)
const navigation = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'workouts', label: 'Workouts', icon: Dumbbell },
  { id: 'nutrition', label: 'Nutrition', icon: Utensils },
  { id: 'weight', label: 'Weight', icon: Scale },
  { id: 'analytics', label: 'Analytics', icon: ChartNoAxesCombined },
  { id: 'settings', label: 'Settings', icon: Settings2 },
] as const
const page = ref<(typeof navigation)[number]['id']>('dashboard')
const main = ref<HTMLElement>()
const loadError = ref('')
const restoreError = ref('')
const saveError = ref('')
const saving = ref(false)
const dataBusy = ref(false)
const saved = ref(false)

const systemTheme = window.matchMedia('(prefers-color-scheme: dark)')
const systemDark = ref(systemTheme.matches)
const dark = computed(() => {
  const theme = state.value?.preferences.theme ?? 'system'
  return theme === 'dark' || (theme === 'system' && systemDark.value)
})
watch(
  dark,
  (value) => document.documentElement.classList.toggle('dark', value),
  { immediate: true },
)

function onSystemTheme(event: MediaQueryListEvent) {
  systemDark.value = event.matches
}

function message(error: unknown, fallback: string) {
  return error instanceof Error && error.message ? error.message : fallback
}

async function load() {
  loadError.value = ''
  try {
    state.value = await loadSettings()
    restoreError.value = state.value.restoreError ?? ''
    if (restoreError.value) page.value = 'settings'
  } catch (error) {
    loadError.value = message(error, 'FitTrack could not load your settings.')
  }
}

async function save(preferences: Preferences) {
  if (saving.value || dataBusy.value) return
  saving.value = true
  saveError.value = ''
  saved.value = false
  const firstRun = !state.value?.setupComplete
  try {
    state.value = await saveSettings(preferences)
    if (firstRun) await navigate('dashboard')
    else saved.value = true
  } catch (error) {
    saveError.value = message(error, 'FitTrack could not save your settings.')
  } finally {
    saving.value = false
  }
}

async function navigate(next: typeof page.value) {
  if (dataBusy.value || (saving.value && next === 'settings')) return
  page.value = next
  saveError.value = ''
  saved.value = false
  // Focus the newly rendered page so keyboard navigation begins in its content.
  await nextTick()
  main.value?.focus({ preventScroll: true })
  window.scrollTo(0, 0)
}

async function reloadAfterReplacement() {
  // Import and restore can replace preferences and first-run state.
  page.value = 'settings'
  saveError.value = ''
  saved.value = false
  state.value = await loadSettings()
  restoreError.value = ''
}

onMounted(() => {
  systemTheme.addEventListener('change', onSystemTheme)
  void load()
})
onUnmounted(() => systemTheme.removeEventListener('change', onSystemTheme))
</script>

<template>
  <div class="min-h-screen lg:flex">
    <aside class="fit-sidebar">
      <div class="flex items-center gap-3 font-semibold tracking-tight">
        <img :src="brandIcon" alt="" width="36" height="36" />
        FitTrack
      </div>
      <nav
        v-if="state?.setupComplete"
        aria-label="Main navigation"
        class="fit-main-nav"
      >
        <button
          v-for="item in navigation"
          :key="item.id"
          type="button"
          class="fit-nav"
          :aria-current="page === item.id ? 'page' : undefined"
          :disabled="saving || dataBusy"
          @click="navigate(item.id)"
        >
          <component
            :is="item.icon"
            :size="17"
            class="shrink-0"
            aria-hidden="true"
          />
          {{ item.label }}
        </button>
      </nav>
    </aside>

    <main
      ref="main"
      tabindex="-1"
      class="@container/main mx-auto w-full min-w-0 max-w-6xl flex-1 p-4 outline-none sm:p-6 xl:p-8"
    >
      <div v-if="loadError" class="fit-panel max-w-lg space-y-4">
        <h1 class="text-2xl font-semibold">Local storage unavailable</h1>
        <p role="alert" class="text-sm text-destructive">{{ loadError }}</p>
      </div>
      <p v-else-if="!state" role="status" class="text-muted-foreground">
        Opening FitTrack…
      </p>
      <DashboardView
        v-else-if="state.setupComplete && page === 'dashboard'"
        :preferences="state.preferences"
        @navigate="navigate"
      />
      <WorkoutsView
        v-else-if="state.setupComplete && page === 'workouts'"
        :weight-unit="state.preferences.weightUnit"
      />
      <NutritionView v-else-if="state.setupComplete && page === 'nutrition'" />
      <WeightView
        v-else-if="state.setupComplete && page === 'weight'"
        :weight-unit="state.preferences.weightUnit"
      />
      <AnalyticsView
        v-else-if="state.setupComplete && page === 'analytics'"
        :weight-unit="state.preferences.weightUnit"
      />
      <div v-else class="fit-page max-w-2xl">
        <header class="space-y-2">
          <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">
            {{ state.setupComplete ? 'Settings' : 'Set up FitTrack' }}
          </h1>
          <p class="text-muted-foreground">
            {{
              state.setupComplete
                ? 'Manage the preferences stored on this device.'
                : 'Choose your preferences. You can change them anytime.'
            }}
          </p>
        </header>
        <fieldset class="fit-panel" :disabled="dataBusy">
          <SettingsForm
            :preferences="state.preferences"
            :first-run="!state.setupComplete"
            :saving="saving"
            :error="saveError"
            :saved="saved"
            @save="save"
          />
        </fieldset>
        <p v-if="restoreError" role="alert" class="text-sm text-destructive">
          {{ restoreError }}
        </p>
        <DataFiles
          v-model:busy="dataBusy"
          :database-path="state.databasePath"
          :backup-path="state.backupPath"
          :saving="saving"
          :reload="reloadAfterReplacement"
        />
      </div>
    </main>
  </div>
</template>
