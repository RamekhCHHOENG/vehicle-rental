<script setup lang="ts">

useSeoMeta({
  title: 'My vehicles',
  robots: 'noindex'
})
import type { Vehicle } from '~/types'

definePageMeta({ middleware: 'owner' })

const api = useApi()
const photoUrl = usePhotoUrl()
const toast = useToast()

const { data: vehicles, refresh } = await useAsyncData(
  'owner-vehicles',
  () => api<Vehicle[]>('/api/owner/vehicles')
)

async function removeVehicle(id: string) {
  if (!confirm('Delete this listing?')) return
  try {
    await api(`/api/owner/vehicles/${id}`, { method: 'DELETE' })
    toast.add({ title: 'Listing deleted' })
    refresh()
  } catch (err: any) {
    toast.add({ title: 'Delete failed', description: err.data?.error, color: 'error' })
  }
}
</script>

<template>
  <UContainer class="py-8">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-3xl font-bold">My vehicles</h1>
      <UButton to="/owner/vehicles/new" icon="i-lucide-plus">Add vehicle</UButton>
    </div>

    <div v-if="!vehicles?.length" class="text-center py-16 text-gray-500">
      <p class="mb-4">You haven't listed any vehicles yet.</p>
      <UButton to="/owner/vehicles/new">List your first vehicle</UButton>
    </div>

    <div v-else class="space-y-4">
      <UCard v-for="v in vehicles" :key="v.id">
        <div class="flex gap-4 items-center">
          <img :src="photoUrl(v.photos?.[0]?.file_path)" class="w-28 h-20 object-cover rounded" :alt="v.make">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <h3 class="font-semibold">{{ v.make }} {{ v.model }} {{ v.year }}</h3>
              <StatusBadge :status="v.status" />
            </div>
            <p class="text-sm text-gray-500">${{ v.price_per_day }}/day · {{ v.location }}</p>
            <p v-if="v.status === 'rejected' && v.rejection_reason" class="text-sm text-red-500 mt-1">
              Reason: {{ v.rejection_reason }}
            </p>
          </div>
          <div class="flex gap-2">
            <UButton :to="`/owner/vehicles/${v.id}`" variant="outline" color="neutral" size="sm">Edit</UButton>
            <UButton variant="outline" color="error" size="sm" @click="removeVehicle(v.id)">Delete</UButton>
          </div>
        </div>
      </UCard>
    </div>
  </UContainer>
</template>
