<script setup lang="ts">
/**
 * The mark: a licence plate whose checkmark breaks out through the top edge —
 * inspected, and going somewhere. The plate stroke is masked where the check
 * crosses it so the two read as one object rather than an overlap.
 *
 * `name` is the only place the brand name appears; change it here to rename.
 */
withDefaults(defineProps<{
  name?: string
  showName?: boolean
  /** Drops the wordmark on narrow screens where nav space is tight. */
  compactOnMobile?: boolean
  size?: number
}>(), {
  name: 'Yan',
  showName: true,
  compactOnMobile: false,
  size: 26
})

// Unique per instance so several logos on one page don't share a mask.
const maskId = `plate-${useId()}`
</script>

<template>
  <span class="inline-flex items-center gap-2.5 select-none">
    <svg
      :width="size * 1.54"
      :height="size"
      viewBox="0 0 40 26"
      fill="none"
      aria-hidden="true"
      class="overflow-visible"
    >
      <mask :id="maskId">
        <rect width="40" height="26" fill="white" />
        <!-- Clears the plate stroke along the check's path. -->
        <path
          d="M9 15.5 L14.5 21 L33 2"
          stroke="black"
          stroke-width="7"
          stroke-linecap="round"
          fill="none"
        />
      </mask>

      <rect
        x="1.6" y="7" width="36.8" height="17" rx="3.6"
        stroke="currentColor"
        stroke-width="2.1"
        :mask="`url(#${maskId})`"
      />

      <path
        d="M9 15.5 L14.5 21 L33 2"
        stroke="#f2a93b"
        stroke-width="3.1"
        stroke-linecap="round"
        stroke-linejoin="round"
        fill="none"
      />
    </svg>

    <span
      v-if="showName"
      class="font-display font-bold"
      :class="compactOnMobile ? 'hidden sm:inline' : ''"
      :style="{ fontSize: `${size * 0.66}px` }"
    >{{ name }}</span>
  </span>
</template>
