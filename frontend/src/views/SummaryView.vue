<template>
  <div class="summary">
    <h1>结果概览</h1>
    <p>您的答案已提交。可点击下方按钮生成报告。</p>
    <button @click="toReport">生成报告</button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
function toReport() {
  router.push('/report')
}
</script>