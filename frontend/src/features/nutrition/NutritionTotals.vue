<script setup lang="ts">
import { formatNumber } from '@/lib/format'
import type { Nutrients } from '@/services/nutrition'
import { nutrientFields } from './food'

defineProps<{ total: Nutrients; prominent?: boolean }>()
</script>

<template>
  <dl v-if="prominent" class="grid grid-cols-2 gap-x-4 gap-y-3 sm:grid-cols-4">
    <div v-for="field in nutrientFields" :key="field.key">
      <dt class="text-sm text-muted-foreground">{{ field.label }}</dt>
      <dd class="mt-1 text-xl font-semibold tabular-nums">
        {{ formatNumber(total[field.key]) }}
        <span class="text-sm font-normal text-muted-foreground">{{
          field.unit
        }}</span>
      </dd>
    </div>
  </dl>
  <p v-else class="text-sm text-muted-foreground">
    {{ formatNumber(total.calories) }} kcal ·
    {{ formatNumber(total.protein) }} g protein ·
    {{ formatNumber(total.carbohydrates) }} g carbs ·
    {{ formatNumber(total.fat) }} g fat
  </p>
</template>
