<script setup lang="ts">
/**
 * The signature element: an inspection stamp. Use one per surface, over
 * vehicle photography — it makes the platform's one real advantage visible.
 */
const props = withDefaults(defineProps<{
  /** ISO date the listing was approved; shown as a short "checked" date. */
  date?: string
  large?: boolean
  /** Set when the seal sits over an image, so it gets its own dark ground. */
  onPhoto?: boolean
}>(), { large: false, onPhoto: false })

const checked = computed(() => {
  if (!props.date) return null
  return new Date(props.date).toLocaleDateString('en-GB', { month: 'short', year: 'numeric' })
})
</script>

<template>
  <span class="seal" :class="{ 'seal--lg': large, 'seal--photo': onPhoto }">
    <svg :width="large ? 14 : 11" :height="large ? 14 : 11" viewBox="0 0 16 16" fill="none" aria-hidden="true">
      <path
        d="M3.5 8.5 L6.5 11.5 L12.5 4.5"
        stroke="currentColor"
        stroke-width="2.4"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
    <span>Inspected<template v-if="checked"> · {{ checked }}</template></span>
  </span>
</template>
