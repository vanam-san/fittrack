// @vitest-environment happy-dom

import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App.vue'
import { loadSettings, saveSettings } from '@/services/settings'
import {
  ChooseInput,
  Replace,
} from '../bindings/github.com/vanam-san/fittrack/internal/desktop/datafilesservice'

vi.mock('@/services/settings', () => ({
  loadSettings: vi.fn(),
  saveSettings: vi.fn(),
}))

vi.mock(
  '../bindings/github.com/vanam-san/fittrack/internal/desktop/datafilesservice',
  () => ({
    Backup: vi.fn(),
    ChooseInput: vi.fn(),
    Export: vi.fn(),
    Replace: vi.fn(),
  }),
)

vi.mock('@/services/workouts', () => ({
  loadWorkouts: vi.fn().mockResolvedValue({
    exercises: [],
    routines: [],
    history: [],
    active: null,
  }),
}))

vi.mock('@/services/nutrition', () => ({
  loadNutritionDay: vi.fn().mockResolvedValue({
    date: '2026-11-23',
    groups: [],
    total: { calories: 0, protein: 0, carbohydrates: 0, fat: 0 },
  }),
}))

vi.mock('@/services/weight', () => ({
  loadWeightOverview: vi.fn().mockResolvedValue({ latest: null, goalKg: null }),
}))

const firstRun = {
  preferences: {
    displayName: '',
    weightUnit: 'kg',
    theme: 'system',
    calorieGoal: 0,
    proteinGoal: 0,
    carbohydrateGoal: 0,
    fatGoal: 0,
  },
  setupComplete: false,
  databasePath: 'fittrack.sqlite3',
  backupPath: 'backups',
}

describe('FitTrack settings', () => {
  beforeEach(() => {
    vi.mocked(loadSettings).mockReset()
    vi.mocked(saveSettings).mockReset()
  })
  afterEach(() => document.documentElement.classList.remove('dark'))

  it('keeps the first-run form open after a failed save and preserves the input for retry', async () => {
    vi.mocked(loadSettings).mockResolvedValue(firstRun)
    vi.mocked(saveSettings)
      .mockRejectedValueOnce(new Error('Could not save settings.'))
      .mockResolvedValueOnce({
        ...firstRun,
        preferences: {
          ...firstRun.preferences,
          displayName: 'Sam',
          weightUnit: 'lb',
        },
        setupComplete: true,
      })
    const app = mount(App)
    try {
      await flushPromises()
      expect(app.get('h1').text()).toBe('Set up FitTrack')
      await app.get<HTMLInputElement>('#display-name').setValue('Sam')
      await app.get<HTMLSelectElement>('#weight-unit').setValue('lb')

      await app.get('form').trigger('submit')
      await flushPromises()
      expect(app.get('[role="alert"]').text()).toBe('Could not save settings.')
      expect(app.get<HTMLInputElement>('#display-name').element.value).toBe(
        'Sam',
      )

      await app.get('form').trigger('submit')
      await flushPromises()
      expect(saveSettings).toHaveBeenLastCalledWith({
        ...firstRun.preferences,
        displayName: 'Sam',
        weightUnit: 'lb',
      })
      expect(app.get('h1').text()).toBe('Dashboard')
    } finally {
      app.unmount()
    }
  })

  it('shows a local-storage error when startup cannot read settings', async () => {
    vi.mocked(loadSettings).mockRejectedValue(
      new Error('Local storage is unavailable.'),
    )
    const app = mount(App)
    try {
      await flushPromises()
      expect(app.get('h1').text()).toBe('Local storage unavailable')
      expect(app.get('[role="alert"]').text()).toBe(
        'Local storage is unavailable.',
      )
      expect(app.find('form').exists()).toBe(false)
    } finally {
      app.unmount()
    }
  })

  it('requires confirmation before restoring and refreshes preferences after recovery', async () => {
    vi.mocked(loadSettings)
      .mockResolvedValueOnce(firstRun)
      .mockResolvedValueOnce({
        ...firstRun,
        setupComplete: true,
        preferences: { ...firstRun.preferences, displayName: 'Recovered' },
      })
    vi.mocked(ChooseInput).mockResolvedValue('C:\\backups\\example.sqlite3')
    vi.mocked(Replace).mockResolvedValue('C:\\backups\\before.sqlite3')

    const app = mount(App, { attachTo: document.body })
    const restore = () =>
      app
        .findAll('button')
        .find((button) => button.text() === 'Restore backup')!
    const dialog = () => document.querySelector('[role="dialog"]')
    const dialogButton = (label: string) =>
      Array.from(dialog()!.querySelectorAll('button')).find(
        (button) => button.textContent?.trim() === label,
      )!

    try {
      await flushPromises()
      expect(app.text()).toContain('Default backup folder')
      await restore().trigger('click')
      await flushPromises()
      expect(dialog()?.textContent).toContain('example.sqlite3')
      expect(Replace).not.toHaveBeenCalled()
      dialogButton('Cancel').click()
      await flushPromises()
      expect(Replace).not.toHaveBeenCalled()

      await restore().trigger('click')
      await flushPromises()
      dialogButton('Restore backup').click()
      await flushPromises()
      expect(Replace).toHaveBeenCalledOnce()
      expect(Replace).toHaveBeenCalledWith(
        'restore',
        'C:\\backups\\example.sqlite3',
      )
      expect(app.get('h1').text()).toBe('Settings')
      expect(app.get<HTMLInputElement>('#display-name').element.value).toBe(
        'Recovered',
      )
      expect(app.text()).toContain('Backup restored.')
    } finally {
      app.unmount()
    }
  })
})
