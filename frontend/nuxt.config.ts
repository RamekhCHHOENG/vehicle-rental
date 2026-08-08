export default defineNuxtConfig({
  modules: ['@nuxt/ui'],
  css: ['~/assets/css/main.css'],
  devtools: { enabled: true },
  runtimeConfig: {
    // Server-only. In Docker the Nuxt container reaches the API over the compose
    // network (http://api:8080), not via the host port. Empty means "same as
    // apiBase", which is correct when running natively. Set NUXT_API_INTERNAL.
    apiInternal: '',
    public: {
      // Browser-facing: must be reachable from the visitor's machine.
      apiBase: 'http://localhost:8090'
    }
  }
})
