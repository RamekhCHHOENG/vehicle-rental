<script setup lang="ts">

useSeoMeta({
  title: 'Admin panel',
  robots: 'noindex'
})
import type { Vehicle, AdminUser } from '~/types'

definePageMeta({ middleware: 'admin' })

const api = useApi()
const photoUrl = usePhotoUrl()
const toast = useToast()

type Section = 'listings' | 'users' | 'provinces' | 'makes' | 'features'

const section = ref<Section>('listings')

// Flat, in one row. The three on the right configure the vocabulary listings
// are written in; the two on the left are the daily work.
const sectionTabs = [
  { label: 'Listings', value: 'listings' },
  { label: 'Users', value: 'users' },
  { label: 'Provinces', value: 'provinces' },
  { label: 'Makes & models', value: 'makes' },
  { label: 'Features', value: 'features' }
]

const metadataSections: Section[] = ['provinces', 'makes', 'features']

/* ---------------- stats ---------------- */

const { data: stats, refresh: refreshStats } = await useAsyncData(
  'admin-stats',
  () => api<Record<string, number>>('/api/admin/stats')
)

/* ---------------- listings ---------------- */

const statusFilter = ref<'pending' | 'approved' | 'rejected'>('pending')
const statusTabs = [
  { label: 'Pending', value: 'pending' },
  { label: 'Approved', value: 'approved' },
  { label: 'Rejected', value: 'rejected' }
]

const { data: vehicles, refresh: refreshVehicles } = await useAsyncData(
  'admin-vehicles',
  () => api<Vehicle[]>('/api/admin/vehicles', { query: { status: statusFilter.value } }),
  { watch: [statusFilter] }
)

// One modal serves both jobs: rejecting a new listing and taking down a live one.
const actioning = ref<Vehicle | null>(null)
const reason = ref('')
const isTakedown = computed(() => actioning.value?.status === 'approved')

async function approve(vehicle: Vehicle) {
  try {
    await api(`/api/admin/vehicles/${vehicle.id}/approve`, { method: 'POST' })
    toast.add({ title: `${vehicleName(vehicle)} approved`, color: 'success' })
    refreshAll()
  } catch (err: any) {
    toast.add({ title: 'Approve failed', description: err.data?.error, color: 'error' })
  }
}

async function submitRejection() {
  if (!actioning.value) return
  const wasLive = isTakedown.value
  try {
    await api(`/api/admin/vehicles/${actioning.value.id}/reject`, {
      method: 'POST',
      body: { reason: reason.value }
    })
    toast.add({ title: wasLive ? 'Listing taken down' : 'Listing rejected' })
    actioning.value = null
    reason.value = ''
    refreshAll()
  } catch (err: any) {
    toast.add({ title: 'Action failed', description: err.data?.error, color: 'error' })
  }
}

function refreshAll() {
  refreshVehicles()
  refreshStats()
}

/* ---------------- users ---------------- */

const userSearch = ref('')
const roleFilter = ref('all')
const roleOptions = [
  { label: 'All roles', value: 'all' },
  { label: 'Renters', value: 'renter' },
  { label: 'Owners', value: 'owner' },
  { label: 'Admins', value: 'admin' }
]

const { data: users, status: usersStatus } = await useAsyncData(
  'admin-users',
  () => api<AdminUser[]>('/api/admin/users', {
    query: {
      role: roleFilter.value === 'all' ? undefined : roleFilter.value,
      q: userSearch.value || undefined
    }
  }),
  { watch: [roleFilter, userSearch] }
)

const roleTone: Record<string, any> = {
  admin: 'error',
  owner: 'info',
  renter: 'neutral'
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-GB', {
    day: 'numeric', month: 'short', year: 'numeric'
  })
}
</script>

