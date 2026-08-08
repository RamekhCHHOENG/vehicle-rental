<script setup lang="ts">
const { user, fetchMe, logout } = useAuth()

// Load session once on app start (works for SSR + client nav)
await useAsyncData('auth-init', async () => {
  if (!user.value) await fetchMe()
  return true
})
</script>

<template>
  <div class="min-h-screen flex flex-col bg-white dark:bg-black text-neutral-900 dark:text-neutral-50">
    <header class="dk-nav sticky top-0 z-40 border-b border-neutral-900/10 dark:border-white/10">
      <UContainer class="flex items-center justify-between h-13">
        <NuxtLink to="/" class="text-[17px] font-semibold tracking-tight">
          🚗 CarRental
        </NuxtLink>

        <nav class="flex items-center gap-1 text-[13px]">
          <UButton to="/vehicles" variant="ghost" color="neutral" size="sm">Browse</UButton>

          <template v-if="user">
            <UButton v-if="user.role === 'renter'" to="/dashboard/bookings" variant="ghost" color="neutral" size="sm">
              My bookings
            </UButton>
            <UButton v-if="user.role === 'owner'" to="/owner" variant="ghost" color="neutral" size="sm">
              My vehicles
            </UButton>
            <UButton v-if="user.role === 'owner'" to="/owner/bookings" variant="ghost" color="neutral" size="sm">
              Requests
            </UButton>
            <UButton v-if="user.role === 'admin'" to="/admin" variant="ghost" color="neutral" size="sm">
              Admin
            </UButton>
            <span class="hidden sm:inline text-[13px] text-neutral-500 dark:text-neutral-400 px-2">
              {{ user.full_name }}
            </span>
            <UButton variant="soft" color="neutral" size="sm" @click="logout">Log out</UButton>
          </template>
          <template v-else>
            <UButton to="/login" variant="ghost" color="neutral" size="sm">Log in</UButton>
            <UButton to="/register" size="sm">Sign up</UButton>
          </template>
        </nav>
      </UContainer>
    </header>

    <main class="flex-1">
      <slot />
    </main>

    <footer class="border-t border-neutral-900/10 dark:border-white/10 py-8 text-center">
      <p class="text-[12px] text-neutral-400 dark:text-neutral-500">
        CarRental — verified vehicles, transparent prices.
      </p>
    </footer>
  </div>
</template>
