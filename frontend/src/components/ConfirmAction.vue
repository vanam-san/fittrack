<script setup lang="ts">
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

defineProps<{
  action: {
    title: string
    description: string
    label: string
    destructive?: boolean
  } | null
}>()
const emit = defineEmits<{ close: []; confirm: [] }>()
</script>

<template>
  <Dialog :open="!!action" @update:open="!$event && emit('close')">
    <DialogContent v-if="action">
      <DialogHeader>
        <DialogTitle>{{ action.title }}</DialogTitle>
        <DialogDescription>{{ action.description }}</DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose as-child
          ><Button type="button" variant="outline">Cancel</Button></DialogClose
        >
        <Button
          type="button"
          :variant="action.destructive ? 'destructive' : 'default'"
          @click="emit('confirm')"
        >
          {{ action.label }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
