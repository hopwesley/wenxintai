<template>
  <div class="report">
    <h1>AI 报告</h1>
    <div v-if="loading">正在生成报告…</div>
    <pre v-else>{{ reportText }}</pre>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getReport } from '../api'
import '@/styles/report.css'

const reportText = ref('')
const loading = ref(true)

onMounted(async () => {
  const sessionId = localStorage.getItem('session_id') || ''
  try {
    const resp = await getReport({ session_id: sessionId, mode: 'A' })
    reportText.value = JSON.stringify(resp.report, null, 2)
  } catch (e) {
    reportText.value = (e as Error).message
  } finally {
    loading.value = false
  }
})
</script>