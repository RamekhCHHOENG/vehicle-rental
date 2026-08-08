<script setup lang="ts">
import type { Vehicle } from '~/types'

const api = useApi()

const filters = reactive({
  type: 'all',
  location: '',
  max_price: undefined as number | undefined
})

const typeOptions = [
  { label: 'All types', value: 'all' },
  { label: 'Cars', value: 'car' },
  { label: 'Motorbikes', value: 'motorbike' }
]

const { data: vehicles, status, refresh } = await useAsyncData(
  'vehicles-browse',
  () => api<Vehicle[]>('/api/vehicles', {
    query: {
      type: filters.type === 'all' ? undefined : filters.type,
      location: filters.location || undefined,
      max_price: filters.max_price || undefined
    }
  })
)

watch(filters, () => refresh())
</script>

<template>
  <UContainer class="py-8">
    <h1 class="text-3xl font-bold mb-6">Browse vehicles</h1>

    <div class="flex flex-wrap gap-3 mb-8">
      <USelect v-model="filters.type" :items="typeOptions" value-key="value" class="w-40" />
      <UInput v-model="filters.location" placeholder="Location…" class="w-48" />
      <UInput v-model.number="filters.max_price" type="number" placeholder="Max $/day" class="w-32" />
    </div>

    <div v-if="status === 'pending'" class="text-center py-16 text-gray-500">Loading…</div>

    <div v-else-if="!vehicles?.length" class="text-center py-16 text-gray-500">
      No vehicles match your search yet. Try different filters.
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      <VehicleCard v-for="v in vehicles" :key="v.id" :vehicle="v" />
    </div>
  </UContainer>
</template>
