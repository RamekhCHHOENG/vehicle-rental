// $fetch wrapper that always targets the Go API and sends the auth cookie.
export function useApi() {
  const config = useRuntimeConfig()

  return $fetch.create({
    baseURL: config.public.apiBase,
    credentials: 'include'
  })
}
