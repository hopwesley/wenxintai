<template>
  <TestLayout>
    <template #header>
      <StepIndicator :steps="stepItems" :current="state.currentStep" />
    </template>

    <section class="report">
      <h1>{{ currentStepTitle }}</h1>
      <div v-if="isPlaceholderStep" class="report__placeholder">report__placeholder</div>
      <template v-else>
        <div v-if="loading" class="report__loading">正在生成报告…</div>
        <div v-else-if="errorMessage" class="report__error">{{ errorMessage }}</div>
        <pre v-else-if="reportText" class="report__content">{{ reportText }}</pre>
        <p v-else class="report__empty">暂无报告内容</p>
      </template>
    </section>

    <template #footer>
      <p>免责声明 基于AI生成，仅供参考</p>
    </template>
  </TestLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/components/StepIndicator.vue'
import { getReport } from '@/api'
import { useTestSession } from '@/store/testSession'
import { STEPS, isVariant, type Variant } from '@/config/testSteps'

const route = useRoute()
const router = useRouter()
const { state, setVariant, setCurrentStep, getSessionId } = useTestSession()
const variant = ref<Variant>('basic')

const loading = ref(true)
const reportText = ref('')
const errorMessage = ref('')
const isPlaceholderStep = computed(() => variant.value === 'pro')

watchEffect(() => {
  const variantParam = String(route.params.variant ?? 'basic')
  if (!isVariant(variantParam)) {
    router.replace({ path: '/test/basic/step/1' })
    return
  }
  variant.value = variantParam
  setVariant(variant.value)
})

watchEffect(() => {
  const stepNumber = Number(route.params.step ?? '4')
  if (Number.isNaN(stepNumber) || stepNumber < 1 || stepNumber > STEPS[variant.value].length) {
    router.replace({ path: `/test/${variant.value}/step/1` })
    return
  }
  setCurrentStep(stepNumber)
})

const stepItems = computed(() =>
  STEPS[variant.value].map((item) => ({ key: item.key, title: (item.titleKey) }))
)

const currentStepTitle = computed(() => {
  const index = state.currentStep - 1
  const item = STEPS[variant.value][index]
  if (!item) return ('report.title')
  return (item.titleKey)
})

onMounted(async () => {
  if (!state.mode || !state.hobby) {
    router.replace({ path: `/test/${variant.value}/step/1` })
    return
  }

  if (!isPlaceholderStep.value && variant.value === 'basic' && Object.keys(state.answersStage2).length === 0) {
    router.replace({ path: `/test/${variant.value}/step/3` })
    return
  }

  if (isPlaceholderStep.value) {
    loading.value = false
    return
  }

  loading.value = true
  errorMessage.value = ''
  reportText.value = ''

  try {
    const sessionId = getSessionId()
    if (!sessionId) {
      if (typeof window !== 'undefined') {
        window.alert('需要邀请码或登录后访问')
      }
      router.replace({ path: '/' })
      return
    }
    const response = await getReport({
      session_id: sessionId,
      mode: state.mode ?? '',
    })
    reportText.value = JSON.stringify(response.report, null, 2)
  } catch (error) {
    console.error('[ReportView] failed to load report', error)
    if (error instanceof Error) {
      errorMessage.value = error.message
      if (error.name === 'NO_SESSION' || error.name === 'INVITE_REQUIRED') {
        if (typeof window !== 'undefined') {
          window.alert(error.message)
        }
        router.replace({ path: '/' })
      }
    } else {
      errorMessage.value = '生成报告失败，请稍后再试'
    }
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.report {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.report h1 {
  margin: 0;
  font-size: 24px;
  color: #1f2937;
}

.report__loading,
.report__error,
.report__empty {
  padding: 24px;
  border-radius: 12px;
  background: #f9fafb;
  color: rgba(30, 41, 59, 0.75);
}

.report__placeholder {
  padding: 24px;
  border-radius: 12px;
  background: rgba(148, 163, 184, 0.12);
  color: rgba(30, 41, 59, 0.75);
}

.report__error {
  color: #dc2626;
}

.report__content {
  background: #0f172a;
  color: #f8fafc;
  padding: 24px;
  border-radius: 12px;
  max-height: 420px;
  overflow: auto;
  font-size: 14px;
}
</style>
