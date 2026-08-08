<script setup lang="ts">
const { register } = useAuth()
const toast = useToast()

const state = reactive({
  full_name: '',
  email: '',
  phone: '',
  password: '',
  role: 'renter' as 'renter' | 'owner'
})
const loading = ref(false)

const roleOptions = [
  { label: 'I want to rent a vehicle', value: 'renter' },
  { label: 'I own vehicles and want to list them', value: 'owner' }
]

async function onSubmit() {
  loading.value = true
  try {
    await register({ ...state })
    navigateTo(state.role === 'owner' ? '/owner' : '/vehicles')
  } catch (err: any) {
    toast.add({ title: 'Sign up failed', description: err.data?.error ?? 'Something went wrong', color: 'error' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UContainer class="py-16 max-w-md">
    <UCard>
      <template #header>
        <h1 class="text-2xl font-bold">Create an account</h1>
      </template>

      <form class="space-y-4" @submit.prevent="onSubmit">
        <UFormField label="Full name" required>
          <UInput v-model="state.full_name" class="w-full" />
        </UFormField>
        <UFormField label="Email" required>
          <UInput v-model="state.email" type="email" class="w-full" />
        </UFormField>
        <UFormField label="Phone">
          <UInput v-model="state.phone" class="w-full" />
        </UFormField>
        <UFormField label="Password" required hint="At least 8 characters">
          <UInput v-model="state.password" type="password" class="w-full" />
        </UFormField>
        <UFormField label="I am here to..." required>
          <URadioGroup v-model="state.role" :items="roleOptions" />
        </UFormField>
        <UButton type="submit" block :loading="loading">Sign up</UButton>
      </form>

      <template #footer>
        <p class="text-sm text-gray-500">
          Already have an account?
          <NuxtLink to="/login" class="text-primary font-medium">Log in</NuxtLink>
        </p>
      </template>
    </UCard>
  </UContainer>
</template>
