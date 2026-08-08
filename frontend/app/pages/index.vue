<script setup lang="ts">
import type { Vehicle } from '~/types'

const api = useApi()

const { data: featured } = await useAsyncData(
  'featured-vehicles',
  () => api<Vehicle[]>('/api/vehicles'),
  { transform: (vehicles) => vehicles.slice(0, 3) }
)

// What a person actually checks before a listing goes live. Naming the checks
// is the promise; a slogan would only assert it.
const inspectionChecks = [
  'Photos match the actual vehicle',
  'Registration and insurance current',
  'Owner identity confirmed',
  'Price includes every fee'
]

const steps = [
  { n: '1', title: 'Pick your dates', body: 'Choose pick-up and return. The full price appears before you go further.' },
  { n: '2', title: 'Check the record', body: 'Real photos, real condition, the inspection date, and what the owner supplies.' },
  { n: '3', title: 'Send the request', body: 'The owner confirms and you arrange the handover. Nothing is charged in between.' }
]
</script>

<template>
  <div>
    <!-- Hero: the claim on the left, the evidence on the right. -->
    <section class="pt-16 pb-20 md:pt-24 md:pb-28">
      <UContainer>
        <div class="grid lg:grid-cols-[1.22fr_1fr] gap-12 lg:gap-14 items-center">
          <div>
            <p class="eyebrow mb-5">Inspected vehicle rental · Cambodia</p>

            <h1 class="display-xl text-[36px] sm:text-[52px] lg:text-[60px] xl:text-[68px]">
              The car you see<br class="hidden sm:block">
              is the car you get.
            </h1>

            <p class="mt-6 text-[16px] sm:text-[17px] leading-relaxed text-[var(--ui-text-muted)] max-w-lg">
              Someone from our team checks every vehicle before it appears here —
              the photos, the papers, and the price. Cars and motorbikes,
              in Phnom Penh, Siem Reap and beyond.
            </p>

            <div class="mt-9 flex flex-wrap gap-3">
              <UButton size="lg" to="/vehicles">Browse vehicles</UButton>
              <UButton size="lg" variant="soft" color="neutral" to="/register">
                List your vehicle
              </UButton>
            </div>
          </div>

          <!-- The signature: an inspection record, not a marketing panel. -->
          <div class="dk-card p-7 lg:p-8 relative">
            <div class="flex items-baseline justify-between mb-1">
              <p class="eyebrow">Inspection record</p>
              <p class="eyebrow numeric">Form 01</p>
            </div>

            <p class="display-md text-[21px] mt-3 mb-2">
              Checked before listing
            </p>

            <ul class="mt-4">
              <li v-for="check in inspectionChecks" :key="check" class="record-row">
                <span class="text-[13px] leading-snug pr-2">{{ check }}</span>
                <svg width="15" height="15" viewBox="0 0 16 16" fill="none" class="shrink-0 translate-y-0.5" aria-hidden="true">
                  <path
                    d="M3.5 8.5 L6.5 11.5 L12.5 4.5"
                    stroke="#f2a93b" stroke-width="2.4"
                    stroke-linecap="round" stroke-linejoin="round"
                  />
                </svg>
              </li>
            </ul>

            <div class="mt-7 pt-5 border-t border-dashed border-[var(--ui-border)] flex items-center justify-between gap-3">
              <p class="text-[12px] text-[var(--ui-text-dimmed)] leading-snug">
                Listings that fail a check<br>never reach the public site.
              </p>
              <VerifiedSeal large />
            </div>
          </div>
        </div>
      </UContainer>
    </section>

    <!-- Booking is genuinely a sequence, so it is numbered. -->
    <section class="py-20 border-y border-[var(--ui-border-muted)] bg-[var(--ui-bg-muted)]">
      <UContainer>
        <p class="eyebrow mb-3">How renting works</p>
        <h2 class="display-md text-[30px] sm:text-[36px] max-w-md mb-12">
          Three steps, and the price never moves.
        </h2>

        <div class="grid md:grid-cols-3 gap-x-10 gap-y-10">
          <div v-for="step in steps" :key="step.n">
            <div class="flex items-center gap-3 mb-3">
              <span
                class="font-display numeric text-[13px] font-bold w-7 h-7 rounded-[var(--r-small)] grid place-items-center bg-saffron-400 text-petrol-900"
              >{{ step.n }}</span>
              <span class="h-px flex-1 bg-[var(--ui-border)]" />
            </div>
            <h3 class="display-md text-[18px] mb-2">{{ step.title }}</h3>
            <p class="text-[13.5px] leading-relaxed text-[var(--ui-text-muted)]">
              {{ step.body }}
            </p>
          </div>
        </div>
      </UContainer>
    </section>

    <!-- Available now -->
    <section v-if="featured?.length" class="py-20">
      <UContainer>
        <div class="flex items-end justify-between mb-9 gap-4">
          <div>
            <p class="eyebrow mb-3">Available now</p>
            <h2 class="display-md text-[30px] sm:text-[36px]">Inspected and ready</h2>
          </div>
          <UButton variant="ghost" color="neutral" size="sm" to="/vehicles" trailing-icon="i-lucide-arrow-right">
            See all
          </UButton>
        </div>

        <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
          <VehicleCard v-for="v in featured" :key="v.id" :vehicle="v" />
        </div>
      </UContainer>
    </section>
  </div>
</template>
