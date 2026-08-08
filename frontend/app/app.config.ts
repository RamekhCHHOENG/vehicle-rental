export default defineAppConfig({
  ui: {
    colors: {
      primary: 'saffron',
      neutral: 'petrol'
    },
    // Radii follow the scale in main.css. Saffron is a light accent, so solid
    // buttons carry ink text rather than white.
    button: {
      slots: {
        base: 'rounded-[var(--r-control)] font-semibold tracking-[-0.01em] transition-all duration-150 active:scale-[0.985] justify-center'
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
        base: 'rounded-[var(--r-small)] font-semibold'
      }
    },
    tabs: {
      slots: {
        list: 'rounded-[var(--r-control)]',
        indicator: 'bg-saffron-400 rounded-[calc(var(--r-control)-2px)]',
        trigger: 'font-medium'
      }
    },
    card: {
      slots: {
        root: 'rounded-[var(--r-surface)]'
      }
    },
    input: { slots: { base: 'rounded-[var(--r-control)]' } },
    select: { slots: { base: 'rounded-[var(--r-control)]' } },
    textarea: { slots: { base: 'rounded-[var(--r-control)]' } },
    modal: { slots: { content: 'rounded-[var(--r-surface)]' } }
  }
})
