<template>
  <div class="overlay overlay--ai">
    <div class="overlay__card overlay__card--ai">
      <div class="overlay__title">{{ title }}</div>
      <div class="overlay__subtitle" v-if="subtitle">
        {{ subtitle }}
      </div>

      <!-- 动态能量条 -->
      <div class="overlay__pulse">
        <span class="overlay__dot"></span>
        <span class="overlay__bar"></span>
        <span class="overlay__bar overlay__bar--delay"></span>
      </div>

      <!-- 日志窗口 -->
      <div v-if="logLines && logLines.length" class="overlay__log-window">
        <p
            v-for="(line, idx) in logLines"
            :key="idx"
            class="overlay__log-text"
        >
          {{ line }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  title: string
  subtitle?: string
  logLines?: string[]
}>()
</script>

<!-- 方式一：共用原来的样式文件 -->
<style scoped src="@/styles/ai_loading.css"></style>
