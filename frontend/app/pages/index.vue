<script setup lang="ts">
import type { Vehicle } from '~/types'

const api = useApi()

const { data: featured } = await useAsyncData(
  'featured-vehicles',
  () => api<Vehicle[]>('/api/vehicles'),
  { transform: (vehicles) => vehicles.slice(0, 3) }
)
</script>

<template>
  <div>
    <!-- Hero -->
    <UContainer class="py-20 text-center">
      <h1 class="text-5xl font-bold tracking-tight">
        Rent a car in <span class="text-primary">3 easy steps</span>
      </h1>
      <p class="mt-4 text-lg text-gray-500 max-w-2xl mx-auto">
        Clear prices with no hidden fees. Real photos. Every vehicle verified by
        our team before it goes live.
      </p>
      <div class="mt-8 flex justify-center gap-3">
        <UButton size="xl" to="/vehicles">Browse vehicles</UButton>
        <UButton size="xl" variant="outline" color="neutral" to="/register">
          List your vehicle
        </UButton>
      </div>
    </UContainer>

    <!-- How it works -->
    <div class="bg-gray-50 dark:bg-gray-900 py-16">
      <UContainer>
        <h2 class="text-2xl font-bold text-center mb-10">How it works</h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-8 text-center">
          <div>
            <p class="text-4xl mb-3">📅</p>
            <h3 class="font-semibold mb-1">1 · Choose your dates</h3>
            <p class="text-sm text-gray-500">Pick your pick-up and return dates and see the full price instantly.</p>
          </div>
          <div>
            <p class="text-4xl mb-3">🔍</p>
            <h3 class="font-semibold mb-1">2 · Review the details</h3>
            <p class="text-sm text-gray-500">Real photos, real condition, transparent total — no surprises at pick-up.</p>
          </div>
          <div>
            <p class="text-4xl mb-3">🚗</p>
            <h3 class="font-semibold mb-1">3 · Confirm and drive</h3>
            <p class="text-sm text-gray-500">The owner confirms your request and you're on the road.</p>
          </div>
        </div>
      </UContainer>
    </div>

    <!-- Featured vehicles -->
    <UContainer v-if="featured?.length" class="py-16">
      <div class="flex items-center justify-between mb-8">
        <h2 class="text-2xl font-bold">Available now</h2>
        <UButton variant="ghost" to="/vehicles" trailing-icon="i-lucide-arrow-right">See all</UButton>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <VehicleCard v-for="v in featured" :key="v.id" :vehicle="v" />
      </div>
    </UContainer>
  </div>
</template>
