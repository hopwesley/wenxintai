<template>
  <div class="overlay overlay--ai" :class="{ 'overlay--ai--active': hasLogs }">
    <div class="overlay__card overlay__card--ai">
      <!-- 头部：LIVE 标记 + 标题 + 元信息 -->
      <div class="overlay__header">
        <div class="overlay__live">
          <span class="overlay__live-dot"></span>
          <span class="overlay__live-label">AI 实时分析中</span>
        </div>

        <div class="overlay__title-row">
          <div class="overlay__title">{{ title }}</div>
          <div v-if="meta" class="overlay__meta">
            <span v-if="meta.mode">模式：{{ meta.mode }}</span>
            <span v-if="meta.grade">年级：{{ meta.grade }}</span>
            <span v-if="meta.stage">阶段：{{ meta.stage }}</span>
          </div>
        </div>

        <div class="overlay__subtitle" v-if="subtitle">
          {{ subtitle }}
        </div>
      </div>

      <!-- 动态能量条：有日志后样式会减弱（看 CSS） -->
      <div class="overlay__pulse">
        <span class="overlay__dot"></span>
        <span class="overlay__bar"></span>
        <span class="overlay__bar overlay__bar--delay"></span>
      </div>

      <!-- 日志窗口 -->
      <div v-if="logLines && logLines.length" class="overlay__log-window">
        <div
            v-for="(line, idx) in logLines"
            :key="idx"
            class="overlay__log-row"
        >
          <span class="overlay__log-index">#{{ idx + 1 }}</span>
          <p class="overlay__log-text">
            {{ line }}
          </p>
        </div>
      </div>

      <!-- 还没收到日志时的占位文案（防止误以为是纯动画） -->
      <div v-else class="overlay__log-placeholder">
        正在与 AI 建立连接，准备基于本次回答生成分析…
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title: string
  subtitle?: string
  logLines?: string[]
  meta?: {
    mode?: string
    grade?: string
    stage?: string
  }
}>()

const hasLogs = computed(() => !!props.logLines && props.logLines.length > 0)
</script>

<style scoped>
@import '@/styles/ai_loading.css';
</style>
