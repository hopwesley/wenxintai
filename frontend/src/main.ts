import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import '@/styles/base.css'
const app = createApp(App)
const pinia = createPinia()

app.use(pinia)   // ← 先挂 Pinia
app.use(router)  // ← 再挂 Router（顺序和这个没硬性要求，只要在 mount 前）
app.mount('#app')
