<script setup lang="ts">
import type { Feature, Province, VehicleMake, VehicleModelRef } from '~/types'

// One section at a time, chosen by the admin page's single tab row. This
// component used to own a second row of tabs of its own, which left two tab
// strips stacked on one screen and no way to tell which owned what.
const props = defineProps<{ section: 'provinces' | 'makes' | 'features' }>()

const api = useApi()
const toast = useToast()

interface SmallLists {
  provinces: Province[]
  features: Feature[]
  usage: { provinces: Record<string, number> }
}

interface AdminMake extends VehicleMake {
  model_count: number
  listing_count: number
}

interface AdminModel extends VehicleModelRef {
  listing_count: number
}

const { data: lists, refresh: refreshLists, status: listsStatus } = await useAsyncData(
  'admin-metadata',
  () => api<SmallLists>('/api/admin/metadata')
)

/* ---------------- makes: searched and paged ---------------- */

const makeSearch = ref('')
const makeOffset = ref(0)
const PAGE = 50

// Debounced by hand rather than pulling in @vueuse: this is the only place that
// needs it, and typing must not fire a request per keystroke against a table of
// a few thousand rows.
const debouncedSearch = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null

watch(makeSearch, (value) => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    debouncedSearch.value = value
    makeOffset.value = 0
  }, 250)
})

const { data: makePage, refresh: refreshMakes, status: makesStatus } = await useAsyncData(
  'admin-makes',
  () => api<{ makes: AdminMake[], total: number }>('/api/admin/makes', {
    query: { q: debouncedSearch.value || undefined, limit: PAGE, offset: makeOffset.value }
  }),
  { watch: [debouncedSearch, makeOffset] }
)

const expandedMake = ref<string | null>(null)
const expandedModels = ref<AdminModel[]>([])
const modelsLoading = ref(false)

async function toggleMake(id: string) {
  if (expandedMake.value === id) {
    expandedMake.value = null
    return
  }
  expandedMake.value = id
  modelsLoading.value = true
  try {
    expandedModels.value = await api<AdminModel[]>(`/api/admin/makes/${id}/models`)
  } finally {
    modelsLoading.value = false
  }
}

/* ---------------- mutations ---------------- */

const busy = ref(false)

async function call(fn: () => Promise<unknown>, success: string) {
  busy.value = true
  try {
    await fn()
    await Promise.all([refreshLists(), refreshMakes()])
    if (expandedMake.value) {
      expandedModels.value = await api<AdminModel[]>(`/api/admin/makes/${expandedMake.value}/models`)
    }
    // The public vocabulary too, so open forms pick this up without a reload.
    await refreshNuxtData('metadata')
    toast.add({ title: success, color: 'success' })
  } catch (err: any) {
    toast.add({
      title: 'Could not save',
      description: err.data?.error ?? 'Something went wrong',
      color: 'error'
    })
  } finally {
    busy.value = false
  }
}

const setActive = (kind: string, id: string, active: boolean) => call(
  () => api(`/api/admin/${kind}/${id}`, { method: 'PUT', body: { active } }),
  active ? 'Restored' : 'Retired — hidden from the forms, listings untouched'
)

const remove = (kind: string, id: string) => call(
  () => api(`/api/admin/${kind}/${id}`, { method: 'DELETE' }),
  'Deleted'
)

/* ---------------- adding ---------------- */

const newProvince = reactive({ name_en: '', name_km: '' })
const newMake = ref('')
const newModel = reactive({ make_id: '', name: '', type: 'car' as 'car' | 'motorbike' })
const newFeature = reactive({ name: '', icon: '', applies_to: '' })

async function addProvince() {
  if (!newProvince.name_en) return
  await call(() => api('/api/admin/provinces', { method: 'POST', body: { ...newProvince } }), 'Province added')
  newProvince.name_en = ''
  newProvince.name_km = ''
}

async function addMake() {
  if (!newMake.value) return
  await call(() => api('/api/admin/makes', { method: 'POST', body: { name: newMake.value } }), 'Make added')
  newMake.value = ''
}

async function addModel() {
  if (!newModel.make_id || !newModel.name) return
  await call(() => api('/api/admin/models', { method: 'POST', body: { ...newModel } }), 'Model added')
  newModel.name = ''
}

