// vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            // 直接使用 import.meta.url + new URL()
            '@': new URL('./src', import.meta.url).pathname,
        },
    },
    server: {
        port: 5173,
        host: '0.0.0.0',  // ✅ 让外网（包括 ngrok）可以访问
        allowedHosts: [
            'sharp-happy-grouse.ngrok-free.app',  // ✅ 你的 ngrok 域名
        ],
        proxy: {
            '/api': {
                target: 'https://www.wenxintai.cn/',
                changeOrigin: true,
            },
        },
    },
})