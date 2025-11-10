<!-- src/features/basic-info/view/TestBasicInfoView.vue -->
<template>
  <TestLayout>
    <template #header>
      <StepIndicator :steps="stepItems" :current="1" />
    </template>

    <!-- 提交后跳第二步前的过渡遮罩 -->
    <div v-if="navigating" class="overlay">
      <div class="overlay__card">正在从服务器获取信息…</div>
    </div>

    <form class="basic-form" @submit.prevent="handleSubmit">
      <h1 class="basic-form__title">{{ stepItems[0]?.title ?? '基础资料' }}</h1>

      <div class="basic-form__field">
        <label :for="ids.age">年龄</label>
        <input
            :id="ids.age"
            type="number"
            min="10"
            max="99"
            inputmode="numeric"
            v-model.number="form.age"
            @blur="touched.age = true"
        />
        <p v-if="ageError" class="basic-form__error">{{ ageError }}</p>
      </div>

      <div class="basic-form__field">
        <label :for="ids.mode">选科模式</label>
        <select :id="ids.mode" v-model="form.mode" @blur="touched.mode = true">
          <option value="" disabled>请选择</option>
          <option value="3+3">3+3</option>
          <option value="3+1+2">3+1+2</option>
        </select>
        <p v-if="modeError" class="basic-form__error">{{ modeError }}</p>
      </div>

      <div class="basic-form__field">
        <label :for="ids.hobby">兴趣方向</label>
        <select :id="ids.hobby" v-model="form.hobby" @blur="touched.hobby = true">
          <option value="" disabled>请选择</option>
          <option v-for="h in hobbies" :key="h" :value="h">{{ h }}</option>
        </select>
        <p v-if="hobbyError" class="basic-form__error">{{ hobbyError }}</p>
      </div>

      <button class="basic-form__submit" type="submit" :disabled="submitting || !isFormValid">
        <span v-if="submitting">处理中…</span>
        <span v-else>开始测试</span>
      </button>
    </form>

    <template #footer>
      <p class="text-xs text-gray-500">仅用于生成测评题目，不会保存隐私信息。</p>
    </template>
  </TestLayout>
</template>

<script setup lang="ts">
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/components/StepIndicator.vue'
import '@/features/basic-info/style/basic-info.css'
import { useBasicInfo } from '@/features/basic-info/logic/useBasicInfo'
const {
  form, touched, ids, hobbies,
  ageError, modeError, hobbyError, isFormValid,
  submitting, navigating, stepItems,
  handleSubmit,
} = useBasicInfo()
</script>