async function addFeature() {
  if (!newFeature.name) return
  await call(() => api('/api/admin/features', { method: 'POST', body: { ...newFeature } }), 'Feature added')
  newFeature.name = ''
  newFeature.icon = ''
}

const makeSelectOptions = computed(() =>
  (makePage.value?.makes ?? []).map(m => ({ label: m.name, value: m.id }))
)
const typeOptions = [
  { label: 'Car', value: 'car' },
  { label: 'Motorbike', value: 'motorbike' }
]
const appliesOptions = [
  { label: 'Cars and motorbikes', value: '' },
  { label: 'Cars only', value: 'car' },
  { label: 'Motorbikes only', value: 'motorbike' }
]

/* ---------------- NHTSA import ---------------- */

interface ImportStatus {
  running: boolean
  started_at?: string
  finished_at?: string
  with_models: boolean
  result?: { makes_added: number, models_added: number, makes_seen: number, failures: number }
  error?: string
}

const importStatus = ref<ImportStatus | null>(null)
let importPoll: ReturnType<typeof setInterval> | null = null

async function readImportStatus() {
  try {
    importStatus.value = await api<ImportStatus>('/api/admin/metadata/import')
    if (!importStatus.value?.running) stopPolling()
  } catch {
    stopPolling()
  }
}

function stopPolling() {
  if (importPoll) {
    clearInterval(importPoll)
    importPoll = null
  }
  // One last refresh so the newly imported rows appear.
  refreshMakes()
}

async function startImport(withModels: boolean) {
  busy.value = true
  try {
    await api('/api/admin/metadata/import', { method: 'POST', body: { with_models: withModels } })
    toast.add({
      title: withModels ? 'Import started' : 'Importing makes',
      description: withModels
        ? 'Around 1,900 makes and their models. This takes a few minutes — you can leave this page.'
        : 'Makes only, this is quick.',
      color: 'success'
    })
    await readImportStatus()
    importPoll = importPoll ?? setInterval(readImportStatus, 3000)
  } catch (err: any) {
    toast.add({ title: 'Could not start', description: err.data?.error, color: 'error' })
  } finally {
    busy.value = false
  }
}

onMounted(() => {
  if (props.section === 'makes') readImportStatus()
})
onBeforeUnmount(() => {
  if (importPoll) clearInterval(importPoll)
  if (searchTimer) clearTimeout(searchTimer)
})
</script>

