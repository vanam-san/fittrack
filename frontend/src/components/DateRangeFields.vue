<script setup lang="ts">
import { useId } from 'vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const from = defineModel<string>('from', { required: true })
const to = defineModel<string>('to', { required: true })
const id = useId()
</script>

<template>
  <div class="grid w-full min-w-0 grid-cols-2 gap-3 @min-[32rem]/main:w-auto">
    <div class="min-w-0 space-y-1.5">
      <Label :for="`${id}-from`">From</Label>
      <Input
        :id="`${id}-from`"
        v-model="from"
        type="date"
        class="text-sm"
        :max="to"
        required
      />
    </div>
    <div class="min-w-0 space-y-1.5">
      <Label :for="`${id}-to`">To</Label>
      <Input
        :id="`${id}-to`"
        v-model="to"
        type="date"
        class="text-sm"
        :min="from"
        required
      />
    </div>
  </div>
</template>
