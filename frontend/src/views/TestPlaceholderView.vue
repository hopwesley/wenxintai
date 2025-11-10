<template>
  <TestLayout>
    <template #header>
      <StepIndicator :steps="stepItems" :current="state.currentStep"/>
    </template>

    <section class="placeholder">
      <h1>{{ currentStepTitle }}</h1>
      <p>placeholder.description</p>
      <button type="button" class="placeholder__action" @click="goHome">placeholder.back</button>
    </section>

    <template #footer>
      <p>disclaimer</p>
    </template>
  </TestLayout>
</template>

<script setup lang="ts">
import {computed, ref, watchEffect} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/components/StepIndicator.vue'
import {useTestSession} from '@/store/testSession'
import {STEPS, isVariant, type Variant} from '@/config/testSteps'

const route = useRoute()
const router = useRouter()
const {state, setVariant, setCurrentStep} = useTestSession()
const variant = ref<Variant>('basic')

watchEffect(() => {
  const variantParam = String(route.params.variant ?? 'basic')
  if (!isVariant(variantParam)) {
    router.replace({path: '/test/basic/step/1'})
    return
  }
  variant.value = variantParam
  setVariant(variant.value)
})

watchEffect(() => {
  const stepNumber = Number(route.params.step ?? '1')
  if (Number.isNaN(stepNumber) || stepNumber < 1 || stepNumber > STEPS[variant.value].length) {
    router.replace({path: `/test/${variant.value}/step/1`})
    return
  }
  setCurrentStep(stepNumber)
})

const stepItems = computed(() =>
    STEPS[variant.value].map((item) => ({key: item.key, title: item.titleKey}))
)

const currentStepTitle = computed(() => {
  const index = state.currentStep - 1
  const item = STEPS[variant.value][index]
  return item ? item.titleKey : ('placeholder.title')
})

function goHome() {
  router.push('/')
}
</script>

<style scoped>
.placeholder {
  min-height: 320px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  text-align: center;
}

.placeholder h1 {
  margin: 0;
  font-size: 24px;
  color: #1f2937;
}

.placeholder p {
  margin: 0;
  font-size: 16px;
  color: rgba(30, 41, 59, 0.75);
}

.placeholder__action {
  margin-top: 16px;
  padding: 10px 24px;
  border-radius: 999px;
  border: none;
  background: rgba(148, 163, 184, 0.2);
  color: #1f2937;
  cursor: pointer;
  font-weight: 600;
}
</style>
