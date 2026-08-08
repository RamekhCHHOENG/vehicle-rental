// Uploaded vehicle photos, served by the Go API off its uploads volume.
//
// Same reason as server/api: one domain for the whole site, so `usePhotoUrl`
// can build a same-origin URL and the browser never leaves this host.
export default defineEventHandler((event) => {
  const target = useRuntimeConfig().apiInternal
  if (!target) {
    throw createError({
      statusCode: 503,
      statusMessage: 'NUXT_API_INTERNAL is not set, so this server cannot reach the API'
    })
  }
  // The Go side serves these under /uploads too, so the path passes through.
  return proxyRequest(event, target + event.path)
})
