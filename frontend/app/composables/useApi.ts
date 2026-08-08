// $fetch wrapper that always targets the Go API and carries the auth cookie.
//
// Two things differ between server and browser:
//   - Base URL. In the browser the API must be reachable from the visitor's
//     machine; during SSR the Nuxt server may need an internal address instead
//     (in Docker, the API is `api:8080` on the compose network).
//   - Cookies. `credentials: 'include'` attaches them automatically in the
//     browser, but during SSR we have to forward the incoming cookie header.
export function useApi() {
  const config = useRuntimeConfig()

  const baseURL = import.meta.server
    ? (config.apiInternal || config.public.apiBase)
    : config.public.apiBase

  const headers = import.meta.server ? useRequestHeaders(['cookie']) : undefined

  return $fetch.create({
    baseURL,
    credentials: 'include',
    headers
  })
}
