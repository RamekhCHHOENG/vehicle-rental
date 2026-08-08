// Every deploy rehashes the built chunks, so a tab left open across one asks
// for filenames that no longer exist and dies with "Failed to fetch
// dynamically imported module". Nuxt recovers from that when it happens during
// route navigation, but not when it happens while a plugin is still starting
// up — which is where ours lands, so the app never finishes mounting.
//
// Reloading picks up the current build. The timestamp guard is what keeps a
// chunk that is genuinely missing from turning into a refresh loop.
export default defineNuxtPlugin(() => {
  const KEY = 'yan:chunk-reload-at'
  const COOLDOWN = 10_000

  window.addEventListener('vite:preloadError', (event) => {
    // Without this Vite rethrows, and the reload never gets to happen.
    event.preventDefault()

    const last = Number(sessionStorage.getItem(KEY) ?? 0)
    if (Date.now() - last < COOLDOWN) return

    sessionStorage.setItem(KEY, String(Date.now()))
    window.location.reload()
  })
})
