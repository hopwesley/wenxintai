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
            'sharp-happy-grouse.ngrok-free.app',
            'primary-exciting-thrush.ngrok-free.app',
            'nonseparable-overneglectful-marlin.ngrok-free.dev',
        ],
        proxy: {
            '/api': {
                target: 'http://localhost:8080/',
                changeOrigin: true,
            },
        },
    },

    build: {
        // 把警告阈值从默认 500KB 提高到 1500KB（1.5MB），看自己需求可以再调
        chunkSizeWarningLimit: 1500,

        // 简单的手动拆包策略
        rollupOptions: {
            output: {
                manualChunks(id) {
                    // 所有来自 node_modules 的包单独拆
                    if (id.includes('node_modules')) {
                        if (id.includes('vue')) {
                            // vue / vue-router 相关放一个包里
                            return 'vue-vendor'
                        }
                        if (id.includes('echarts')) {
                            // 如果你用了 echarts，就单独拆一个包
                            return 'echarts'
                        }
                        // 其他三方依赖统一打进 vendor
                        return 'vendor'
                    }
                },
            },
        },
    },
})