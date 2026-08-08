<script setup lang="ts">
import type { Vehicle, Booking } from '~/types'

const props = defineProps<{ vehicle: Vehicle }>()

const { user } = useAuth()
const api = useApi()
const toast = useToast()

// Three steps: dates → review → sent.
const step = ref(1)
const startDate = ref('')
const endDate = ref('')
const loading = ref(false)

const today = new Date().toISOString().slice(0, 10)

const days = computed(() => {
  if (!startDate.value || !endDate.value) return 0
  const ms = new Date(endDate.value).getTime() - new Date(startDate.value).getTime()
  return Math.round(ms / 86_400_000)
})

const totalPrice = computed(() => days.value * props.vehicle.price_per_day)
const datesValid = computed(() => startDate.value >= today && days.value >= 1)

async function confirmBooking() {
  loading.value = true
  try {
    await api<Booking>('/api/bookings', {
      method: 'POST',
      body: {
        vehicle_id: props.vehicle.id,
        start_date: startDate.value,
        end_date: endDate.value
      }
    })
    step.value = 3
  } catch (err: any) {
    toast.add({
      title: 'Booking not sent',
      description: err.data?.error ?? 'Something went wrong. Try again.',
      color: 'error'
    })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="dk-card p-6">
    <!-- Price and progress -->
    <div class="flex items-baseline justify-between pb-5 mb-5 border-b border-dashed border-[var(--ui-border)]">
      <p class="font-display numeric text-[26px] font-bold">
        ${{ vehicle.price_per_day }}<span class="text-[13px] font-medium text-[var(--ui-text-dimmed)]">/day</span>
      </p>
      <div class="flex gap-1.5" role="presentation">
        <span
          v-for="s in 3" :key="s"
          class="w-1.5 h-1.5 rounded-full transition-colors"
          :class="step >= s ? 'bg-saffron-400' : 'bg-[var(--ui-border)]'"
        />
      </div>
    </div>

    <div v-if="!user" class="text-center py-3 space-y-3">
      <p class="text-[13.5px] text-[var(--ui-text-muted)]">Log in to book this vehicle.</p>
      <UButton to="/login" block>Log in</UButton>
    </div>

    <div v-else-if="user.role !== 'renter'" class="text-center py-3">
      <p class="text-[13.5px] text-[var(--ui-text-muted)]">
        Bookings are made from a renter account.
      </p>
    </div>

    <!-- Step 1 -->
    <div v-else-if="step === 1" class="space-y-4">
      <p class="eyebrow">Step 1 of 3 · Dates</p>
      <UFormField label="Pick-up">
        <UInput v-model="startDate" type="date" :min="today" class="w-full" />
      </UFormField>
      <UFormField label="Return">
        <UInput v-model="endDate" type="date" :min="startDate || today" class="w-full" />
      </UFormField>
      <UButton block :disabled="!datesValid" @click="step = 2">See the price</UButton>
    </div>

    <!-- Step 2 -->
    <div v-else-if="step === 2" class="space-y-4">
      <p class="eyebrow">Step 2 of 3 · Review</p>

      <ul>
        <li class="record-row">
          <span class="text-[12.5px] text-[var(--ui-text-muted)]">Pick-up</span>
          <span class="text-[13px] numeric">{{ startDate }}</span>
        </li>
        <li class="record-row">
          <span class="text-[12.5px] text-[var(--ui-text-muted)]">Return</span>
          <span class="text-[13px] numeric">{{ endDate }}</span>
        </li>
        <li class="record-row">
          <span class="text-[12.5px] text-[var(--ui-text-muted)]">
            {{ days }} day{{ days > 1 ? 's' : '' }} × ${{ vehicle.price_per_day }}
          </span>
          <span class="text-[13px] numeric">${{ totalPrice.toFixed(2) }}</span>
        </li>
      </ul>

      <div class="flex items-baseline justify-between pt-4 border-t border-[var(--ui-border)]">
        <span class="display-md text-[15px]">Total</span>
        <span class="font-display numeric text-[24px] font-bold">${{ totalPrice.toFixed(2) }}</span>
      </div>

      <p class="text-[11.5px] text-[var(--ui-text-dimmed)] leading-snug">
        This is the full amount. No booking fee, no deposit taken online.
      </p>

      <div class="flex gap-2">
        <UButton variant="soft" color="neutral" class="flex-1" @click="step = 1">Back</UButton>
        <UButton class="flex-1" :loading="loading" @click="confirmBooking">Send request</UButton>
      </div>
    </div>

    <!-- Step 3 -->
    <div v-else class="py-3 space-y-3">
      <p class="eyebrow">Step 3 of 3 · Sent</p>
      <p class="display-md text-[19px]">Request sent to the owner</p>
      <p class="text-[13px] text-[var(--ui-text-muted)] leading-relaxed">
        They confirm the dates and you arrange the handover directly. You'll see
        the status under your bookings.
      </p>
      <UButton to="/dashboard/bookings" block variant="soft" color="neutral">
        View my bookings
      </UButton>
    </div>
  </div>
</template>
