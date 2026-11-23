import { formatNumber } from '@/lib/format'
import type { WorkoutSet } from '@/services/workouts'

export const trackingLabels: Record<string, string> = {
  weight_reps: 'Weight and reps',
  reps: 'Reps',
  duration: 'Duration',
  distance_duration: 'Distance and duration',
}

export function tracksReps(tracking: string) {
  return tracking === 'reps' || tracking === 'weight_reps'
}

export function formatSet(set: WorkoutSet, tracking: string) {
  switch (tracking) {
    case 'weight_reps':
      return `${formatNumber(set.weight)} ${set.weightUnit} × ${set.reps}`
    case 'reps':
      return `${set.reps} reps`
    case 'distance_duration':
      return `${formatNumber(set.distanceM)} m in ${set.durationSeconds} sec`
    default:
      return `${set.durationSeconds} sec`
  }
}

export function formatDuration(seconds: number) {
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`
}
