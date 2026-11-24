<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import ConfirmAction from '@/components/ConfirmAction.vue'
import { Button } from '@/components/ui/button'
import { androidFiles, pickAndroidDocument } from './androidFiles'
import {
  Backup,
  ChooseInput,
  Export,
  Replace,
} from '../../../bindings/github.com/vanam-san/fittrack/internal/desktop/datafilesservice'

const props = defineProps<{
  databasePath: string
  backupPath: string
  saving: boolean
  reload: () => Promise<void>
}>()
const busy = defineModel<boolean>('busy', { required: true })
type Replacement = 'import' | 'restore'
const pending = ref<{ operation: Replacement; path: string }>()
const message = ref('')
const error = ref('')
let trigger: HTMLButtonElement | null = null

const confirmation = computed(() => {
  if (!pending.value) return null
  const restoring = pending.value.operation === 'restore'
  const filename = pending.value.path.split(/[/\\]/).at(-1)
  return {
    title: restoring ? 'Restore this backup?' : 'Import these records?',
    description: `${filename} will replace your current records and preferences. FitTrack saves a recovery backup first.`,
    label: restoring ? 'Restore backup' : 'Import records',
    destructive: true,
  }
})

async function run(work: () => Promise<void>) {
  if (busy.value || props.saving) return
  busy.value = true
  message.value = ''
  error.value = ''
  try {
    await work()
  } catch (failure) {
    error.value =
      failure instanceof Error
        ? failure.message
        : 'FitTrack could not complete the file operation.'
  } finally {
    busy.value = false
    if (!pending.value) await refocus()
  }
}

async function refocus() {
  await nextTick()
  trigger?.focus()
}

function exportRecords(event: MouseEvent, format: 'csv' | 'json') {
  trigger = event.currentTarget as HTMLButtonElement
  return run(async () => {
    const path = await Export(format)
    if (!path) return
    if (androidFiles) {
      if (await pickAndroidDocument('save', path))
        message.value = 'Records exported.'
    } else {
      message.value = `Exported to ${path}.`
    }
  })
}

function createBackup(event: MouseEvent) {
  trigger = event.currentTarget as HTMLButtonElement
  return run(async () => {
    const path = await Backup()
    if (!path) return
    if (androidFiles) {
      if (await pickAndroidDocument('save', path))
        message.value = 'Backup exported.'
    } else {
      message.value = `Backup saved to ${path}.`
    }
  })
}

function choose(event: MouseEvent, operation: Replacement) {
  trigger = event.currentTarget as HTMLButtonElement
  return run(async () => {
    const path = androidFiles
      ? await pickAndroidDocument('open')
      : await ChooseInput(operation)
    if (path) pending.value = { operation, path }
  })
}

function cancel() {
  if (pending.value) androidFiles?.discard(pending.value.path)
  pending.value = undefined
  void refocus()
}

function replace() {
  const selected = pending.value
  if (!selected) return
  pending.value = undefined
  return run(async () => {
    let checkpoint: string
    try {
      checkpoint = await Replace(selected.operation, selected.path)
    } finally {
      androidFiles?.discard(selected.path)
    }
    // The replacement may contain different settings or require first-run setup.
    await props.reload()
    const result =
      selected.operation === 'restore'
        ? 'Backup restored.'
        : 'Records imported.'
    message.value = androidFiles
      ? `${result} A recovery backup was saved in private app storage.`
      : `${result} Recovery backup saved to ${checkpoint}.`
  })
}
</script>

<template>
  <section
    class="fit-panel space-y-4"
    aria-labelledby="storage-title"
    :aria-busy="busy"
  >
    <h2 id="storage-title" class="font-semibold">Data and recovery</h2>
    <dl v-if="!androidFiles" class="space-y-3 text-sm">
      <div>
        <dt class="font-medium">Local database</dt>
        <dd>
          <code class="break-all text-muted-foreground">{{
            databasePath
          }}</code>
        </dd>
      </div>
      <div>
        <dt class="font-medium">Default backup folder</dt>
        <dd>
          <code class="break-all text-muted-foreground">{{ backupPath }}</code>
        </dd>
      </div>
    </dl>
    <p class="text-sm text-muted-foreground">
      Export or import all records and preferences as CSV or JSON. Use a
      database backup for full recovery.
    </p>
    <div class="flex flex-wrap gap-2">
      <Button
        type="button"
        variant="outline"
        :disabled="busy || saving"
        @click="exportRecords($event, 'csv')"
      >
        Export CSV
      </Button>
      <Button
        type="button"
        variant="outline"
        :disabled="busy || saving"
        @click="exportRecords($event, 'json')"
      >
        Export JSON
      </Button>
      <Button
        type="button"
        variant="outline"
        :disabled="busy || saving"
        @click="choose($event, 'import')"
      >
        Import records
      </Button>
      <Button
        type="button"
        variant="outline"
        :disabled="busy || saving"
        @click="createBackup"
      >
        {{ androidFiles ? 'Export backup' : 'Create backup' }}
      </Button>
      <Button
        type="button"
        variant="outline"
        :disabled="busy || saving"
        @click="choose($event, 'restore')"
      >
        Restore backup
      </Button>
    </div>
    <p v-if="busy" role="status" class="text-sm text-muted-foreground">
      Working…
    </p>
    <p v-if="message" role="status" class="break-all text-sm">{{ message }}</p>
    <p v-if="error" role="alert" class="break-words text-sm text-destructive">
      {{ error }}
    </p>
    <ConfirmAction :action="confirmation" @close="cancel" @confirm="replace" />
  </section>
</template>
