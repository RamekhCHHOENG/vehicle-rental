// $fetch wrapper that always targets the Go API and sends the auth cookie.
// In the browser, credentials:'include' attaches the cookie automatically.
// During SSR the request runs on the Nuxt server, so we must forward the
// browser's cookie header to the API by hand.
export function useApi() {
  const config = useRuntimeConfig()
  const headers = import.meta.server ? useRequestHeaders(['cookie']) : undefined

  return $fetch.create({
    baseURL: config.public.apiBase,
    credentials: 'include',
    headers
  })
}