<template>
  <div>
    <!-- ---------------- provinces ---------------- -->
    <div v-if="section === 'provinces'">
      <p class="text-[13px] text-[var(--ui-text-muted)] mb-5 max-w-2xl">
        Where a vehicle can be listed. Anything an owner has already used can be
        retired but not deleted — retiring hides it from the forms and leaves
        those listings saying exactly what their owners said.
      </p>

      <div class="dk-card p-6">
        <div class="flex flex-wrap gap-3 mb-6">
          <UInput v-model="newProvince.name_en" placeholder="Province name" class="w-52" />
          <UInput v-model="newProvince.name_km" placeholder="ឈ្មោះខេត្ត (Khmer)" class="w-52" />
          <UButton :loading="busy" @click="addProvince">Add province</UButton>
        </div>

        <p v-if="listsStatus === 'pending'" class="py-8 text-center text-[13px] text-[var(--ui-text-dimmed)]">
          Loading…
        </p>
        <ul v-else>
          <li v-for="province in lists?.provinces ?? []" :key="province.id" class="record-row">
            <span class="flex items-center gap-2 min-w-0">
              <span
                class="text-[13px] font-medium truncate"
                :class="province.active ? '' : 'opacity-50 line-through'"
              >{{ province.name_en }}</span>
              <span class="text-[12px] text-[var(--ui-text-dimmed)] truncate">{{ province.name_km }}</span>
            </span>
            <span class="flex items-center gap-3 flex-none">
              <span class="text-[11.5px] text-[var(--ui-text-dimmed)] numeric">
                {{ lists?.usage?.provinces?.[province.id] ?? 0 }} listings
              </span>
              <UButton
                size="xs" variant="ghost" color="neutral" :loading="busy"
                @click="setActive('provinces', province.id, !province.active)"
              >{{ province.active ? 'Retire' : 'Restore' }}</UButton>
              <UButton
                v-if="!(lists?.usage?.provinces?.[province.id] ?? 0)"
                size="xs" variant="ghost" color="error" :loading="busy"
                @click="remove('provinces', province.id)"
              >Delete</UButton>
            </span>
          </li>
        </ul>
      </div>
    </div>

    <!-- ---------------- makes and models ---------------- -->
    <div v-else-if="section === 'makes'" class="space-y-4">
      <!-- Import first: typing two thousand manufacturers by hand is not a plan. -->
      <div class="dk-card p-6">
        <p class="eyebrow mb-2">Import from NHTSA</p>
        <p class="text-[13px] text-[var(--ui-text-muted)] mb-4 max-w-2xl">
          The US Department of Transportation publishes an open vehicle database:
          195 car makes, 1,684 motorcycle makes, and their models. Importing
          copies it into these tables, so listings keep working if that service
          does not. Names arrive shouted and the long tail is obscure — retire
          what does not belong here.
        </p>

        <div class="flex flex-wrap items-center gap-3">
          <UButton :loading="busy" :disabled="importStatus?.running" @click="startImport(true)">
            Import makes and models
          </UButton>
          <UButton
            variant="soft" color="neutral" :loading="busy" :disabled="importStatus?.running"
            @click="startImport(false)"
          >
            Makes only (fast)
          </UButton>

          <span v-if="importStatus?.running" class="flex items-center gap-2 text-[12.5px] text-[var(--ui-text-muted)]">
            <UIcon name="i-lucide-loader-circle" class="size-4 animate-spin" />
            Running — one request per make, this takes a few minutes.
          </span>
          <span
            v-else-if="importStatus?.result"
            class="text-[12.5px] numeric"
            :class="importStatus.error ? 'text-red-500' : 'text-[var(--ui-text-muted)]'"
          >
            Last run: {{ importStatus.result.makes_added }} makes and
            {{ importStatus.result.models_added }} models added
            from {{ importStatus.result.makes_seen }} seen<template v-if="importStatus.result.failures">
              , {{ importStatus.result.failures }} failed</template>.
            <template v-if="importStatus.error"> {{ importStatus.error }}</template>
          </span>
        </div>
      </div>

      <div class="dk-card p-6">
        <div class="flex flex-wrap gap-3 mb-4">
          <UInput v-model="newMake" placeholder="Make, e.g. BYD" class="w-52" />
          <UButton :loading="busy" @click="addMake">Add make</UButton>
        </div>
        <div class="flex flex-wrap gap-3">
          <USelectMenu
            v-model="newModel.make_id"
            :items="makeSelectOptions"
            value-key="value"
            :search-input="{ placeholder: 'Search makes…' }"
            placeholder="Make"
            class="w-44"
          />
          <UInput v-model="newModel.name" placeholder="Model, e.g. Atto 3" class="w-44" />
          <USelect v-model="newModel.type" :items="typeOptions" value-key="value" class="w-36" />
          <UButton :loading="busy" variant="soft" @click="addModel">Add model</UButton>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-3">
        <UInput
          v-model="makeSearch"
          placeholder="Search makes"
          icon="i-lucide-search"
          class="w-64"
          :loading="makesStatus === 'pending'"
        />
        <p class="text-[12px] text-[var(--ui-text-dimmed)] numeric">
          {{ makePage?.total ?? 0 }} makes
        </p>
        <div class="ml-auto flex items-center gap-2">
          <UButton
            size="xs" variant="ghost" color="neutral"
            :disabled="makeOffset === 0"
            @click="makeOffset = Math.max(0, makeOffset - PAGE)"
          >Previous</UButton>
          <UButton
            size="xs" variant="ghost" color="neutral"
            :disabled="makeOffset + PAGE >= (makePage?.total ?? 0)"
            @click="makeOffset += PAGE"
          >Next</UButton>
        </div>
      </div>

      <div v-for="make in makePage?.makes ?? []" :key="make.id" class="dk-card p-5">
        <div class="flex items-center justify-between gap-3">
          <button type="button" class="flex items-center gap-2 min-w-0 text-left" @click="toggleMake(make.id)">
            <UIcon
              name="i-lucide-chevron-right"
              class="size-4 flex-none transition-transform"
              :class="expandedMake === make.id ? 'rotate-90' : ''"
            />
            <span
              class="display-md text-[15px] truncate"
              :class="make.active ? '' : 'opacity-50 line-through'"
            >{{ make.name }}</span>
            <span class="text-[11.5px] text-[var(--ui-text-dimmed)] numeric flex-none">
              {{ make.model_count }} models · {{ make.listing_count }} listings
            </span>
          </button>

          <span class="flex items-center gap-2 flex-none">
            <UButton
              size="xs" variant="ghost" color="neutral" :loading="busy"
              @click="setActive('makes', make.id, !make.active)"
            >{{ make.active ? 'Retire' : 'Restore' }}</UButton>
            <UButton
              v-if="!make.listing_count"
              size="xs" variant="ghost" color="error" :loading="busy"
              @click="remove('makes', make.id)"
            >Delete</UButton>
          </span>
        </div>

        <div v-if="expandedMake === make.id" class="mt-3 pl-6">
          <p v-if="modelsLoading" class="py-3 text-[12.5px] text-[var(--ui-text-dimmed)]">Loading models…</p>
          <ul v-else>
            <li v-for="model in expandedModels" :key="model.id" class="record-row">
              <span class="flex items-center gap-2 min-w-0">
                <span
                  class="text-[13px] truncate"
                  :class="model.active ? '' : 'opacity-50 line-through'"
                >{{ model.name }}</span>
                <span class="dk-chip !text-[11px] !py-0.5 !px-2 capitalize">{{ model.type }}</span>
              </span>
              <span class="flex items-center gap-3 flex-none">
                <span class="text-[11.5px] text-[var(--ui-text-dimmed)] numeric">{{ model.listing_count }}</span>
                <UButton
                  size="xs" variant="ghost" color="neutral" :loading="busy"
                  @click="setActive('models', model.id, !model.active)"
                >{{ model.active ? 'Retire' : 'Restore' }}</UButton>
                <UButton
                  v-if="!model.listing_count"
                  size="xs" variant="ghost" color="error" :loading="busy"
                  @click="remove('models', model.id)"
                >Delete</UButton>
              </span>
            </li>
            <li v-if="!expandedModels.length" class="py-3 text-[12.5px] text-[var(--ui-text-dimmed)]">
              No models yet — this make cannot be listed against until it has one.
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- ---------------- features ---------------- -->
    <div v-else>
      <p class="text-[13px] text-[var(--ui-text-muted)] mb-5 max-w-2xl">
        What a listing can include. Renters filter on these, so they are a fixed
        vocabulary rather than a sentence in the description.
      </p>

      <div class="dk-card p-6">
        <div class="flex flex-wrap gap-3 mb-6">
          <UInput v-model="newFeature.name" placeholder="Feature, e.g. Roof rack" class="w-52" />
          <UInput v-model="newFeature.icon" placeholder="Icon, e.g. i-lucide-package" class="w-56" />
          <USelect v-model="newFeature.applies_to" :items="appliesOptions" value-key="value" class="w-52" />
          <UButton :loading="busy" @click="addFeature">Add feature</UButton>
        </div>

        <ul>
          <li v-for="feature in lists?.features ?? []" :key="feature.id" class="record-row">
            <span class="flex items-center gap-2 min-w-0">
              <UIcon v-if="feature.icon" :name="feature.icon" class="size-4 flex-none" />
              <span
                class="text-[13px] font-medium truncate"
                :class="feature.active ? '' : 'opacity-50 line-through'"
              >{{ feature.name }}</span>
              <span class="text-[11.5px] text-[var(--ui-text-dimmed)] truncate">
                {{ feature.applies_to ? `${feature.applies_to}s only` : 'both' }}
              </span>
            </span>
            <span class="flex items-center gap-3 flex-none">
              <UButton
                size="xs" variant="ghost" color="neutral" :loading="busy"
                @click="setActive('features', feature.id, !feature.active)"
              >{{ feature.active ? 'Retire' : 'Restore' }}</UButton>
              <UButton
                size="xs" variant="ghost" color="error" :loading="busy"
                @click="remove('features', feature.id)"
              >Delete</UButton>
            </span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
