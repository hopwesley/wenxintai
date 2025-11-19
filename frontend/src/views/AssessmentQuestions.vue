<template>
  <TestLayout :key="route.fullPath">
    <!-- 顶部步骤条 -->
    <template #header>
      <StepIndicator :steps="stepItems" :current="currentStep"/>
    </template>

    <section class="questions">
      <!-- 顶部标题 + 进度 -->
      <header class="questions__header">
        <h1>{{ currentStepTitle }}</h1>
        <p class="questions__progress" v-if="totalPages > 1">
          第 {{ pageStartIndex + 1 }}–{{ pageEndIndex }} 题 / 共 {{ totalCount }} 题
        </p>
      </header>

      <!-- 主区域：根据 loading / error / 正常显示不同内容 -->
      <div v-if="loading" class="questions__loading">
        正在为你准备本阶段的专属题目…
      </div>
      <div v-else-if="errorMessage" class="questions__error">
        {{ errorMessage }}
      </div>
      <div v-else>
        <!-- 整个答题区域用 form 包裹 -->
        <form @submit.prevent="handleNext">
          <!-- 当前页题目列表：每页 5 题 -->
          <section>
            <article
                v-for="(question, idx) in pagedQuestions"
                :key="question.id"
                class="question"
                :class="{ 'question--highlight': isQuestionHighlighted(question.id) }"
            >
              <!-- 题干：序号 + 文本 -->
              <p class="question__text">
                {{ pageStartIndex + idx + 1 }}. {{ question.text }}
              </p>

              <!-- 选项：5 个尺度 -->
              <div class="question__options">
                <label
                    v-for="opt in scaleOptions"
                    :key="opt.value"
                    class="question__option"
                >
                  <input
                      type="radio"
                      :name="`q-${question.id}`"
                      :value="opt.value"
                      v-model="answers[question.id]"
                  />
                  <span class="question__option-label">
                    {{ opt.label }}
                  </span>
                </label>
              </div>
            </article>
          </section>

          <!-- 底部翻页按钮 -->
          <footer class="questions__footer">
            <button
                v-if="totalPages > 1"
                type="button"
                class="btn btn-secondary questions__nav"
                @click="handlePrev"
                :disabled="isFirstPage || isSubmitting"
            >
              返回上一页
            </button>

            <button
                type="submit"
                class="btn btn-primary questions__nav"
                :disabled="isSubmitting"
            >
              {{ isLastPage ? '提交本阶段' : '下一页' }}
            </button>
          </footer>
        </form>
      </div>
    </section>

    <!-- 🌌 AI 生成题目中的炫酷遮罩：默认 loading=true 时显示 -->
    <div v-if="loading" class="overlay overlay--ai">
      <div class="overlay__card overlay__card--ai">
        <div class="overlay__title">AI 正在为你生成专属题目…</div>
        <div class="overlay__subtitle">
          正在分析你的测试设置，智能规划本阶段题目结构
        </div>

        <!-- 动态能量条 / 小点点，让它看起来在“运算” -->
        <div class="overlay__pulse">
          <span class="overlay__dot"></span>
          <span class="overlay__bar"></span>
          <span class="overlay__bar overlay__bar--delay"></span>
        </div>

        <!-- 日志窗口：固定高度 + 只展示 latestMessage 的尾部片段 -->
        <div v-if="truncatedLatestMessage" class="overlay__log-window">
          <p
              v-for="(line, idx) in truncatedLatestMessage"
              :key="idx"
              class="overlay__log-text"
          >
            {{ line }}
          </p>
        </div>
      </div>
    </div>

    <!-- 提交中的遮罩层（保持简单文案） -->
    <div v-if="isSubmitting" class="overlay">
      <div class="overlay__card">
        正在提交本阶段答案，请稍候…
      </div>
    </div>
  </TestLayout>
</template>

<script setup lang="ts">
import {computed, onMounted, ref} from 'vue'
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/views/components/StepIndicator.vue'
import {useQuestionsStageView} from '@/controller/AssessmentQuestions'
import {useTestSession} from "@/store/testSession";
import {useSubscriptBySSE} from "@/controller/common";
import {useAlert} from "@/controller/useAlert";
import {router} from "@/router";

interface Question {
  id: number;
  text: string;
  dimension: string;
}

interface ScaleOption {
  value: number;
  label: string;
}

// 公共视图逻辑（步骤条、标题、loading 等）
const {
  route,
  loading,
  stepItems,
  currentStep,
  currentStepTitle,
  showLoading,
  hideLoading,
  validateTestStage,
} = useQuestionsStageView()

// --- 答题页自己的状态 ---

const pageSize = 5                      // 每页 5 题
const currentPage = ref(1)              // 当前页，从 1 开始
const questions = ref<Question[]>([])   // SSE 拿到的全部题目
const answers = ref<Record<number, number>>({})  // 每题的作答：question.id -> 1~5
const highlightedQuestions = ref<Record<number, boolean>>({}) // 未作答高亮
const errorMessage = ref('')

const logLines = ref<string[]>([])
const MAX_LOG_LINES = 8
const truncatedLatestMessage = computed(() => logLines.value)

