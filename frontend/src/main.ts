import { createApp } from 'vue'
import App from './App.vue'
import { router } from './router'
import { createI18n } from './i18n'
import '@/styles/base.css'

const i18n = createI18n()

createApp(App).use(router).use(i18n).mount('#app')
