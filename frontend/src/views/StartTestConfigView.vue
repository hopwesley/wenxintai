<template>
  <TestLayout>
    <template #header>
      <StepIndicator :steps="stepItems" :current="currentStepIndex"/>
    </template>
    <main class="config-page">
      <section class="config-card">
        <header class="config-header">
          <p class="config-badge">邀请码：{{ inviteCode }}</p>
          <h1>完善测试设置</h1>
          <p class="config-desc">请选择年级与测试模式，我们会基于您的选择生成专属题目。</p>
        </header>

        <form class="config-form" @submit.prevent="handleSubmit">
          <!-- 年级：下拉 -->
          <label class="config-field">
            <span>年级</span>
            <select v-model="form.grade" :disabled="submitting" required>
              <option value="">请选择年级</option>
              <option value="初二">初二</option>
              <option value="初三">初三</option>
              <option value="高一">高一</option>
            </select>
          </label>

          <!-- 模式：下拉（必选） -->
          <label class="config-field">
            <span>测试模式</span>
            <select v-model="form.mode" :disabled="submitting" required>
              <option value="">请选择模式</option>
              <option value="3+3">3+3 模式</option>
              <option value="3+1+2">3+1+2 模式</option>
            </select>
          </label>

          <!-- 兴趣：下拉（可选，来自后端） -->
          <label class="config-field">
            <span>兴趣偏好（可选）</span>
            <select v-model="form.hobby" :disabled="submitting">
              <option value="">请选择爱好</option>
              <option v-for="h in hobbies" :key="h" :value="h">{{ h }}</option>
            </select>
          </label>

          <p v-if="errorMessage" class="config-error">{{ errorMessage }}</p>

          <button class="config-submit" type="submit" :disabled="!canSubmit || submitting">
            <span v-if="submitting">创建中…</span>
            <span v-else>下一步</span>
          </button>
        </form>
      </section>
    </main>
  </TestLayout>
</template>

<script setup lang="ts">
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/views/components/StepIndicator.vue'
import { useStartTestConfig } from '@/controller/StartTestConfigView'

const {
  inviteCode,
  hobbies,
  form,
  submitting,
  errorMessage,
  canSubmit,
  stepItems,
  currentStepIndex,
  handleSubmit,
} = useStartTestConfig()
</script>

<style scoped src="@/styles/start-test-config.css"></style>

