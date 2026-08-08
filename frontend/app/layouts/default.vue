<script setup lang="ts">
const { user, fetchMe, logout } = useAuth()

await useAsyncData('auth-init', async () => {
  if (!user.value) await fetchMe()
  return true
})
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <header class="dk-nav sticky top-0 z-40 border-b border-[var(--ui-border-muted)]">
      <UContainer class="flex items-center justify-between h-14 gap-4">
        <NuxtLink to="/" class="shrink-0" aria-label="Yan home">
          <AppLogo compact-on-mobile />
        </NuxtLink>

        <nav class="flex items-center gap-0.5 text-[13px] [&_a]:whitespace-nowrap [&_button]:whitespace-nowrap">
          <UButton to="/vehicles" variant="ghost" color="neutral" size="sm">Browse</UButton>

          <template v-if="user">
            <UButton v-if="user.role === 'renter'" to="/dashboard/bookings" variant="ghost" color="neutral" size="sm">
              Bookings
            </UButton>
            <UButton v-if="user.role === 'owner'" to="/owner" variant="ghost" color="neutral" size="sm">
              Vehicles
            </UButton>
            <UButton v-if="user.role === 'owner'" to="/owner/bookings" variant="ghost" color="neutral" size="sm">
              Requests
            </UButton>
            <UButton v-if="user.role === 'admin'" to="/admin" variant="ghost" color="neutral" size="sm">
              Admin
            </UButton>
            <ThemeToggle />
            <span class="hidden md:inline text-[13px] text-[var(--ui-text-muted)] px-2">
              {{ user.full_name }}
            </span>
            <UButton variant="soft" color="neutral" size="sm" @click="logout">Log out</UButton>
          </template>

          <template v-else>
            <ThemeToggle />
            <UButton to="/login" variant="ghost" color="neutral" size="sm">Log in</UButton>
            <UButton to="/register" size="sm">Sign up</UButton>
          </template>
        </nav>
      </UContainer>
    </header>

    <main class="flex-1">
      <slot />
    </main>

    <footer class="border-t border-[var(--ui-border-muted)] mt-20">
      <UContainer class="py-10 flex flex-wrap items-center justify-between gap-4">
        <div class="flex items-center gap-2.5">
          <AppLogo :size="20" />
          <span class="text-[13px] text-[var(--ui-text-dimmed)]">យាន</span>
        </div>
        <p class="text-[12px] text-[var(--ui-text-dimmed)]">
          Every vehicle inspected before it goes live.
        </p>
      </UContainer>
    </footer>
  </div>
</template>
