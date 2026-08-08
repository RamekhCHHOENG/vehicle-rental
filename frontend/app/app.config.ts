export default defineAppConfig({
  ui: {
    colors: {
      primary: 'saffron',
      neutral: 'petrol'
    },
    // Pill buttons with the designkit press effect. Saffron is a light
    // accent, so solid buttons carry ink text rather than white.
    button: {
      slots: {
        base: 'rounded-full font-semibold tracking-[-0.01em] transition-all duration-150 active:scale-[0.975] justify-center'
      },
      variants: {
        color: {
          primary: ''
        }
      },
      compoundVariants: [
        {
          color: 'primary',
          variant: 'solid',
          class: 'text-petrol-900 bg-saffron-400 hover:bg-saffron-300 focus-visible:outline-saffron-400'
        }
      ]
    },
    badge: {
      slots: {
        base: 'rounded-full font-semibold'
      }
    },
    tabs: {
      slots: {
        indicator: 'bg-saffron-400',
        trigger: 'font-medium'
      }
    },
    card: {
      slots: {
        root: 'rounded-[20px]'
      }
    },
    input: { slots: { base: 'rounded-[11px]' } },
    select: { slots: { base: 'rounded-[11px]' } },
    textarea: { slots: { base: 'rounded-[11px]' } },
    modal: { slots: { content: 'rounded-[22px]' } }
  }
})
