<script setup lang="ts">
import { Activity, LayoutDashboard, Settings2 } from 'lucide-vue-next'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import DashboardView from '@/features/dashboard/DashboardView.vue'
import SettingsForm from '@/features/settings/SettingsForm.vue'
import {
  loadSettings,
  saveSettings,
  type Preferences,
} from '@/services/settings'

const state = ref<Awaited<ReturnType<typeof loadSettings>>>()
const page = ref<'dashboard' | 'settings'>('dashboard')
const main = ref<HTMLElement>()
const loadError = ref('')
const saveError = ref('')
const saving = ref(false)
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
  } catch (error) {
    loadError.value = message(error, 'FitTrack could not load your settings.')
  }
}

async function save(preferences: Preferences) {
  if (saving.value) return
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
  if (saving.value && next === 'settings') return
  page.value = next
  saveError.value = ''
  saved.value = false
  await nextTick()
  main.value?.focus()
}

onMounted(() => {
  systemTheme.addEventListener('change', onSystemTheme)
  void load()
})
onUnmounted(() => systemTheme.removeEventListener('change', onSystemTheme))
</script>

<template>
  <div class="min-h-screen md:flex">
    <aside class="fit-sidebar">
      <div class="flex items-center gap-3 font-semibold tracking-tight">
        <span class="fit-brand"
          ><Activity :size="20" aria-hidden="true"
        /></span>
        FitTrack
      </div>
      <nav
        v-if="state?.setupComplete"
        aria-label="Main navigation"
        class="flex gap-1 md:flex-col"
      >
        <button
          type="button"
          class="fit-nav"
          :aria-current="page === 'dashboard' ? 'page' : undefined"
          :disabled="saving"
          @click="navigate('dashboard')"
        >
          <LayoutDashboard :size="17" aria-hidden="true" /> Dashboard
        </button>
        <button
          type="button"
          class="fit-nav"
          :aria-current="page === 'settings' ? 'page' : undefined"
          :disabled="saving"
          @click="navigate('settings')"
        >
          <Settings2 :size="17" aria-hidden="true" /> Settings
        </button>
      </nav>
    </aside>

    <main
      ref="main"
      tabindex="-1"
      class="mx-auto w-full min-w-0 max-w-6xl flex-1 p-5 outline-none sm:p-8 lg:p-12"
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
      />
      <div v-else class="max-w-2xl space-y-6">
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
        <div class="fit-panel">
          <SettingsForm
            :preferences="state.preferences"
            :first-run="!state.setupComplete"
            :saving="saving"
            :error="saveError"
            :saved="saved"
            @save="save"
          />
        </div>
        <section
          v-if="state.setupComplete"
          class="fit-panel space-y-2"
          aria-labelledby="storage-title"
        >
          <h2 id="storage-title" class="font-semibold">Local database</h2>
          <code class="block break-all text-sm text-muted-foreground">{{
            state.databasePath
          }}</code>
        </section>
      </div>
    </main>
  </div>
</template>
