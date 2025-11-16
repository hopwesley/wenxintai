<template>
  <TestLayout :key="route.fullPath">
    <template #header>
      <StepIndicator :steps="stepItems" :current="currentStep"/>
    </template>

    <section class="questions">
      <header class="questions__header">
        <h1>{{ currentStepTitle }}</h1>
        <p class="questions__progress" v-if="totalPages > 1">{{ currentPage }} / {{ totalPages }}</p>
      </header>

      <div v-if="loading" class="questions__loading">正在从服务器获取信息…</div>
      <div v-else-if="errorMessage" class="questions__error">{{ errorMessage }}</div>
      <div v-else>
        <form @submit.prevent="handleNext">
          <div v-for="q in currentPageQuestions" :key="q.id" class="question"
               :class="{ 'question--highlight': highlightedId === q.id }" ref="setRef">
            <p class="question__text">{{ q.text }}</p>
            <div class="question__options">
              <label v-for="opt in scaleOptions" :key="opt.value" class="question__option">
                <input type="radio" :name="q.id" :value="opt.value" :checked="getAnswer(q.id) === opt.value"
                       @change="onSelect(q.id, opt.value)"/>
                <span>{{ opt.label }}</span>
              </label>
            </div>
          </div>
        </form>
      </div>

      <footer class="questions__footer" v-if="!loading && !errorMessage">
        <button type="button" class="questions__nav questions__nav--prev" @click="handlePrev">上一步</button>
        <button type="button" class="questions__nav questions__nav--next"
                :disabled="!isCurrentPageComplete || submitting" @click="handleNext">
          <span v-if="submitting">提交中…</span>
          <span v-else>{{ nextLabel }}</span>
        </button>
      </footer>
    </section>

    <!-- 全屏遮罩 -->
    <div v-if="loading" class="overlay">
      <div class="overlay__card">正在从服务器获取信息…</div>
    </div>
  </TestLayout>
</template>
<script setup lang="ts">
import {onMounted, ref} from 'vue'
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/views/components/StepIndicator.vue'
import {applyTest, useQuestionsStageView} from '@/controller/AssessmentQuestions'
import {useTestSession} from "@/store/testSession";
import {StageBasic, TestTypeBasic} from "@/controller/common";

const {
  route,
  loading,
  stepItems,
  currentStep,
  currentStepTitle,
  showLoading,
  hideLoading,
} = useQuestionsStageView()

const totalPages = ref(1)
const currentPage = ref(1)
const currentPageQuestions = ref<{ id: string; text: string }[]>([])
const errorMessage = ref('')
const submitting = ref(false)
const nextLabel = ref('下一步')
const isCurrentPageComplete = ref(true)
const highlightedId = ref<string | null>(null)
const scaleOptions = ref<{ value: number; label: string }[]>([])


function getAnswer(_id: string) {
  return undefined
}

function onSelect(_id: string, _value: number) {
}

async function handlePrev() {
}

async function handleNext() {
}

const {state} = useTestSession()
onMounted(async () => {
  showLoading()
  errorMessage.value = ''

  const scaleKey = String(route.params.scale ?? StageBasic)
  const testType = state.testType || TestTypeBasic

  try {
    const resp = await applyTest(scaleKey, {
      test_type: testType,
      invite_code: state.inviteCode || undefined,
      wechat_openid: state.wechatOpenId || undefined,
      grade: state.grade || undefined,
      mode: state.mode || undefined,
      hobby: state.hobby || undefined,
      session_id: state.sessionId || undefined,
    })

    console.log('[QuestionsStageView] apply_test resp:', resp)

  } catch (err) {
    console.error('[QuestionsStageView] applyTest error', err)
    errorMessage.value = '初始化测试失败，请返回首页重试'
    hideLoading()
  }finally {
    hideLoading()
  }
})
</script>


<style scoped src="@/styles/questions-stage.css"></style>
