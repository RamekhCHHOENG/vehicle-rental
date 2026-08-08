<script setup lang="ts">
import type { Vehicle } from '~/types'

const api = useApi()

useSeoMeta({
  title: 'Browse vehicles',
  description:
    'Inspected cars and motorbikes for rent in Cambodia. Filter by type, location and price — what you see is the full price.'
})

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

const hasFilters = computed(() =>
  filters.type !== 'all' || !!filters.location || !!filters.max_price
)

function clearFilters() {
  filters.type = 'all'
  filters.location = ''
  filters.max_price = undefined
}
</script>

<template>
  <UContainer class="py-12">
    <p class="eyebrow mb-3">Every listing inspected</p>
    <h1 class="display-md text-[34px] sm:text-[40px] mb-8">Browse vehicles</h1>

    <div class="flex flex-wrap items-center gap-3 mb-9">
      <USelect v-model="filters.type" :items="typeOptions" value-key="value" class="w-40" />
      <UInput v-model="filters.location" placeholder="Location" icon="i-lucide-map-pin" class="w-52" />
      <UInput v-model.number="filters.max_price" type="number" placeholder="Max $/day" class="w-36" />
      <UButton v-if="hasFilters" variant="ghost" color="neutral" size="sm" @click="clearFilters">
        Clear
      </UButton>
      <p v-if="vehicles?.length" class="ml-auto text-[12px] text-[var(--ui-text-dimmed)] numeric">
        {{ vehicles.length }} vehicle{{ vehicles.length === 1 ? '' : 's' }}
      </p>
    </div>

    <p v-if="status === 'pending'" class="py-20 text-center text-[13px] text-[var(--ui-text-dimmed)]">
      Loading…
    </p>

    <!-- An empty screen is an invitation to act. -->
    <div v-else-if="!vehicles?.length" class="dk-card py-16 px-8 text-center">
      <p class="display-md text-[19px] mb-2">No vehicles match this search</p>
      <p class="text-[13.5px] text-[var(--ui-text-muted)] max-w-sm mx-auto">
        Try a wider price range or a different location. New vehicles appear here
        as soon as they pass inspection.
      </p>
      <UButton v-if="hasFilters" class="mt-6" variant="soft" color="neutral" @click="clearFilters">
        Clear filters
      </UButton>
    </div>

    <div v-else class="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
      <VehicleCard v-for="v in vehicles" :key="v.id" :vehicle="v" />
    </div>
  </UContainer>
</template>
