<script setup lang="ts">
import type { Booking } from '~/types'

definePageMeta({ middleware: 'auth' })

const api = useApi()
const photoUrl = usePhotoUrl()
const toast = useToast()

const { data: bookings, refresh } = await useAsyncData(
  'my-bookings',
  () => api<Booking[]>('/api/bookings')
)

const reviewing = ref<Booking | null>(null)
const rating = ref(5)
const comment = ref('')

async function cancelBooking(b: Booking) {
  if (!confirm('Cancel this booking?')) return
  try {
    await api(`/api/bookings/${b.id}/cancel`, { method: 'POST' })
    toast.add({ title: 'Booking cancelled' })
    refresh()
  } catch (err: any) {
    toast.add({ title: 'Cancel failed', description: err.data?.error, color: 'error' })
  }
}

async function submitReview() {
  if (!reviewing.value) return
  try {
    await api(`/api/bookings/${reviewing.value.id}/review`, {
      method: 'POST',
      body: { rating: rating.value, comment: comment.value }
    })
    toast.add({ title: 'Thanks for your review!' })
    reviewing.value = null
    rating.value = 5
    comment.value = ''
    refresh()
  } catch (err: any) {
    toast.add({ title: 'Review failed', description: err.data?.error, color: 'error' })
  }
}
</script>

<template>
  <UContainer class="py-8">
    <h1 class="text-3xl font-bold mb-6">My bookings</h1>

    <div v-if="!bookings?.length" class="text-center py-16 text-gray-500">
      <p class="mb-4">No bookings yet.</p>
      <UButton to="/vehicles">Browse vehicles</UButton>
    </div>

    <div v-else class="space-y-4">
      <UCard v-for="b in bookings" :key="b.id">
        <div class="flex gap-4 items-center flex-wrap">
          <img :src="photoUrl(b.vehicle?.photos?.[0]?.file_path)" class="w-28 h-20 object-cover rounded">
          <div class="flex-1 min-w-48">
            <div class="flex items-center gap-2 flex-wrap">
              <NuxtLink :to="`/vehicles/${b.vehicle_id}`" class="font-semibold hover:underline">
                {{ b.vehicle?.make }} {{ b.vehicle?.model }}
              </NuxtLink>
              <StatusBadge :status="b.status" />
            </div>
            <p class="text-sm text-gray-500">
              {{ b.start_date.slice(0, 10) }} → {{ b.end_date.slice(0, 10) }}
            </p>
            <p class="text-sm font-medium text-primary">${{ b.total_price.toFixed(2) }} total</p>
            <p v-if="b.review" class="text-sm text-amber-500 mt-1">
              Your review: {{ '★'.repeat(b.review.rating) }}
            </p>
          </div>
          <div class="flex gap-2">
            <UButton
              v-if="b.status === 'requested' || b.status === 'confirmed'"
              variant="outline" color="error" size="sm"
              @click="cancelBooking(b)"
            >
              Cancel
            </UButton>
            <UButton
              v-if="b.status === 'completed' && !b.review"
              size="sm"
              @click="reviewing = b"
            >
              Leave review
            </UButton>
          </div>
        </div>
      </UCard>
    </div>

    <!-- Review dialog -->
    <UModal :open="!!reviewing" @update:open="reviewing = null" title="Rate your rental">
      <template #body>
        <div class="space-y-4">
          <p class="text-sm text-gray-500">
            How was the {{ reviewing?.vehicle?.make }} {{ reviewing?.vehicle?.model }}?
          </p>
          <div class="flex gap-1 text-3xl">
            <button
              v-for="star in 5" :key="star"
              class="transition-transform hover:scale-110"
              :class="star <= rating ? 'text-amber-400' : 'text-gray-300 dark:text-gray-600'"
              @click="rating = star"
            >★</button>
          </div>
          <UTextarea v-model="comment" :rows="3" placeholder="Share details for future renters…" class="w-full" />
          <div class="flex justify-end gap-2">
            <UButton variant="outline" color="neutral" @click="reviewing = null">Cancel</UButton>
            <UButton @click="submitReview">Submit review</UButton>
          </div>
        </div>
      </template>
    </UModal>
  </UContainer>
</template>
