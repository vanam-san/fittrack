import {
  Load,
  Save,
} from '../../bindings/github.com/vanam-san/fittrack/internal/desktop/settingsservice'
import type { Preferences } from '../../bindings/github.com/vanam-san/fittrack/internal/settings/models'

export type { Preferences }

export const loadSettings = Load
export const saveSettings = Save
