<script setup lang="ts">
import type { Vehicle } from '~/types'

const props = defineProps<{ vehicle: Vehicle }>()
const photoUrl = usePhotoUrl()

const cover = computed(() => photoUrl(props.vehicle.photos?.[0]?.file_path))
</script>

<template>
  <NuxtLink :to="`/vehicles/${vehicle.id}`" class="block group">
    <UCard class="overflow-hidden transition-shadow group-hover:shadow-lg">
      <template #header>
        <img
          :src="cover"
          :alt="`${vehicle.make} ${vehicle.model}`"
          class="w-full h-44 object-cover -m-4 mb-0"
          style="width: calc(100% + 2rem)"
        >
      </template>

      <div class="space-y-1">
        <div class="flex items-center justify-between">
          <h3 class="font-semibold">{{ vehicle.make }} {{ vehicle.model }}</h3>
          <span class="text-primary font-bold">${{ vehicle.price_per_day }}/day</span>
        </div>
        <p class="text-sm text-gray-500 capitalize">
          {{ vehicle.type }} · {{ vehicle.year }} · {{ vehicle.transmission }}
          <template v-if="vehicle.seats"> · {{ vehicle.seats }} seats</template>
        </p>
        <p class="text-sm text-gray-500">📍 {{ vehicle.location }}</p>
      </div>
    </UCard>
  </NuxtLink>
</template>
