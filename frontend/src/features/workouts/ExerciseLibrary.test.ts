// @vitest-environment happy-dom

import { mount } from '@vue/test-utils'
import { expect, it } from 'vitest'
import ExerciseLibrary from './ExerciseLibrary.vue'
import type { Exercise } from '@/services/workouts'

it('groups exercises alphabetically and combines category and text filters', async () => {
  const exercise = (
    id: number,
    name: string,
    data: Partial<Exercise> = {},
  ): Exercise => ({
    id,
    name,
    primaryMuscle: 'Back',
    secondaryMuscles: 'Biceps',
    equipment: 'Cable machine',
    tracking: 'weight_reps',
    builtIn: true,
    notes: 'Pull with control.',
    ...data,
  })
  const library = mount(ExerciseLibrary, {
    props: {
      busy: false,
      exercises: [
        exercise(1, 'Seated Cable Row'),
        exercise(2, 'Running', {
          primaryMuscle: 'Legs',
          tracking: 'distance_duration',
        }),
        exercise(3, 'Lat Pulldown'),
        exercise(4, 'Custom Row', { primaryMuscle: ' back ', builtIn: false }),
      ],
    },
  })
  try {
    expect(
      library
        .findAll('h2')
        .map((heading) => heading.text().replace(/\s+/g, ' ')),
    ).toEqual(['Back (3)', 'Cardio and conditioning (1)'])
    expect(library.findAll('h3').map((heading) => heading.text())).toEqual([
      'Custom Row',
      'Lat Pulldown',
      'Seated Cable Row',
      'Running',
    ])
    await library.get('select').setValue('back')
    await library.get('input[type="search"]').setValue('row')
    expect(library.findAll('h3').map((heading) => heading.text())).toEqual([
      'Custom Row',
      'Seated Cable Row',
    ])
    await library.get('select').setValue('cardio and conditioning')
    expect(library.text()).toContain('No exercises match your filters.')
  } finally {
    library.unmount()
  }
})
