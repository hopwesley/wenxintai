<template>
  <Teleport to="body">
    <Transition name="global-loading-fade">
      <div
          v-if="visible"
          class="global-loading-overlay"
      >
        <div class="global-loading-card">
          <div class="global-loading-spinner"></div>
          <p class="global-loading-text">
            {{ message }}
          </p>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import {useGlobalLoading} from '@/controller/useGlobalLoading'

const {visible, message} = useGlobalLoading()
</script>

<style scoped>
.global-loading-overlay {
  position: fixed;
  inset: 0;
  z-index: 2100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.45); /* 深色半透明 */
  backdrop-filter: blur(4px);
}

/* 中央的卡片 */
.global-loading-card {
  min-width: 220px;
  max-width: 320px;
  padding: 20px 24px 18px;
  border-radius: 16px;
  background: rgba(15, 23, 42, 0.92);
  box-shadow: 0 18px 45px rgba(15, 23, 42, 0.7);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

/* 旋转圈圈 */
.global-loading-spinner {
  width: 40px;
  height: 40px;
  border-radius: 999px;
  border: 3px solid rgba(148, 163, 184, 0.45);
  border-top-color: #38bdf8; /* 可以改成你的主题色 */
  animation: global-loading-spin 0.8s linear infinite;
}

/* 文案 */
.global-loading-text {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #e5e7eb;
  text-align: center;
}

/* 渐隐动效 */
.global-loading-fade-enter-active,
.global-loading-fade-leave-active {
  transition: opacity 0.18s ease-out;
}

.global-loading-fade-enter-from,
.global-loading-fade-leave-to {
  opacity: 0;
}

@keyframes global-loading-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
