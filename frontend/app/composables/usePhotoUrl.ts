// Turns an API-relative photo path (/uploads/xyz.jpg) into a full URL.
export function usePhotoUrl() {
  const config = useRuntimeConfig()
  return (path?: string) =>
    path ? `${config.public.apiBase}${path}` : '/placeholder-vehicle.svg'
}
