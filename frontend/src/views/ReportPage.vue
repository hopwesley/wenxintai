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
        <div v-if="streamingText" class="live">
          <div class="live__header">
            <h2>实时生成</h2>
            <span v-if="statusPhase" class="phase">{{ statusPhase }}</span>
          </div>
          <pre class="payload payload--live">{{ streamingText }}</pre>
          <p class="muted">累计 tokens：{{ tokenCount }}</p>
          <p v-if="streamError" class="error">{{ streamError }}</p>
          <button type="button" class="secondary" @click="stopStream">停止实时更新</button>
        </div>
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
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getReport, getProgress, type ReportResponse, type ProgressResponse } from '@/api/assessment'
import { clearQuestionSets, clearAll, recordReport } from '@/store/assessmentFlow'
import { connect, type StreamClient } from '@/stream/client'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const errorMessage = ref('')
const report = ref<ReportResponse | null>(null)
const progress = ref<ProgressResponse | null>(null)
const streamingText = ref('')
const statusPhase = ref('')
const tokenCount = ref(0)
const streamError = ref('')
const streamClient = ref<StreamClient | null>(null)

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
    await loadInitial(value)
    setupStream(value)
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  streamClient.value?.close()
  streamClient.value = null
})

async function loadInitial(assessmentId: string) {
  loading.value = true
  errorMessage.value = ''
  streamError.value = ''
  try {
    const progressData = await getProgress(assessmentId)
    progress.value = progressData
    try {
      const reportData = await getReport(assessmentId)
      report.value = reportData
      recordReport(reportData.report_id)
    } catch (error) {
      if (error && typeof error === 'object' && 'status' in (error as any) && (error as any).status === 404) {
        report.value = null
      } else if (error instanceof Error) {
        throw error
      }
    }
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

async function refreshReport(assessmentId: string) {
  try {
    const reportData = await getReport(assessmentId)
    report.value = reportData
    recordReport(reportData.report_id)
  } catch (error) {
    console.error('[ReportPage] failed to refresh report', error)
    if (error instanceof Error) {
      errorMessage.value = error.message
    }
  }
}

function setupStream(assessmentId: string) {
  if (typeof window === 'undefined') return
  streamClient.value?.close()
  streamingText.value = ''
  statusPhase.value = ''
  tokenCount.value = 0
  streamError.value = ''
  const client = connect(assessmentId)
  streamClient.value = client
  client.onStatus((payload) => {
    statusPhase.value = payload.phase ?? ''
  })
  client.onToken((payload) => {
    if (typeof payload.text === 'string' && payload.text) {
      streamingText.value += payload.text
    }
  })
  client.onProgress((payload) => {
    if (typeof payload.tokens === 'number') {
      tokenCount.value = payload.tokens
    }
  })
  client.onFinal(async ({ report_id }) => {
    streamClient.value?.close()
    streamClient.value = null
    if (report_id) {
      recordReport(report_id)
    }
    await refreshReport(assessmentId)
  })
  client.onError((payload) => {
    streamError.value = payload.message ?? '实时生成出现异常'
  })
}

function goHome() {
  router.push({ name: 'assessment-start' })
}

function resetFlow() {
  clearQuestionSets()
  clearAll()
  router.push({ name: 'assessment-start' })
}

function stopStream() {
  streamClient.value?.close()
  streamClient.value = null
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

.payload--live {
  background: rgba(37, 99, 235, 0.1);
  color: #0f172a;
}

.live {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  border: 1px solid rgba(37, 99, 235, 0.35);
  border-radius: 12px;
}

.live__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.live__header h2 {
  margin: 0;
  font-size: 16px;
  color: #1f2937;
}

.phase {
  font-size: 12px;
  color: #2563eb;
  background: rgba(37, 99, 235, 0.12);
  padding: 2px 8px;
  border-radius: 999px;
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
