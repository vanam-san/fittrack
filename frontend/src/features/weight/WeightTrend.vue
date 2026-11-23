<script setup lang="ts">
import MetricChart from '@/components/MetricChart.vue'
import { displayWeight } from '@/lib/weight'
import type { Measurement } from '@/services/weight'

defineProps<{
  measurements: Measurement[]
  unit: string
  from: string
  to: string
  table?: boolean
}>()
</script>

<template>
  <MetricChart
    title="Weight progression"
    type="line"
    :unit="unit"
    :from="from"
    :to="to"
    :dates="measurements.map(({ date }) => date)"
    :series="[
      {
        name: 'Weight',
        values: measurements.map(({ weightKg }) =>
          displayWeight(weightKg, unit),
        ),
      },
    ]"
    :empty="
      measurements.length ? undefined : 'No measurements in this date range.'
    "
    :table="table"
  />
</template>
