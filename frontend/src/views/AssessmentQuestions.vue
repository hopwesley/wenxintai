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
      </footer>
    </section>

    <!-- 全屏遮罩 -->
    <div v-if="loading" class="overlay">
      <div class="overlay__card">正在从服务器获取信息…</div>
      <p v-if="latestMessage">{{ latestMessage }}</p>
    </div>
  </TestLayout>
</template>
<script setup lang="ts">
import {onMounted, ref} from 'vue'
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/views/components/StepIndicator.vue'
import {useQuestionsStageView} from '@/controller/AssessmentQuestions'
import {useTestSession} from "@/store/testSession";
import {useSubscriptBySSE} from "@/controller/common";
import {useAlert} from "@/controller/useAlert";
import {router} from "@/router";

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

const totalPages = ref(1)
const currentPage = ref(1)
const currentPageQuestions = ref<{ id: string; text: string }[]>([])
const errorMessage = ref('')
const highlightedId = ref<string | null>(null)
const scaleOptions = ref<{ value: number; label: string }[]>([])
let latestMessage = ref("")

const {state, getPublicID} = useTestSession()
const public_id: string | undefined = getPublicID()
const {showAlert} = useAlert()
const routes = state.testRoutes ?? []
const {businessType, testStage} = route.params as { businessType: string; testStage: string }

console.log('[QuestionsStageView] apply_test resp:', testStage, businessType, public_id, routes)

function getAnswer(_id: string) {
  return undefined
}

function onSelect(_id: string, _value: number) {
}

onMounted(() => {
  errorMessage.value = ''

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

  let message="";
  const sseCtrl = useSubscriptBySSE(public_id, businessType, testStage, {
    autoStart: false,
    onOpen() {
      showLoading()
    },

    onError(err) {
      console.log("------>>> sse channel error:", err)
      showAlert('获取测试流程失败，请稍后再试:' + err)
      hideLoading()
    },

    onMsg(msg) {
      message+=msg;
      latestMessage.value = message;
    },

    onClose() {
      console.log("------>>> sse closed:")
      hideLoading()
    },

    onDone(questionStr) {
      console.log("------>>> go questions:", questionStr)
      hideLoading()
    }
  })

  sseCtrl.start()
})

async function handleNext() {
}

</script>

<style scoped src="@/styles/questions-stage.css"></style>
