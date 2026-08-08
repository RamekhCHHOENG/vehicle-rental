// Forwards every /api request to the Go API over the compose network.
//
// This is what lets the site and the API share one domain: the browser only
// ever talks to this origin, so there is no CORS exchange and no cross-site
// cookie to worry about. Server-side rendering does not come through here —
// `useApi` already calls the API directly on the internal address.
export default defineEventHandler((event) => {
  const target = useRuntimeConfig().apiInternal
  if (!target) {
    throw createError({
      statusCode: 503,
      statusMessage: 'NUXT_API_INTERNAL is not set, so this server cannot reach the API'
    })
  }
  // event.path keeps the /api prefix and the query string, and the Go router
  // mounts its routes under /api, so it is forwarded unchanged.
  return proxyRequest(event, target + event.path)
})