const isSubmitting = ref(false)

// 量表选项：1–5
const scaleOptions = ref<ScaleOption[]>([
  {value: 1, label: '从不'},
  {value: 2, label: '很少'},
  {value: 3, label: '一般'},
  {value: 4, label: '经常'},
  {value: 5, label: '非常多'},
])

// 统计与分页计算
const totalCount = computed(() => questions.value.length)

const totalPages = computed(() => {
  return totalCount.value > 0 ? Math.ceil(totalCount.value / pageSize) : 1
})

const pageStartIndex = computed(() => (currentPage.value - 1) * pageSize)

const pageEndIndex = computed(() =>
    Math.min(pageStartIndex.value + pageSize, totalCount.value)
)

const pagedQuestions = computed(() =>
    questions.value.slice(pageStartIndex.value, pageEndIndex.value)
)

const isFirstPage = computed(() => currentPage.value <= 1)
const isLastPage = computed(() => currentPage.value >= totalPages.value)

function isQuestionHighlighted(id: number): boolean {
  return !!highlightedQuestions.value[id]
}

// 会话 & 路由信息
const {state, getPublicID} = useTestSession()
const public_id: string | undefined = getPublicID()
const {showAlert} = useAlert()
const routes = state.testRoutes ?? []
const {businessType, testStage} = route.params as { businessType: string; testStage: string }

console.log('[QuestionsStageView] apply_test resp:', testStage, businessType, public_id, routes)

// --- 生命周期：挂载时启动 SSE，拉取题目 ---

onMounted(() => {
  errorMessage.value = ''

  // 基础校验：无 public_id / 无 routes / stage 非法时直接返回首页
  if (!public_id || !routes.length || !validateTestStage(testStage)) {
    showAlert('测试流程异常，请返回首页重新开始', () => {
      router.replace('/').then()
    })
    return
  }

  const idx = routes.findIndex(r => r.router === String(testStage || ''))
  if (idx === -1) {
    showAlert('测试流程异常，未能识别当前步骤，请返回首页重新开始', () => {
      router.replace('/').then()
    })
    return
  }

  let message = ''
  const sseCtrl = useSubscriptBySSE(public_id, businessType, testStage, {
    autoStart: false,

    onOpen() {
      showLoading()
    },

    onError(err) {
      console.log('------>>> sse channel error:', err)
      hideLoading()
      showAlert('获取测试流程失败，请稍后再试:' + err)
    },

    // 流式内容累加：全部都进 message，再同步给 latestMessage
    onMsg(chunk) {
      message += chunk
      if (message.length < 20) {
        return
      }
      logLines.value.push(message)
      if (logLines.value.length > MAX_LOG_LINES) {
        logLines.value.splice(0, logLines.value.length - MAX_LOG_LINES)
      }
      message='';
    },

    onClose() {
      console.log('------>>> sse closed:')
      hideLoading()
    },

    // 服务端发送 "done" 时的最终结果
    onDone(questionStr) {
      const raw = (questionStr && questionStr.trim().length > 0) ? questionStr : message
      console.log('------>>> go questions:', raw)

      try {
        const parsed = JSON.parse(raw) as Question[]
        if (!Array.isArray(parsed) || parsed.length === 0) {
          throw new Error('empty questions')
        }

        questions.value = parsed
        currentPage.value = 1
        highlightedQuestions.value = {}
      } catch (e) {
        console.error('[QuestionsStageView] 解析题目失败:', e)
        errorMessage.value = '获取测试题目失败，请稍后再试'
        showAlert('获取测试题目失败，请稍后再试')
      } finally {
        hideLoading()
      }
    },
  })

  sseCtrl.start()
})

// --- 翻页逻辑 ---

function handlePrev() {
  if (isFirstPage.value || isSubmitting.value) return
  currentPage.value -= 1
  highlightedQuestions.value = {}   // 切页时清掉高亮
}

// 下一页 / 提交
async function handleNext() {
  if (!questions.value.length) {
    return
  }

  // 1. 先校验当前页是否全部作答
  const pageQs = pagedQuestions.value
  const missingIds: number[] = []

  for (const q of pageQs) {
    const v = answers.value[q.id]
    if (v == null) {
      missingIds.push(q.id)
    }
  }

  if (missingIds.length > 0) {
    const map: Record<number, boolean> = {}
    missingIds.forEach(id => {
      map[id] = true
    })
    highlightedQuestions.value = map
    showAlert('请先完成本页所有题目')
    return
  }

  // 当前页全部作答，清除高亮
  highlightedQuestions.value = {}

  // 2. 不是最后一页：翻到下一页
  if (currentPage.value < totalPages.value) {
    currentPage.value += 1
    return
  }

  // 3. 最后一页：暂时只做前端提示，后续再接入后端提交逻辑
  isSubmitting.value = true
  try {
    console.log('[QuestionsStageView] 当前阶段答题结果:', answers.value)
    showAlert('本阶段所有题目已完成（提交逻辑待接入）')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped src="@/styles/questions-stage.css"></style>
