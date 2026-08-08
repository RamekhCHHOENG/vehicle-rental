<script setup lang="ts">
const props = defineProps<{ status: string }>()

const colorMap: Record<string, any> = {
  pending: 'warning',
  approved: 'success',
  rejected: 'error',
  requested: 'warning',
  confirmed: 'success',
  cancelled: 'neutral',
  completed: 'info'
}

const color = computed(() => colorMap[props.status] ?? 'neutral')
</script>

<template>
  <UBadge :color="color" variant="subtle" class="capitalize">{{ status }}</UBadge>
</template>
