<!-- src/features/questions-stage/view/QuestionsStageView.vue -->
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
import {applyTest, useQuestionsStageView} from '@/controller/QuestionsStageControl'
import {useTestSession} from "@/store/testSession";

// 1) 从我们刚刚写的 TS 逻辑里拿：route / loading / 步骤条 / 标题
const {
  route,
  loading,
  stepItems,
  currentStep,
  currentStepTitle,
  showLoading,
  hideLoading,
} = useQuestionsStageView()

// 2) 下面这些还是占位，避免模板报错，后面我们再一点点补全真实逻辑

// 分页占位
const totalPages = ref(1)
const currentPage = ref(1)
const currentPageQuestions = ref<{ id: string; text: string }[]>([])

// 错误 & 提交状态
const errorMessage = ref('')
const submitting = ref(false)
const nextLabel = ref('下一步')

// 当前页是否完成
const isCurrentPageComplete = ref(true)

// 题目 & 选项占位
const highlightedId = ref<string | null>(null)
const scaleOptions = ref<{ value: number; label: string }[]>([])

function setRef() {
}

function getAnswer(_id: string) {
  return undefined
}

function onSelect(_id: string, _value: number) {
}

// 上一步 / 下一步 先留空实现，后面再接路由跳转逻辑
async function handlePrev() {
}

async function handleNext() {
}


const {state /*, setCurrentTestId*/} = useTestSession()
onMounted(async () => {
  showLoading()
  errorMessage.value = ''

  const scaleKey = String(route.params.scale ?? '')  // riasec / asc ...
  const testType = state.testType || 'basic'        // basic / pro 等

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

    // TODO: 把问卷 ID 存到全局 store，后面提交答案 / 查报告都会用到
    // setCurrentTestId(resp.test_id)

    // 注意：这里“只建 tests_record 记录 + 启动 AI”，题目还没生成，
    // 我建议暂时先“不关遮罩”，等后续 SSE / 拉题目成功后再 hideLoading()
    // 如果你现在想先看界面，也可以临时在这里 hideLoading() 试效果。
    // hideLoading()
  } catch (err) {
    console.error('[QuestionsStageView] applyTest error', err)
    errorMessage.value = '初始化测试失败，请返回首页重试'
    hideLoading()
  }
})
</script>


<style scoped src="@/styles/questions-stage.css"></style>
