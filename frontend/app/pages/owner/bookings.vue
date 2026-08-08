<script setup lang="ts">
import type { Booking } from '~/types'

definePageMeta({ middleware: 'owner' })

const api = useApi()
const toast = useToast()

const { data: bookings, refresh } = await useAsyncData(
  'owner-bookings',
  () => api<Booking[]>('/api/owner/bookings')
)

async function act(b: Booking, action: 'confirm' | 'reject' | 'complete') {
  try {
    await api(`/api/owner/bookings/${b.id}/${action}`, { method: 'POST' })
    toast.add({ title: `Booking ${action}ed`.replace('confirmed', 'confirmed ✓'), color: action === 'reject' ? 'neutral' : 'success' })
    refresh()
  } catch (err: any) {
    toast.add({ title: 'Action failed', description: err.data?.error, color: 'error' })
  }
}
</script>

<template>
  <UContainer class="py-8">
    <h1 class="text-3xl font-bold mb-6">Booking requests</h1>

    <div v-if="!bookings?.length" class="text-center py-16 text-gray-500">
      No booking requests yet. They'll appear here once renters book your vehicles.
    </div>

    <div v-else class="space-y-4">
      <UCard v-for="b in bookings" :key="b.id">
        <div class="flex gap-4 items-center flex-wrap">
          <div class="flex-1 min-w-48">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="font-semibold">{{ b.vehicle?.make }} {{ b.vehicle?.model }}</span>
              <StatusBadge :status="b.status" />
            </div>
            <p class="text-sm text-gray-500">
              {{ b.start_date.slice(0, 10) }} → {{ b.end_date.slice(0, 10) }}
              · <span class="font-medium text-primary">${{ b.total_price.toFixed(2) }}</span>
            </p>
            <p class="text-sm text-gray-500 mt-1">
              Renter: {{ b.renter?.full_name }}
              <template v-if="b.renter?.phone"> · 📞 {{ b.renter.phone }}</template>
            </p>
          </div>
          <div class="flex gap-2">
            <template v-if="b.status === 'requested'">
              <UButton color="success" size="sm" @click="act(b, 'confirm')">Confirm</UButton>
              <UButton color="error" variant="outline" size="sm" @click="act(b, 'reject')">Reject</UButton>
            </template>
            <UButton
              v-if="b.status === 'confirmed'"
              size="sm" variant="outline"
              @click="act(b, 'complete')"
            >
              Mark returned
            </UButton>
          </div>
        </div>
      </UCard>
    </div>
  </UContainer>
</template>
