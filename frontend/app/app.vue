<script setup lang="ts">
// Site-wide head defaults. Pages set their own title; the template appends
// the brand. Pages without a title get the full marketing line instead.
useHead({
  htmlAttrs: { lang: 'en' },
  titleTemplate: (title?: string) =>
    title ? `${title} · Yan` : 'Yan — inspected car & motorbike rental in Cambodia',
  link: [{ rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }],
  meta: [{ name: 'theme-color', content: '#fcfbf8' }]
})

useSeoMeta({
  description:
    'Rent cars and motorbikes in Phnom Penh and Siem Reap. Every vehicle inspected before it goes live — real photos, full price up front, book in three steps.',
  ogSiteName: 'Yan',
  ogType: 'website'
})
</script>

<template>
  <UApp>
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>
