import type { VehicleModelRef } from '~/types'

/**
 * One make's models, fetched when a make is chosen.
 *
 * Models are not part of `/api/metadata`: the catalogue imported from NHTSA runs
 * to a few thousand makes, and shipping every model with every page load would
 * be tens of megabytes. Keyed per make and type so switching back to a make
 * already looked at is instant.
 */
export function useModels(
  makeId: Ref<string | undefined>,
  type: Ref<string | undefined>
) {
  const api = useApi()

  return useAsyncData<VehicleModelRef[]>(
    () => `models-${makeId.value || 'none'}-${type.value || 'any'}`,
    () => {
      if (!makeId.value) return Promise.resolve([])
      return api<VehicleModelRef[]>(`/api/makes/${makeId.value}/models`, {
        query: { type: type.value || undefined }
      })
    },
    { default: () => [], watch: [makeId, type] }
  )
}
