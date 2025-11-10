<!-- src/features/questions-stage/view/QuestionsStageView.vue -->
<template>
  <TestLayout>
    <template #header>
      <StepIndicator :steps="stepItems" :current="currentStep" />
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
          <div v-for="q in currentPageQuestions" :key="q.id" class="question" :class="{ 'question--highlight': highlightedId === q.id }" ref="setRef">
            <p class="question__text">{{ q.text }}</p>
            <div class="question__options">
              <label v-for="opt in scaleOptions" :key="opt.value" class="question__option">
                <input type="radio" :name="q.id" :value="opt.value" :checked="getAnswer(q.id) === opt.value" @change="onSelect(q.id, opt.value)" />
                <span>{{ opt.label }}</span>
              </label>
            </div>
          </div>
        </form>
      </div>

      <footer class="questions__footer" v-if="!loading && !errorMessage">
        <button type="button" class="questions__nav questions__nav--prev" @click="handlePrev">上一步</button>
        <button type="button" class="questions__nav questions__nav--next" :disabled="!isCurrentPageComplete || submitting" @click="handleNext">
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
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/components/StepIndicator.vue'
import '@/styles/questions-stage.css'
import { useQuestionsStage } from '@/logic/useQuestionsStage'

const {
  // 状态
  loading, submitting, errorMessage,
  currentStep, currentStepTitle, stepItems,
  // 分页
  currentPage, totalPages, currentPageQuestions, isCurrentPageComplete, nextLabel,
  // 题目
  scaleOptions, getAnswer, onSelect, handlePrev, handleNext,
  // 高亮
  highlightedId, setRef,
} = useQuestionsStage({ stage: 1, pageSize: 5 }) // 这是“第二步（第一阶段）”；第三步传 { stage: 2 }
</script>
