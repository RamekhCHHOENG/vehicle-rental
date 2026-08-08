<script setup lang="ts">

useSeoMeta({
  title: 'Log in',
  robots: 'noindex'
})
const { login } = useAuth()
const toast = useToast()

const state = reactive({ email: '', password: '' })
const loading = ref(false)

async function onSubmit() {
  loading.value = true
  try {
    await login(state.email, state.password)
    navigateTo('/')
  } catch (err: any) {
    toast.add({ title: 'Login failed', description: err.data?.error ?? 'Something went wrong', color: 'error' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UContainer class="py-16 max-w-md">
    <UCard>
      <template #header>
        <h1 class="text-2xl font-bold">Log in</h1>
      </template>

      <form class="space-y-4" @submit.prevent="onSubmit">
        <UFormField label="Email" required>
          <UInput v-model="state.email" type="email" placeholder="you@example.com" class="w-full" />
        </UFormField>
        <UFormField label="Password" required>
          <UInput v-model="state.password" type="password" class="w-full" />
        </UFormField>
        <UButton type="submit" block :loading="loading">Log in</UButton>
      </form>

      <template #footer>
        <p class="text-sm text-gray-500">
          No account yet?
          <NuxtLink to="/register" class="text-primary font-medium">Sign up</NuxtLink>
        </p>
      </template>
    </UCard>
  </UContainer>
</template>
