<script setup lang="ts">

useSeoMeta({
  title: 'Edit listing',
  robots: 'noindex'
})
import type { Vehicle, VehiclePhoto } from '~/types'

definePageMeta({ middleware: 'owner' })

const route = useRoute()
const api = useApi()
const config = useRuntimeConfig()
const photoUrl = usePhotoUrl()
const toast = useToast()

const { data, refresh } = await useAsyncData(
  `owner-vehicle-${route.params.id}`,
  () => api<{ vehicle: Vehicle }>(`/api/vehicles/${route.params.id}`)
)
const vehicle = computed(() => data.value?.vehicle)

const uploading = ref(false)
const fileInput = ref<HTMLInputElement>()

async function uploadPhotos(event: Event) {
  const files = (event.target as HTMLInputElement).files
  if (!files?.length) return

  uploading.value = true
  try {
    for (const file of files) {
      const form = new FormData()
      form.append('photo', file)
      await $fetch(`/api/owner/vehicles/${route.params.id}/photos`, {
        baseURL: config.public.apiBase,
        method: 'POST',
        credentials: 'include',
        body: form
      })
    }
    toast.add({ title: `${files.length} photo(s) uploaded` })
    refresh()
  } catch (err: any) {
    toast.add({ title: 'Upload failed', description: err.data?.error, color: 'error' })
  } finally {
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function deletePhoto(photo: VehiclePhoto) {
  try {
    await api(`/api/owner/vehicles/${route.params.id}/photos/${photo.id}`, { method: 'DELETE' })
    refresh()
  } catch (err: any) {
    toast.add({ title: 'Delete failed', description: err.data?.error, color: 'error' })
  }
}

function onSaved() {
  toast.add({ title: 'Changes saved' })
  refresh()
}
</script>

<template>
  <UContainer class="py-8 max-w-2xl">
    <div v-if="vehicle">
      <div class="flex items-center gap-3 mb-6">
        <h1 class="text-3xl font-bold">{{ vehicle.make }} {{ vehicle.model }}</h1>
        <StatusBadge :status="vehicle.status" />
      </div>

      <UAlert
        v-if="vehicle.status === 'rejected' && vehicle.rejection_reason"
        color="error"
        variant="subtle"
        title="Listing rejected"
        :description="vehicle.rejection_reason"
        class="mb-6"
      />

      <UCard class="mb-6">
        <template #header><h2 class="font-semibold">Photos</h2></template>

        <div class="grid grid-cols-3 gap-3 mb-4" v-if="vehicle.photos?.length">
          <div v-for="photo in vehicle.photos" :key="photo.id" class="relative group">
            <img :src="photoUrl(photo.file_path)" class="w-full h-24 object-cover rounded">
            <UButton
              icon="i-lucide-x"
              size="xs"
              color="error"
              class="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity"
              @click="deletePhoto(photo)"
            />
          </div>
        </div>
        <p v-else class="text-sm text-gray-500 mb-4">
          No photos yet. Listings with real photos build trust and get approved faster.
        </p>

        <input
          ref="fileInput"
          type="file"
          accept="image/jpeg,image/png,image/webp"
          multiple
          class="hidden"
          @change="uploadPhotos"
        >
        <UButton variant="outline" :loading="uploading" @click="fileInput?.click()">
          Upload photos
        </UButton>
      </UCard>

      <UCard>
        <template #header><h2 class="font-semibold">Details</h2></template>
        <VehicleForm :vehicle="vehicle" @saved="onSaved" />
      </UCard>
    </div>
  </UContainer>
</template>
