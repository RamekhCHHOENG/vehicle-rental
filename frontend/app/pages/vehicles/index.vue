<script setup lang="ts">
import type { Vehicle } from '~/types'

const api = useApi()

useSeoMeta({
  title: 'Browse vehicles',
  description:
    'Inspected cars and motorbikes for rent in Cambodia. Filter by type, location and price — what you see is the full price.'
})

const { data: metadata } = await useMetadata()

const filters = reactive({
  type: 'all',
  province_id: 'all',
  make_id: 'all',
  min_seats: undefined as number | undefined,
  max_price: undefined as number | undefined,
  features: [] as string[]
})

const typeOptions = computed(() => [
  { label: 'All types', value: 'all' },
  ...(metadata.value?.vehicle_types ?? []).map(t => ({ label: `${t.label}s`, value: t.value }))
])

const provinceOptions = computed(() => [
  { label: 'All provinces', value: 'all' },
  ...(metadata.value?.provinces ?? []).map(p => ({ label: p.name_en, value: p.id }))
])

// Only makes that sell the type being browsed, so the list shortens as the
// search narrows instead of offering makes with nothing to show. The flags are
// derived server-side — the models themselves are far too many to ship here.
const makeOptions = computed(() => [
  { label: 'All makes', value: 'all' },
  ...(metadata.value?.makes ?? [])
    .filter(m => filters.type === 'all'
      || (filters.type === 'car' ? m.has_cars : m.has_motorbikes))
    .map(m => ({ label: m.name, value: m.id }))
])

const featureOptions = computed(() =>
  (metadata.value?.features ?? []).filter(f => filters.type === 'all' || !f.applies_to || f.applies_to === filters.type)
)

const { data: vehicles, status, refresh } = await useAsyncData(
  'vehicles-browse',
  () => api<Vehicle[]>('/api/vehicles', {
    query: {
      type: filters.type === 'all' ? undefined : filters.type,
      province_id: filters.province_id === 'all' ? undefined : filters.province_id,
      make_id: filters.make_id === 'all' ? undefined : filters.make_id,
      min_seats: filters.min_seats || undefined,
      max_price: filters.max_price || undefined,
      features: filters.features.length ? filters.features.join(',') : undefined
    }
  })
)

// Narrowing the type can strand a make that no longer applies.
watch(() => filters.type, () => {
  if (!makeOptions.value.some(o => o.value === filters.make_id)) filters.make_id = 'all'
  const stillValid = new Set(featureOptions.value.map(f => f.code))
  filters.features = filters.features.filter(code => stillValid.has(code))
})

watch(filters, () => refresh())

function toggleFeature(code: string) {
  const i = filters.features.indexOf(code)
  if (i === -1) filters.features.push(code)
  else filters.features.splice(i, 1)
}

const hasFilters = computed(() =>
  filters.type !== 'all' || filters.province_id !== 'all' || filters.make_id !== 'all'
  || !!filters.min_seats || !!filters.max_price || filters.features.length > 0
)

function clearFilters() {
  filters.type = 'all'
  filters.province_id = 'all'
  filters.make_id = 'all'
  filters.min_seats = undefined
  filters.max_price = undefined
  filters.features = []
}
</script>

<template>
  <UContainer class="py-12">
    <p class="eyebrow mb-3">Every listing inspected</p>
    <h1 class="display-md text-[34px] sm:text-[40px] mb-8">Browse vehicles</h1>

    <div class="flex flex-wrap items-center gap-3 mb-4">
      <USelect v-model="filters.type" :items="typeOptions" value-key="value" class="w-36" />
      <USelectMenu
        v-model="filters.province_id"
        :items="provinceOptions"
        value-key="value"
        icon="i-lucide-map-pin"
        :search-input="{ placeholder: 'Search provinces…' }"
        class="w-48"
      />
      <USelectMenu
        v-model="filters.make_id"
        :items="makeOptions"
        value-key="value"
        :search-input="{ placeholder: 'Search makes…' }"
        class="w-40"
      />
      <UInput v-model.number="filters.min_seats" type="number" placeholder="Min seats" class="w-32" />
      <UInput v-model.number="filters.max_price" type="number" placeholder="Max $/day" class="w-32" />
      <UButton v-if="hasFilters" variant="ghost" color="neutral" size="sm" @click="clearFilters">
        Clear
      </UButton>
      <p v-if="vehicles?.length" class="ml-auto text-[12px] text-[var(--ui-text-dimmed)] numeric">
        {{ vehicles.length }} vehicle{{ vehicles.length === 1 ? '' : 's' }}
      </p>
    </div>

    <div v-if="featureOptions.length" class="flex flex-wrap gap-2 mb-9">
      <button
        v-for="feature in featureOptions"
        :key="feature.id"
        type="button"
        class="dk-chip"
        :class="filters.features.includes(feature.code) ? 'dk-chip-on' : ''"
        @click="toggleFeature(feature.code)"
      >
        <UIcon v-if="feature.icon" :name="feature.icon" class="size-3.5" />
        {{ feature.name }}
      </button>
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
