import { createApp } from 'vue'
import App from './App.vue'
import { router } from './router'

// Bootstrap the Vue application. The router is registered as a plugin to
// enable client-side navigation between views. See src/router/index.ts for
// route definitions.
createApp(App).use(router).mount('#app')