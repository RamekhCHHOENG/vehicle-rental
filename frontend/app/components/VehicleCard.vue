<script setup lang="ts">
import type { Vehicle } from '~/types'

const props = defineProps<{ vehicle: Vehicle }>()
const photoUrl = usePhotoUrl()

const cover = computed(() => photoUrl(props.vehicle.photos?.[0]?.file_path))
</script>

<template>
  <NuxtLink :to="`/vehicles/${vehicle.id}`" class="block">
    <article class="dk-card dk-card-interactive overflow-hidden">
      <img
        :src="cover"
        :alt="`${vehicle.make} ${vehicle.model}`"
        class="w-full h-44 object-cover"
      >
      <div class="p-5 space-y-1.5">
        <div class="flex items-baseline justify-between gap-2">
          <h3 class="text-[15px] font-semibold tracking-tight truncate">
            {{ vehicle.make }} {{ vehicle.model }}
          </h3>
          <p class="text-[15px] font-semibold text-apple-500 dark:text-apple-400 whitespace-nowrap">
            ${{ vehicle.price_per_day }}<span class="text-[11px] font-normal text-neutral-400">/day</span>
          </p>
        </div>
        <p class="text-[12px] text-neutral-500 dark:text-neutral-400 capitalize">
          {{ vehicle.type }} · {{ vehicle.year }} · {{ vehicle.transmission }}
          <template v-if="vehicle.seats"> · {{ vehicle.seats }} seats</template>
        </p>
        <p class="text-[12px] text-neutral-400 dark:text-neutral-500">{{ vehicle.location }}</p>
      </div>
    </article>
  </NuxtLink>
</template>
