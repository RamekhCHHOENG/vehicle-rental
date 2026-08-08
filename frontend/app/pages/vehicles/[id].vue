<script setup lang="ts">
import type { Vehicle } from '~/types'

const route = useRoute()
const api = useApi()
const photoUrl = usePhotoUrl()

const { data, error } = await useAsyncData(
  `vehicle-${route.params.id}`,
  () => api<{ vehicle: Vehicle, avg_rating: number, review_count: number }>(`/api/vehicles/${route.params.id}`)
)

const vehicle = computed(() => data.value?.vehicle)
const selectedPhoto = ref(0)

// The spec sheet reads as a record, matching the inspection language.
const specs = computed(() => {
  if (!vehicle.value) return []
  const v = vehicle.value
  return [
    { label: 'Type', value: v.type },
    { label: 'Transmission', value: v.transmission },
    ...(v.seats ? [{ label: 'Seats', value: String(v.seats) }] : []),
    { label: 'Year', value: String(v.year) },
    { label: 'Location', value: v.location }
  ]
})
</script>

<template>
  <UContainer class="py-10">
    <div v-if="error" class="dk-card py-20 text-center">
      <p class="display-md text-[20px] mb-2">This vehicle isn't available</p>
      <p class="text-[13.5px] text-[var(--ui-text-muted)] mb-6">
        It may have been taken down, or the link is wrong.
      </p>
      <UButton to="/vehicles">Browse vehicles</UButton>
    </div>

    <div v-else-if="vehicle" class="grid lg:grid-cols-5 gap-10">
      <!-- Photos + record -->
      <div class="lg:col-span-3 space-y-5">
        <div class="relative">
          <img
            :src="photoUrl(vehicle.photos?.[selectedPhoto]?.file_path)"
            :alt="`${vehicle.make} ${vehicle.model}`"
            class="w-full h-[340px] object-cover rounded-[var(--r-media)]"
          >
          <VerifiedSeal
            v-if="vehicle.status === 'approved'"
            class="absolute left-4 bottom-4"
            :date="vehicle.created_at"
            on-photo
            large
          />
        </div>

        <div v-if="vehicle.photos?.length > 1" class="flex gap-2 overflow-x-auto pb-1">
          <button
            v-for="(photo, i) in vehicle.photos"
            :key="photo.id"
            class="rounded-[var(--r-control)] overflow-hidden border-2 transition-colors shrink-0"
            :class="i === selectedPhoto ? 'border-saffron-400' : 'border-transparent'"
            :aria-label="`Show photo ${i + 1}`"
            @click="selectedPhoto = i"
          >
            <img :src="photoUrl(photo.file_path)" class="w-24 h-16 object-cover" alt="">
          </button>
        </div>

        <div>
          <div class="flex items-center gap-3 flex-wrap">
            <h1 class="display-md text-[32px]">
              {{ vehicle.make }} {{ vehicle.model }}
            </h1>
            <StatusBadge v-if="vehicle.status !== 'approved'" :status="vehicle.status" />
          </div>

          <p v-if="data && data.review_count > 0" class="mt-2 text-[13px]">
            <span class="text-saffron-500">★</span>
            <span class="numeric font-semibold">{{ data.avg_rating.toFixed(1) }}</span>
            <span class="text-[var(--ui-text-muted)]">
              from {{ data.review_count }} rental{{ data.review_count > 1 ? 's' : '' }}
            </span>
          </p>

          <p v-if="vehicle.description" class="mt-5 text-[14.5px] leading-relaxed text-[var(--ui-text-muted)] whitespace-pre-line max-w-2xl">
            {{ vehicle.description }}
          </p>
        </div>

        <div class="dk-card p-6 max-w-md">
          <p class="eyebrow mb-2">Vehicle record</p>
          <ul>
            <li v-for="spec in specs" :key="spec.label" class="record-row">
              <span class="text-[12.5px] text-[var(--ui-text-muted)]">{{ spec.label }}</span>
              <span class="text-[13px] font-medium capitalize">{{ spec.value }}</span>
            </li>
          </ul>
        </div>
      </div>

      <!-- Booking -->
      <div class="lg:col-span-2">
        <div class="lg:sticky lg:top-20">
          <BookingForm :vehicle="vehicle" />
        </div>
      </div>
    </div>
  </UContainer>
</template>
