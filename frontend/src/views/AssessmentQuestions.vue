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
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/views/components/StepIndicator.vue'
import {useQuestionsStagePage} from '@/controller/AssessmentQuestions'

const {
  route,
  loading,
  stepItems,
  currentStep,
  currentStepTitle,

  totalPages,
  totalCount,
  pageStartIndex,
  pageEndIndex,
  pagedQuestions,
  scaleOptions,
  answers,
  isFirstPage,
  isLastPage,
  isSubmitting,
  errorMessage,

  truncatedLatestMessage,
  isQuestionHighlighted,
  handlePrev,
  handleNext,
} = useQuestionsStagePage()
</script>

<style scoped src="@/styles/questions-stage.css"></style>
