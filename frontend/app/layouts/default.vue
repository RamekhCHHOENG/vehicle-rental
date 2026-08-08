<script setup lang="ts">
const { user, fetchMe, logout } = useAuth()

// Load session once on app start (works for SSR + client nav)
await useAsyncData('auth-init', async () => {
  if (!user.value) await fetchMe()
  return true
})
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <header class="border-b border-gray-200 dark:border-gray-800">
      <UContainer class="flex items-center justify-between h-16">
        <NuxtLink to="/" class="text-xl font-bold">🚗 CarRental</NuxtLink>

        <nav class="flex items-center gap-2">
          <UButton to="/vehicles" variant="ghost" color="neutral">Browse</UButton>

          <template v-if="user">
            <UButton v-if="user.role === 'renter'" to="/dashboard/bookings" variant="ghost" color="neutral">
              My bookings
            </UButton>
            <UButton v-if="user.role === 'owner'" to="/owner" variant="ghost" color="neutral">
              My vehicles
            </UButton>
            <UButton v-if="user.role === 'owner'" to="/owner/bookings" variant="ghost" color="neutral">
              Requests
            </UButton>
            <UButton v-if="user.role === 'admin'" to="/admin" variant="ghost" color="neutral">
              Admin
            </UButton>
            <span class="text-sm text-gray-500 px-2">{{ user.full_name }}</span>
            <UButton variant="outline" color="neutral" @click="logout">Log out</UButton>
          </template>
          <template v-else>
            <UButton to="/login" variant="ghost" color="neutral">Log in</UButton>
            <UButton to="/register">Sign up</UButton>
          </template>
        </nav>
      </UContainer>
    </header>

    <main class="flex-1">
      <slot />
    </main>

    <footer class="border-t border-gray-200 dark:border-gray-800 py-6 text-center text-sm text-gray-500">
      CarRental — verified vehicles, transparent prices.
    </footer>
  </div>
</template>
