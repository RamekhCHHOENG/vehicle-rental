<script setup lang="ts">
const colorMode = useColorMode()

// Rendered only after mount: the server can't know the visitor's stored theme,
// so drawing the icon during SSR guarantees a hydration mismatch.
const mounted = ref(false)
onMounted(() => { mounted.value = true })

const isDark = computed({
  get: () => colorMode.value === 'dark',
  set: (dark) => { colorMode.preference = dark ? 'dark' : 'light' }
})
</script>

<template>
  <ClientOnly>
    <UButton
      :icon="isDark ? 'i-lucide-sun' : 'i-lucide-moon'"
      color="neutral"
      variant="ghost"
      size="sm"
      :aria-label="isDark ? 'Switch to light theme' : 'Switch to dark theme'"
      @click="isDark = !isDark"
    />
    <template #fallback>
      <div class="w-8 h-8" />
    </template>
  </ClientOnly>
</template>
