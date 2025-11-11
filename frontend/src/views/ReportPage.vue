<template>
  <main class="page">
    <section class="card">
      <header class="header">
        <div>
          <h1>评测报告</h1>
          <p v-if="progress" class="muted">当前状态：{{ progress.label }} ({{ progress.status }})</p>
        </div>
        <button type="button" class="secondary" @click="goHome">返回首页</button>
      </header>

      <p v-if="loading" class="muted">报告加载中…</p>
      <p v-else-if="errorMessage" class="error">{{ errorMessage }}</p>
      <div v-else class="report">
        <p v-if="report?.summary" class="summary">摘要：{{ report.summary }}</p>
        <pre class="payload">{{ formattedReport }}</pre>
      </div>

      <div class="actions">
        <button type="button" class="secondary" @click="resetFlow">清除本地缓存</button>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getReport, getProgress, type ReportResponse, type ProgressResponse } from '@/api/assessment'
import { clearQuestionSets, clearAll, recordReport } from '@/store/assessmentFlow'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const errorMessage = ref('')
const report = ref<ReportResponse | null>(null)
const progress = ref<ProgressResponse | null>(null)

const formattedReport = computed(() => {
  if (!report.value?.full) return '空报告'
  try {
    return JSON.stringify(report.value.full, null, 2)
  } catch (error) {
    return String(report.value.full)
  }
})

watch(
  () => route.params.assessmentId,
  async (value) => {
    if (!value || typeof value !== 'string') {
      return
    }
    await loadReport(value)
  },
  { immediate: true }
)

async function loadReport(assessmentId: string) {
  loading.value = true
  errorMessage.value = ''
  try {
    const [reportData, progressData] = await Promise.all([
      getReport(assessmentId),
      getProgress(assessmentId)
    ])
    report.value = reportData
    progress.value = progressData
    recordReport(reportData.report_id)
  } catch (error) {
    console.error('[ReportPage] failed to load report', error)
    if (error instanceof Error) {
      errorMessage.value = error.message
    } else {
      errorMessage.value = '报告获取失败，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}

function goHome() {
  router.push({ name: 'assessment-start' })
}

function resetFlow() {
  clearQuestionSets()
  clearAll()
  router.push({ name: 'assessment-start' })
}
</script>

<style scoped>
.page {
  margin: 0 auto;
  max-width: 800px;
  padding: 32px 16px 64px;
}

.card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.08);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.header h1 {
  margin: 0;
  color: #111827;
}

.muted {
  color: #6b7280;
  margin: 0;
}

.summary {
  margin: 0;
  color: #1f2937;
  font-weight: 500;
}

.payload {
  margin: 0;
  background: #0f172a;
  color: #f8fafc;
  padding: 16px;
  border-radius: 12px;
  max-height: 420px;
  overflow: auto;
  font-size: 13px;
  line-height: 1.5;
}

.actions {
  display: flex;
  justify-content: flex-end;
}

button {
  font: inherit;
  padding: 10px 16px;
  border-radius: 8px;
  border: none;
  background: #2563eb;
  color: #fff;
  cursor: pointer;
}

button.secondary {
  background: #e5e7eb;
  color: #111827;
}

.error {
  color: #dc2626;
  margin: 0;
}
</style>
