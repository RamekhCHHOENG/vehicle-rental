export default defineAppConfig({
  ui: {
    colors: {
      primary: 'apple',
      neutral: 'neutral'
    },
    // Pill buttons with the designkit press effect
    button: {
      slots: {
        base: 'rounded-full font-semibold tracking-[-0.01em] transition-all duration-150 active:scale-[0.975] justify-center'
      }
    },
    badge: {
      slots: {
        base: 'rounded-full font-semibold'
      }
    },
    card: {
      slots: {
        root: 'rounded-[20px]'
      }
    },
    input: {
      slots: {
        base: 'rounded-[11px]'
      }
    },
    select: {
      slots: {
        base: 'rounded-[11px]'
      }
    },
    textarea: {
      slots: {
        base: 'rounded-[11px]'
      }
    },
    modal: {
      slots: {
        content: 'rounded-[22px]'
      }
    }
  }
})
