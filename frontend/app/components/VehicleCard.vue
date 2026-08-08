<script setup lang="ts">
import type { Vehicle } from '~/types'

const props = defineProps<{ vehicle: Vehicle }>()
const photoUrl = usePhotoUrl()

const cover = computed(() => photoUrl(props.vehicle.photos?.[0]?.file_path))
</script>

<template>
  <NuxtLink :to="`/vehicles/${vehicle.id}`" class="block group">
    <article class="dk-card dk-card-interactive overflow-hidden h-full flex flex-col">
      <div class="relative">
        <img
          :src="cover"
          :alt="`${vehicle.make} ${vehicle.model}`"
          class="w-full h-44 object-cover"
          loading="lazy"
        >
        <!-- One seal per surface. -->
        <VerifiedSeal class="absolute left-3 bottom-3" :date="vehicle.created_at" on-photo />
      </div>

      <div class="p-5 flex-1 flex flex-col">
        <div class="flex items-baseline justify-between gap-3">
          <h3 class="display-md text-[16px] truncate">
            {{ vehicle.make }} {{ vehicle.model }}
          </h3>
          <p class="font-display numeric text-[16px] font-bold whitespace-nowrap">
            ${{ vehicle.price_per_day }}<span class="text-[11px] font-medium text-[var(--ui-text-dimmed)]">/day</span>
          </p>
        </div>

        <p class="mt-2 text-[12.5px] text-[var(--ui-text-muted)] capitalize">
          {{ vehicle.type }} · {{ vehicle.year }} · {{ vehicle.transmission }}
          <template v-if="vehicle.seats"> · {{ vehicle.seats }} seats</template>
        </p>
        <p class="mt-auto pt-3 text-[12.5px] text-[var(--ui-text-dimmed)]">
          {{ vehicle.location }}
        </p>
      </div>
    </article>
  </NuxtLink>
</template>
