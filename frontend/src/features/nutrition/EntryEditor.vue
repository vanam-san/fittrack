<script setup lang="ts">
import { computed, ref } from 'vue'
import { Button } from '@/components/ui/button'
import FoodFields from './FoodFields.vue'
import NutritionTotals from './NutritionTotals.vue'
import { categoryLabels } from './food'
import type { Entry, Food, LogPlacement } from '@/services/nutrition'

const props = defineProps<{ entry: Entry; presets: Food[]; busy: boolean }>()
const emit = defineEmits<{
  save: [entry: Entry]
  logPreset: [id: number, placement: LogPlacement]
  cancel: []
}>()
const placement = ref<LogPlacement>({
  date: props.entry.date,
  category: props.entry.category,
  quantity: props.entry.quantity,
})
const food = ref<Food>({ ...props.entry.food })
const presetId = ref(props.entry.id === 0 ? props.entry.food.id : 0)
const preset = computed(() =>
  props.presets.find((item) => item.id === presetId.value),
)

function submit() {
  if (preset.value) {
    emit('logPreset', preset.value.id, { ...placement.value })
  } else {
    emit('save', {
      ...props.entry,
      ...placement.value,
      food: { ...food.value },
    })
  }
}
</script>

<template>
  <form class="fit-panel space-y-4" @submit.prevent="submit">
    <h2 class="text-lg font-semibold">
      {{ entry.id ? 'Edit entry' : 'Add food or drink' }}
    </h2>
    <fieldset :disabled="busy" class="space-y-4">
      <label v-if="!entry.id && presets.length" class="fit-label"
        >Food source
        <select v-model.number="presetId" class="fit-field">
          <option :value="0">Enter manually</option>
          <option v-for="item in presets" :key="item.id" :value="item.id">
            {{ item.name }} · {{ item.serving }}
          </option>
        </select>
      </label>
      <div v-if="preset" class="rounded-xl border px-3 py-2">
        <p class="text-sm font-semibold">
          {{ preset.name }} · {{ preset.serving }} per serving
        </p>
        <NutritionTotals :total="preset" class="mt-1" />
      </div>
      <FoodFields v-else v-model="food" />
      <div class="grid gap-3 sm:grid-cols-3">
        <label class="fit-label"
          >Date
          <input
            v-model="placement.date"
            class="fit-field"
            type="date"
            required
            min="0001-01-01"
            max="9999-12-31"
          />
        </label>
        <label class="fit-label"
          >Category
          <select v-model="placement.category" class="fit-field">
            <option
              v-for="(label, value) in categoryLabels"
              :key="value"
              :value="value"
            >
              {{ label }}
            </option>
          </select>
        </label>
        <label class="fit-label"
          >Servings consumed
          <input
            v-model.number="placement.quantity"
            class="fit-field"
            type="number"
            min="0.000001"
            max="1000"
            step="any"
            required
          />
        </label>
      </div>
    </fieldset>
    <div class="flex gap-2">
      <Button type="submit" :disabled="busy">{{
        entry.id ? 'Save changes' : 'Log entry'
      }}</Button>
      <Button
        type="button"
        variant="outline"
        :disabled="busy"
        @click="emit('cancel')"
        >Cancel</Button
      >
    </div>
  </form>
</template>
