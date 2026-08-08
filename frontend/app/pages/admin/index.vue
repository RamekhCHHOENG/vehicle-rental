<script setup lang="ts">
import type { Vehicle } from '~/types'

definePageMeta({ middleware: 'admin' })

const api = useApi()
const photoUrl = usePhotoUrl()
const toast = useToast()

const statusFilter = ref<'pending' | 'approved' | 'rejected'>('pending')
const statusTabs = [
  { label: 'Pending', value: 'pending' },
  { label: 'Approved', value: 'approved' },
  { label: 'Rejected', value: 'rejected' }
]

const { data: stats, refresh: refreshStats } = await useAsyncData(
  'admin-stats',
  () => api<Record<string, number>>('/api/admin/stats')
)

const { data: vehicles, refresh: refreshVehicles } = await useAsyncData(
  'admin-vehicles',
  () => api<Vehicle[]>('/api/admin/vehicles', { query: { status: statusFilter.value } }),
  { watch: [statusFilter] }
)

const rejectingVehicle = ref<Vehicle | null>(null)
const rejectionReason = ref('')

async function approve(vehicle: Vehicle) {
  try {
    await api(`/api/admin/vehicles/${vehicle.id}/approve`, { method: 'POST' })
    toast.add({ title: `${vehicle.make} ${vehicle.model} approved`, color: 'success' })
    refreshAll()
  } catch (err: any) {
    toast.add({ title: 'Approve failed', description: err.data?.error, color: 'error' })
  }
}

async function reject() {
  if (!rejectingVehicle.value) return
  try {
    await api(`/api/admin/vehicles/${rejectingVehicle.value.id}/reject`, {
      method: 'POST',
      body: { reason: rejectionReason.value }
    })
    toast.add({ title: 'Listing rejected' })
    rejectingVehicle.value = null
    rejectionReason.value = ''
    refreshAll()
  } catch (err: any) {
    toast.add({ title: 'Reject failed', description: err.data?.error, color: 'error' })
  }
}

function refreshAll() {
  refreshVehicles()
  refreshStats()
}
</script>

<template>
  <UContainer class="py-8">
    <h1 class="text-3xl font-bold mb-6">Admin panel</h1>

    <!-- Key metrics -->
    <div v-if="stats" class="grid grid-cols-2 lg:grid-cols-6 gap-4 mb-8">
      <UCard><p class="text-2xl font-bold">{{ stats.total_users }}</p><p class="text-sm text-gray-500">Users</p></UCard>
      <UCard><p class="text-2xl font-bold">{{ stats.total_vehicles }}</p><p class="text-sm text-gray-500">Vehicles</p></UCard>
      <UCard><p class="text-2xl font-bold text-amber-500">{{ stats.pending_vehicles }}</p><p class="text-sm text-gray-500">Pending</p></UCard>
      <UCard><p class="text-2xl font-bold text-green-500">{{ stats.approved_vehicles }}</p><p class="text-sm text-gray-500">Approved</p></UCard>
      <UCard><p class="text-2xl font-bold">{{ stats.total_bookings }}</p><p class="text-sm text-gray-500">Bookings</p></UCard>
      <UCard><p class="text-2xl font-bold">{{ stats.completed_bookings }}</p><p class="text-sm text-gray-500">Completed</p></UCard>
    </div>

    <!-- Verification queue -->
    <div class="flex items-center gap-4 mb-4">
      <h2 class="text-xl font-semibold">Listings</h2>
      <UTabs v-model="statusFilter" :items="statusTabs" :content="false" size="sm" />
    </div>

    <div v-if="!vehicles?.length" class="text-center py-12 text-gray-500">
      No {{ statusFilter }} listings.
    </div>

    <div v-else class="space-y-4">
      <UCard v-for="v in vehicles" :key="v.id">
        <div class="flex gap-4">
          <img :src="photoUrl(v.photos?.[0]?.file_path)" class="w-32 h-24 object-cover rounded" :alt="v.make">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <NuxtLink :to="`/vehicles/${v.id}`" class="font-semibold hover:underline">
                {{ v.make }} {{ v.model }} {{ v.year }}
              </NuxtLink>
              <StatusBadge :status="v.status" />
            </div>
            <p class="text-sm text-gray-500 capitalize">
              {{ v.type }} · {{ v.transmission }} · ${{ v.price_per_day }}/day · {{ v.location }}
            </p>
            <p class="text-sm text-gray-500 mt-1">
              Owner: {{ v.owner?.full_name }} ({{ v.owner?.email }})
            </p>
            <p class="text-sm text-gray-500">{{ v.photos?.length ?? 0 }} photo(s)</p>
          </div>
          <div v-if="v.status === 'pending'" class="flex flex-col gap-2">
            <UButton color="success" size="sm" @click="approve(v)">Approve</UButton>
            <UButton color="error" variant="outline" size="sm" @click="rejectingVehicle = v">Reject</UButton>
          </div>
        </div>
      </UCard>
    </div>

    <!-- Reject reason dialog -->
    <UModal :open="!!rejectingVehicle" @update:open="rejectingVehicle = null" title="Reject listing">
      <template #body>
        <div class="space-y-4">
          <p class="text-sm text-gray-500">
            Tell {{ rejectingVehicle?.owner?.full_name ?? 'the owner' }} what to fix —
            they'll see this reason on their dashboard.
          </p>
          <UTextarea v-model="rejectionReason" :rows="3" placeholder="e.g. Photos are missing or unclear" class="w-full" />
          <div class="flex justify-end gap-2">
            <UButton variant="outline" color="neutral" @click="rejectingVehicle = null">Cancel</UButton>
            <UButton color="error" :disabled="!rejectionReason.trim()" @click="reject">Reject listing</UButton>
          </div>
        </div>
      </template>
    </UModal>
  </UContainer>
</template>
