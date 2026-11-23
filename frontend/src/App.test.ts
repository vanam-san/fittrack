// @vitest-environment happy-dom

import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App.vue'
import { loadSettings, saveSettings } from '@/services/settings'

vi.mock('@/services/settings', () => ({
  loadSettings: vi.fn(),
  saveSettings: vi.fn(),
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
}

describe('FitTrack settings', () => {
  beforeEach(() => {
    vi.resetAllMocks()
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
})
