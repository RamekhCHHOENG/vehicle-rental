<script setup lang="ts">
import type { Vehicle, Booking } from '~/types'

const props = defineProps<{ vehicle: Vehicle }>()

const { user } = useAuth()
const api = useApi()
const toast = useToast()

// 3-step flow: 1 dates → 2 review → 3 done
const step = ref(1)
const startDate = ref('')
const endDate = ref('')
const loading = ref(false)
const booking = ref<Booking | null>(null)

const today = new Date().toISOString().slice(0, 10)

const days = computed(() => {
  if (!startDate.value || !endDate.value) return 0
  const ms = new Date(endDate.value).getTime() - new Date(startDate.value).getTime()
  return Math.round(ms / 86_400_000)
})

const totalPrice = computed(() => days.value * props.vehicle.price_per_day)

const datesValid = computed(() =>
  startDate.value >= today && days.value >= 1
)

async function confirmBooking() {
  loading.value = true
  try {
    booking.value = await api<Booking>('/api/bookings', {
      method: 'POST',
      body: {
        vehicle_id: props.vehicle.id,
        start_date: startDate.value,
        end_date: endDate.value
      }
    })
    step.value = 3
  } catch (err: any) {
    toast.add({ title: 'Booking failed', description: err.data?.error ?? 'Something went wrong', color: 'error' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <span class="text-2xl font-bold text-primary">${{ vehicle.price_per_day }}<span class="text-sm font-normal text-gray-500">/day</span></span>
        <div class="flex gap-1">
          <span v-for="s in 3" :key="s" class="w-2 h-2 rounded-full" :class="step >= s ? 'bg-primary' : 'bg-gray-300 dark:bg-gray-700'" />
        </div>
      </div>
    </template>

    <!-- Not logged in -->
    <div v-if="!user" class="text-center py-4 space-y-3">
      <p class="text-gray-500">Log in to book this vehicle.</p>
      <UButton to="/login" block>Log in</UButton>
    </div>

    <!-- Owners/admins don't book -->
    <div v-else-if="user.role !== 'renter'" class="text-center py-4">
      <p class="text-gray-500">Only renter accounts can make bookings.</p>
    </div>

    <!-- Step 1: pick dates -->
    <div v-else-if="step === 1" class="space-y-4">
      <p class="font-medium">Step 1 · Choose your dates</p>
      <UFormField label="Pick-up date">
        <UInput v-model="startDate" type="date" :min="today" class="w-full" />
      </UFormField>
      <UFormField label="Return date">
        <UInput v-model="endDate" type="date" :min="startDate || today" class="w-full" />
      </UFormField>
      <UButton block :disabled="!datesValid" @click="step = 2">Continue</UButton>
    </div>

    <!-- Step 2: review -->
    <div v-else-if="step === 2" class="space-y-4">
      <p class="font-medium">Step 2 · Review your booking</p>
      <ul class="text-sm space-y-2">
        <li class="flex justify-between"><span class="text-gray-500">Vehicle</span><span>{{ vehicle.make }} {{ vehicle.model }}</span></li>
        <li class="flex justify-between"><span class="text-gray-500">Pick-up</span><span>{{ startDate }}</span></li>
        <li class="flex justify-between"><span class="text-gray-500">Return</span><span>{{ endDate }}</span></li>
        <li class="flex justify-between"><span class="text-gray-500">Duration</span><span>{{ days }} day{{ days > 1 ? 's' : '' }}</span></li>
        <li class="flex justify-between font-bold text-base border-t pt-2 border-gray-200 dark:border-gray-700">
          <span>Total</span><span class="text-primary">${{ totalPrice.toFixed(2) }}</span>
        </li>
      </ul>
      <p class="text-xs text-gray-500">No hidden fees — this is the full price.</p>
      <div class="flex gap-2">
        <UButton variant="outline" color="neutral" class="flex-1" @click="step = 1">Back</UButton>
        <UButton class="flex-1" :loading="loading" @click="confirmBooking">Step 3 · Confirm</UButton>
      </div>
    </div>

    <!-- Step 3: done -->
    <div v-else class="text-center py-4 space-y-3">
      <p class="text-4xl">🎉</p>
      <p class="font-medium">Booking requested!</p>
      <p class="text-sm text-gray-500">The owner will confirm your request soon.</p>
      <UButton to="/dashboard/bookings" block variant="outline">View my bookings</UButton>
    </div>
  </UCard>
</template>
