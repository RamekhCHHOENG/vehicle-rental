<script setup lang="ts">
definePageMeta({ middleware: 'owner' })

const toast = useToast()

function onSaved(vehicle: any) {
  toast.add({
    title: 'Listing created',
    description: 'Now add photos — listings with real photos get approved faster.'
  })
  navigateTo(`/owner/vehicles/${vehicle.id}`)
}
</script>

<template>
  <UContainer class="py-8 max-w-2xl">
    <h1 class="text-3xl font-bold mb-2">List a new vehicle</h1>
    <p class="text-gray-500 mb-6">
      Your listing will be reviewed by our team before it appears publicly.
    </p>
    <UCard>
      <VehicleForm @saved="onSaved" />
    </UCard>
  </UContainer>
</template>
