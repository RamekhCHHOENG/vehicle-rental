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
</script>

<template>
  <UContainer class="py-8">
    <div v-if="error" class="text-center py-16">
      <p class="text-xl text-gray-500">Vehicle not found.</p>
      <UButton to="/vehicles" variant="outline" class="mt-4">Back to browse</UButton>
    </div>

    <div v-else-if="vehicle" class="grid grid-cols-1 lg:grid-cols-5 gap-8">
      <!-- Photos + details -->
      <div class="lg:col-span-3 space-y-4">
        <img
          :src="photoUrl(vehicle.photos?.[selectedPhoto]?.file_path)"
          :alt="`${vehicle.make} ${vehicle.model}`"
          class="w-full h-80 object-cover rounded-lg"
        >
        <div v-if="vehicle.photos?.length > 1" class="flex gap-2 overflow-x-auto">
          <img
            v-for="(photo, i) in vehicle.photos"
            :key="photo.id"
            :src="photoUrl(photo.file_path)"
            class="w-20 h-14 object-cover rounded cursor-pointer border-2"
            :class="i === selectedPhoto ? 'border-primary' : 'border-transparent'"
            @click="selectedPhoto = i"
          >
        </div>

        <div>
          <div class="flex items-center gap-3">
            <h1 class="text-3xl font-bold">{{ vehicle.make }} {{ vehicle.model }} {{ vehicle.year }}</h1>
            <StatusBadge v-if="vehicle.status !== 'approved'" :status="vehicle.status" />
          </div>
          <p v-if="data && data.review_count > 0" class="mt-1 text-amber-500">
            ★ {{ data.avg_rating.toFixed(1) }}
            <span class="text-gray-500">({{ data.review_count }} review{{ data.review_count > 1 ? 's' : '' }})</span>
          </p>

          <div class="flex flex-wrap gap-2 mt-4">
            <UBadge variant="subtle" class="capitalize">{{ vehicle.type }}</UBadge>
            <UBadge variant="subtle" class="capitalize">{{ vehicle.transmission }}</UBadge>
            <UBadge v-if="vehicle.seats" variant="subtle">{{ vehicle.seats }} seats</UBadge>
            <UBadge variant="subtle">📍 {{ vehicle.location }}</UBadge>
          </div>

          <p v-if="vehicle.description" class="mt-4 text-gray-600 dark:text-gray-300 whitespace-pre-line">
            {{ vehicle.description }}
          </p>
        </div>
      </div>

      <!-- Booking panel (booking form arrives in the booking phase) -->
      <div class="lg:col-span-2">
        <BookingForm :vehicle="vehicle" />
      </div>
    </div>
  </UContainer>
</template>
