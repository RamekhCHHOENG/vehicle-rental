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
    <!-- Hero — apple.com style: huge, tight, centered -->
    <section class="py-24 md:py-32 text-center">
      <UContainer>
        <h1 class="text-[44px] md:text-[64px] font-semibold leading-[1.05] tracking-[-0.04em]">
          Rent a car in<br class="md:hidden"> three easy steps.
        </h1>
        <p class="mt-5 text-[17px] md:text-[19px] text-neutral-500 dark:text-neutral-400 max-w-xl mx-auto leading-relaxed">
          Clear prices. Real photos. Every vehicle verified before it goes live.
        </p>
        <div class="mt-8 flex justify-center gap-3">
          <UButton size="lg" to="/vehicles">Browse vehicles</UButton>
          <UButton size="lg" variant="soft" color="neutral" to="/register">
            List your vehicle
          </UButton>
        </div>
      </UContainer>
    </section>

    <!-- How it works -->
    <section class="bg-neutral-50 dark:bg-neutral-900/60 py-20">
      <UContainer>
        <h2 class="text-[28px] font-semibold tracking-[-0.03em] text-center mb-12">How it works</h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div class="dk-card p-7 text-center">
            <p class="text-4xl mb-4">📅</p>
            <h3 class="text-[15px] font-semibold tracking-tight mb-1.5">1 · Choose your dates</h3>
            <p class="text-[13px] leading-relaxed text-neutral-500 dark:text-neutral-400">
              Pick your pick-up and return dates and see the full price instantly.
            </p>
          </div>
          <div class="dk-card p-7 text-center">
            <p class="text-4xl mb-4">🔍</p>
            <h3 class="text-[15px] font-semibold tracking-tight mb-1.5">2 · Review the details</h3>
            <p class="text-[13px] leading-relaxed text-neutral-500 dark:text-neutral-400">
              Real photos, real condition, transparent total — no surprises at pick-up.
            </p>
          </div>
          <div class="dk-card p-7 text-center">
            <p class="text-4xl mb-4">🚗</p>
            <h3 class="text-[15px] font-semibold tracking-tight mb-1.5">3 · Confirm and drive</h3>
            <p class="text-[13px] leading-relaxed text-neutral-500 dark:text-neutral-400">
              The owner confirms your request and you're on the road.
            </p>
          </div>
        </div>
      </UContainer>
    </section>

    <!-- Featured vehicles -->
    <section v-if="featured?.length" class="py-20">
      <UContainer>
        <div class="flex items-baseline justify-between mb-8">
          <h2 class="text-[28px] font-semibold tracking-[-0.03em]">Available now</h2>
          <UButton variant="ghost" size="sm" to="/vehicles" trailing-icon="i-lucide-arrow-right">
            See all
          </UButton>
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          <VehicleCard v-for="v in featured" :key="v.id" :vehicle="v" />
        </div>
      </UContainer>
    </section>
  </div>
</template>
