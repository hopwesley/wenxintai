<template>
  <router-view />
  <GlobalAlert />
  <WeChatLoginDialog v-model:open="wechatLoginOpen" />
</template>

<script setup lang="ts">
import GlobalAlert from '@/views/components/GlobalAlert.vue'
import WeChatLoginDialog from '@/views/components/WeChatLoginDialog.vue'
import { useAuthStore } from '@/store/auth'
import { storeToRefs } from 'pinia'

const authStore = useAuthStore()
const { wechatLoginOpen } = storeToRefs(authStore)
</script>

<style>
body {
  margin: 0;
  font-family: sans-serif;
}
</style>