<template>
  <UContainer class="py-10">
    <p class="eyebrow mb-3">Verification &amp; oversight</p>
    <h1 class="display-md text-[34px] mb-8">Admin panel</h1>

    <!-- Key metrics -->
    <div v-if="stats" class="grid grid-cols-2 lg:grid-cols-6 gap-3 mb-10">
      <div class="dk-card p-4">
        <p class="text-2xl font-semibold tracking-tight">{{ stats.total_users }}</p>
        <p class="text-[11px] text-neutral-500 dark:text-neutral-400 mt-0.5">Users</p>
      </div>
      <div class="dk-card p-4">
        <p class="text-2xl font-semibold tracking-tight">{{ stats.total_vehicles }}</p>
        <p class="text-[11px] text-neutral-500 dark:text-neutral-400 mt-0.5">Vehicles</p>
      </div>
      <div class="dk-card p-4">
        <p class="text-2xl font-semibold tracking-tight text-amber-500">{{ stats.pending_vehicles }}</p>
        <p class="text-[11px] text-neutral-500 dark:text-neutral-400 mt-0.5">Pending</p>
      </div>
      <div class="dk-card p-4">
        <p class="text-2xl font-semibold tracking-tight text-green-600 dark:text-green-500">{{ stats.approved_vehicles }}</p>
        <p class="text-[11px] text-neutral-500 dark:text-neutral-400 mt-0.5">Approved</p>
      </div>
      <div class="dk-card p-4">
        <p class="text-2xl font-semibold tracking-tight">{{ stats.total_bookings }}</p>
        <p class="text-[11px] text-neutral-500 dark:text-neutral-400 mt-0.5">Bookings</p>
      </div>
      <div class="dk-card p-4">
        <p class="text-2xl font-semibold tracking-tight">{{ stats.completed_bookings }}</p>
        <p class="text-[11px] text-neutral-500 dark:text-neutral-400 mt-0.5">Completed</p>
      </div>
    </div>

    <UTabs v-model="section" :items="sectionTabs" :content="false" class="mb-6" />

    <!-- ============ LISTINGS ============ -->
    <AdminMetadata
      v-if="metadataSections.includes(section)"
      :section="(section as 'provinces' | 'makes' | 'features')"
    />

    <div v-else-if="section === 'listings'">
      <UTabs v-model="statusFilter" :items="statusTabs" :content="false" size="sm" class="mb-5" />

      <p v-if="!vehicles?.length" class="text-center py-16 text-[13px] text-neutral-400">
        No {{ statusFilter }} listings.
      </p>

      <div v-else class="space-y-3">
        <div v-for="v in vehicles" :key="v.id" class="dk-card p-4">
          <div class="flex gap-4">
            <img :src="photoUrl(v.photos?.[0]?.file_path)" class="w-32 h-24 object-cover rounded-[var(--r-control)] flex-none" :alt="vehicleName(v)">

            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 flex-wrap">
                <NuxtLink :to="`/vehicles/${v.id}`" class="text-[15px] font-semibold tracking-tight hover:underline">
                  {{ vehicleName(v) }} {{ v.year }}
                </NuxtLink>
                <StatusBadge :status="v.status" />
                <UBadge v-if="v.active_bookings" color="warning" variant="subtle" size="sm">
                  {{ v.active_bookings }} active booking{{ v.active_bookings > 1 ? 's' : '' }}
                </UBadge>
              </div>
              <p class="text-[12px] text-neutral-500 dark:text-neutral-400 capitalize mt-1">
                {{ v.type }} · {{ v.transmission }} · ${{ v.price_per_day }}/day · {{ provinceName(v) }}
              </p>
              <p class="text-[12px] text-neutral-500 dark:text-neutral-400">
                Owner: {{ v.owner?.full_name }} ({{ v.owner?.email }})
              </p>
              <p class="text-[12px] text-neutral-400 dark:text-neutral-500">
                {{ v.photos?.length ?? 0 }} photo(s)
              </p>
              <p v-if="v.status === 'rejected' && v.rejection_reason" class="text-[12px] text-red-500 mt-1">
                Reason: {{ v.rejection_reason }}
              </p>
            </div>

            <div class="flex flex-col gap-2 flex-none">
              <template v-if="v.status === 'pending'">
                <UButton color="success" size="sm" @click="approve(v)">Approve</UButton>
                <UButton color="error" variant="soft" size="sm" @click="actioning = v">Reject</UButton>
              </template>
              <UButton
                v-else-if="v.status === 'approved'"
                color="error" variant="soft" size="sm"
                @click="actioning = v"
              >
                Take down
              </UButton>
              <UButton
                v-else-if="v.status === 'rejected'"
                color="success" variant="soft" size="sm"
                @click="approve(v)"
              >
                Restore
              </UButton>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ============ USERS ============ -->
    <div v-else>
      <div class="flex flex-wrap gap-3 mb-5">
        <UInput
          v-model="userSearch"
          placeholder="Search name or email…"
          icon="i-lucide-search"
          class="w-64"
        />
        <USelect v-model="roleFilter" :items="roleOptions" value-key="value" class="w-40" />
        <p class="ml-auto self-center text-[12px] text-neutral-400">
          {{ users?.length ?? 0 }} user{{ users?.length === 1 ? '' : 's' }}
        </p>
      </div>

      <div class="dk-card overflow-hidden !p-0">
        <div class="overflow-x-auto">
          <table class="w-full text-[13px]">
            <thead>
              <tr class="text-left text-[10px] uppercase tracking-wider text-neutral-400 bg-neutral-50 dark:bg-neutral-900/60">
                <th class="font-semibold px-5 py-3">Name</th>
                <th class="font-semibold px-5 py-3">Contact</th>
                <th class="font-semibold px-5 py-3">Role</th>
                <th class="font-semibold px-5 py-3 text-right">Vehicles</th>
                <th class="font-semibold px-5 py-3 text-right">Bookings</th>
                <th class="font-semibold px-5 py-3">Joined</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="usersStatus === 'pending'">
                <td colspan="6" class="px-5 py-12 text-center text-neutral-400">Loading…</td>
              </tr>
              <tr v-else-if="!users?.length">
                <td colspan="6" class="px-5 py-12 text-center text-neutral-400">
                  No users match this search.
                </td>
              </tr>
              <tr
                v-for="u in users"
                :key="u.id"
                class="border-t border-neutral-900/8 dark:border-white/8 hover:bg-neutral-50 dark:hover:bg-neutral-900/40"
              >
                <td class="px-5 py-3.5 font-medium">{{ u.full_name }}</td>
                <td class="px-5 py-3.5 text-neutral-500 dark:text-neutral-400">
                  <p>{{ u.email }}</p>
                  <p v-if="u.phone" class="text-[11px] text-neutral-400">{{ u.phone }}</p>
                </td>
                <td class="px-5 py-3.5">
                  <UBadge :color="roleTone[u.role]" variant="subtle" size="sm" class="capitalize">
                    {{ u.role }}
                  </UBadge>
                </td>
                <td class="px-5 py-3.5 text-right tabular-nums" :class="u.vehicle_count ? '' : 'text-neutral-300 dark:text-neutral-600'">
                  {{ u.vehicle_count }}
                </td>
                <td class="px-5 py-3.5 text-right tabular-nums" :class="u.booking_count ? '' : 'text-neutral-300 dark:text-neutral-600'">
                  {{ u.booking_count }}
                </td>
                <td class="px-5 py-3.5 text-neutral-500 dark:text-neutral-400 whitespace-nowrap">
                  {{ formatDate(u.created_at) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Reject / take-down dialog -->
    <UModal
      :open="!!actioning"
      @update:open="actioning = null"
      :title="isTakedown ? 'Take down listing' : 'Reject listing'"
    >
      <template #body>
        <div class="space-y-4">
          <UAlert
            v-if="isTakedown && actioning?.active_bookings"
            color="warning"
            variant="subtle"
            :title="`${actioning.active_bookings} active booking(s) on this vehicle`"
            description="Taking the listing down hides it from search. Existing bookings stay in place — contact the renters if the vehicle is unavailable."
          />

          <p class="text-[13px] text-neutral-500 dark:text-neutral-400">
            {{ isTakedown
              ? 'This removes the listing from public search. The owner sees your reason and can fix and resubmit.'
              : 'Tell the owner what to fix — they see this on their dashboard.' }}
          </p>

          <UTextarea
            v-model="reason"
            :rows="3"
            :placeholder="isTakedown ? 'e.g. Vehicle reported unsafe by a renter' : 'e.g. Photos are missing or unclear'"
            class="w-full"
          />

          <div class="flex justify-end gap-2">
            <UButton variant="soft" color="neutral" @click="actioning = null">Cancel</UButton>
            <UButton color="error" :disabled="!reason.trim()" @click="submitRejection">
              {{ isTakedown ? 'Take down' : 'Reject listing' }}
            </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </UContainer>
</template>
