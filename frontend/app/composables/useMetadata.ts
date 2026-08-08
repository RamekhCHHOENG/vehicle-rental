import type { Metadata } from '~/types'

/**
 * The vocabulary every listing is written in — provinces, makes and their
 * models, features, and the fixed enums.
 *
 * One shared key, so a page using it in three components fetches once, and the
 * payload carries from server render into the browser instead of being
 * refetched on hydration.
 */
export function useMetadata() {
  const api = useApi()

  return useAsyncData<Metadata>(
    'metadata',
    () => api<Metadata>('/api/metadata'),
    {
      default: () => ({
        provinces: [],
        makes: [],
        features: [],
        vehicle_types: [],
        transmissions: [],
        seat_options: []
      }),
      getCachedData: (key, nuxtApp) => nuxtApp.payload.data[key] ?? nuxtApp.static.data[key]
    }
  )
}
