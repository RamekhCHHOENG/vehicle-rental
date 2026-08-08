<script setup lang="ts">
import type { Vehicle } from '~/types'

const props = defineProps<{ vehicle?: Vehicle }>()
const emit = defineEmits<{ saved: [vehicle: Vehicle] }>()

const api = useApi()
const toast = useToast()

const state = reactive({
  type: props.vehicle?.type ?? 'car',
  make: props.vehicle?.make ?? '',
  model: props.vehicle?.model ?? '',
  year: props.vehicle?.year ?? new Date().getFullYear(),
  transmission: props.vehicle?.transmission ?? 'auto',
  seats: props.vehicle?.seats ?? 5,
  price_per_day: props.vehicle?.price_per_day ?? 30,
  location: props.vehicle?.location ?? '',
  description: props.vehicle?.description ?? ''
})

const loading = ref(false)

const typeOptions = [
  { label: 'Car', value: 'car' },
  { label: 'Motorbike', value: 'motorbike' }
]
const transmissionOptions = [
  { label: 'Automatic', value: 'auto' },
  { label: 'Manual', value: 'manual' }
]

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
  <form class="space-y-4" @submit.prevent="onSubmit">
    <div class="grid grid-cols-2 gap-4">
      <UFormField label="Type" required>
        <USelect v-model="state.type" :items="typeOptions" value-key="value" class="w-full" />
      </UFormField>
      <UFormField label="Transmission" required>
        <USelect v-model="state.transmission" :items="transmissionOptions" value-key="value" class="w-full" />
      </UFormField>
      <UFormField label="Make" required>
        <UInput v-model="state.make" placeholder="Toyota" class="w-full" />
      </UFormField>
      <UFormField label="Model" required>
        <UInput v-model="state.model" placeholder="Camry" class="w-full" />
      </UFormField>
      <UFormField label="Year" required>
        <UInput v-model.number="state.year" type="number" class="w-full" />
      </UFormField>
      <UFormField label="Seats">
        <UInput v-model.number="state.seats" type="number" class="w-full" />
      </UFormField>
      <UFormField label="Price per day (USD)" required>
        <UInput v-model.number="state.price_per_day" type="number" step="0.5" class="w-full" />
      </UFormField>
      <UFormField label="Location" required>
        <UInput v-model="state.location" placeholder="Phnom Penh" class="w-full" />
      </UFormField>
    </div>
    <UFormField label="Description" hint="Real condition, features, pick-up details">
      <UTextarea v-model="state.description" :rows="4" class="w-full" />
    </UFormField>
    <UButton type="submit" block :loading="loading">
      {{ props.vehicle ? 'Save changes' : 'Create listing' }}
    </UButton>
  </form>
</template>
