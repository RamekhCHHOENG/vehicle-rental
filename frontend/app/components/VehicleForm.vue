<script setup lang="ts">
import type { Vehicle } from '~/types'

const props = defineProps<{ vehicle?: Vehicle }>()
const emit = defineEmits<{ saved: [vehicle: Vehicle] }>()

const api = useApi()
const toast = useToast()
const { data: metadata, status: metadataStatus } = await useMetadata()

const state = reactive({
  type: props.vehicle?.type ?? 'car',
  make_id: props.vehicle?.make_id ?? '',
  model_id: props.vehicle?.model_id ?? '',
  province_id: props.vehicle?.province_id ?? '',
  year: props.vehicle?.year ?? new Date().getFullYear(),
  transmission: props.vehicle?.transmission ?? 'auto',
  seats: props.vehicle?.seats ?? 5,
  price_per_day: props.vehicle?.price_per_day ?? 30,
  description: props.vehicle?.description ?? '',
  feature_ids: props.vehicle?.features?.map(f => f.id) ?? []
})

const loading = ref(false)

// A make is offered only if it actually sells this kind of vehicle: picking
// "motorbike" should not present Lexus. The flags come from the server, since
// the models themselves are fetched separately.
const makeOptions = computed(() =>
  (metadata.value?.makes ?? [])
    .filter(m => (state.type === 'car' ? m.has_cars : m.has_motorbikes))
    .map(m => ({ label: m.name, value: m.id }))
)

// Models arrive per make. The catalogue is thousands of makes deep, so they
// cannot travel with the rest of the metadata.
const selectedMake = computed(() => state.make_id || undefined)
const selectedType = computed(() => state.type)
const { data: models, status: modelsStatus } = await useModels(selectedMake, selectedType)

const modelOptions = computed(() =>
  (models.value ?? []).map(model => ({ label: model.name, value: model.id }))
)

const provinceOptions = computed(() =>
  (metadata.value?.provinces ?? []).map(p => ({
    label: p.name_km ? `${p.name_en} · ${p.name_km}` : p.name_en,
    value: p.id
  }))
)

const seatOptions = computed(() =>
  (metadata.value?.seat_options ?? []).map(n => ({ label: `${n}`, value: n }))
)

// Features are shown when they suit the chosen type; an empty applies_to means
// both. Anything already ticked that stops applying is dropped, so a car
// converted to a motorbike cannot keep its child seat.
const featureOptions = computed(() =>
  (metadata.value?.features ?? []).filter(f => !f.applies_to || f.applies_to === state.type)
)

// Changing type or make can strip the model out from under the selection. The
// server rejects a mismatch anyway; clearing it here means the owner sees the
// empty select rather than a validation error after saving.
watch(() => state.type, () => {
  if (!makeOptions.value.some(o => o.value === state.make_id)) state.make_id = ''
  state.model_id = ''
  const stillValid = new Set(featureOptions.value.map(f => f.id))
  state.feature_ids = state.feature_ids.filter(id => stillValid.has(id))
  if (state.type === 'motorbike') state.seats = 2
})

watch(() => state.make_id, (id, previous) => {
  if (previous !== undefined && id !== previous) state.model_id = ''
})

function toggleFeature(id: string) {
  const i = state.feature_ids.indexOf(id)
  if (i === -1) state.feature_ids.push(id)
  else state.feature_ids.splice(i, 1)
}

async function onSubmit() {
  loading.value = true
  try {
    const saved = props.vehicle
      ? await api<Vehicle>(`/api/owner/vehicles/${props.vehicle.id}`, { method: 'PUT', body: state })
      : await api<Vehicle>('/api/owner/vehicles', { method: 'POST', body: state })
    emit('saved', saved)
  } catch (err: any) {
    toast.add({ title: 'Save failed', description: err.data?.error ?? 'Something went wrong', color: 'error' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <p v-if="metadataStatus === 'pending'" class="py-10 text-center text-[13px] text-[var(--ui-text-dimmed)]">
    Loading…
  </p>

  <form v-else class="space-y-4" @submit.prevent="onSubmit">
    <div class="grid grid-cols-2 gap-4">
      <UFormField label="Type" required>
        <USelect v-model="state.type" :items="metadata?.vehicle_types ?? []" value-key="value" class="w-full" />
      </UFormField>
      <UFormField label="Transmission" required>
        <USelect v-model="state.transmission" :items="metadata?.transmissions ?? []" value-key="value" class="w-full" />
      </UFormField>

      <UFormField label="Make" required>
        <USelectMenu
          v-model="state.make_id"
          :items="makeOptions"
          value-key="value"
          :search-input="{ placeholder: 'Search makes…' }"
          placeholder="Choose a make"
          class="w-full"
        />
      </UFormField>
      <UFormField
        label="Model"
        required
        :hint="state.make_id ? undefined : 'Choose a make first'"
      >
        <USelectMenu
          v-model="state.model_id"
          :items="modelOptions"
          value-key="value"
          :disabled="!state.make_id"
          :loading="modelsStatus === 'pending'"
          :search-input="{ placeholder: 'Search models…' }"
          placeholder="Choose a model"
          class="w-full"
        />
      </UFormField>

      <UFormField label="Year" required>
        <UInput v-model.number="state.year" type="number" class="w-full" />
      </UFormField>
      <UFormField label="Seats" required>
        <USelect v-model="state.seats" :items="seatOptions" value-key="value" class="w-full" />
      </UFormField>

      <UFormField label="Price per day (USD)" required>
        <UInput v-model.number="state.price_per_day" type="number" step="0.5" class="w-full" />
      </UFormField>
      <UFormField label="Province" required>
        <USelectMenu
          v-model="state.province_id"
          :items="provinceOptions"
          value-key="value"
          :search-input="{ placeholder: 'Search provinces…' }"
          placeholder="Choose a province"
          class="w-full"
        />
      </UFormField>
    </div>

    <UFormField label="What's included" hint="Renters filter on these">
      <div class="flex flex-wrap gap-2 pt-1">
        <button
          v-for="feature in featureOptions"
          :key="feature.id"
          type="button"
          class="dk-chip"
          :class="state.feature_ids.includes(feature.id) ? 'dk-chip-on' : ''"
          @click="toggleFeature(feature.id)"
        >
          <UIcon v-if="feature.icon" :name="feature.icon" class="size-3.5" />
          {{ feature.name }}
        </button>
      </div>
    </UFormField>

    <UFormField label="Description" hint="Real condition, features, pick-up details">
      <UTextarea v-model="state.description" :rows="4" class="w-full" />
    </UFormField>

    <UButton type="submit" block :loading="loading">
      {{ props.vehicle ? 'Save changes' : 'Create listing' }}
    </UButton>
  </form>
</template>
