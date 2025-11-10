<template>
  <TestLayout>
    <template #header>
      <StepIndicator :steps="stepItems" :current="state.currentStep" />
    </template>

    <form class="basic-form" @submit.prevent="handleSubmit">
      <h1 class="basic-form__title">{{ stepItems[0]?.title ?? '' }}</h1>

      <div class="basic-form__field">
        <label :for="ids.age">{{ t('form.age.label') }}</label>
        <input
          :id="ids.age"
          type="number"
          min="10"
          max="99"
          inputmode="numeric"
          :placeholder="t('form.age.placeholder')"
          v-model.number="form.age"
          @blur="touched.age = true"
        />
        <p v-if="ageError" class="basic-form__error">{{ ageError }}</p>
      </div>

      <div class="basic-form__field">
        <label :for="ids.mode">{{ t('form.mode.label') }}</label>
        <select
          :id="ids.mode"
          v-model="form.mode"
          @blur="touched.mode = true"
        >
          <option value="" disabled>{{ t('form.mode.placeholder') }}</option>
          <option value="3+3">3+3</option>
          <option value="3+1+2">3+1+2</option>
        </select>
        <p v-if="modeError" class="basic-form__error">{{ modeError }}</p>
      </div>

      <div class="basic-form__field">
        <label :for="ids.hobby">{{ t('form.hobby.label') }}</label>
        <select
          :id="ids.hobby"
          v-model="form.hobby"
          @blur="touched.hobby = true"
        >
          <option value="" disabled>{{ t('form.hobby.placeholder') }}</option>
          <option v-for="item in hobbyOptions" :key="item" :value="item">{{ item }}</option>
        </select>
        <p v-if="hobbyError" class="basic-form__error">{{ hobbyError }}</p>
      </div>

      <div class="basic-form__actions">
        <button type="submit" class="basic-form__submit" :disabled="!isFormValid">{{ t('btn.start') }}</button>
      </div>
    </form>

    <template #footer>
      <p>{{ t('disclaimer') }}</p>
    </template>
  </TestLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/components/StepIndicator.vue'
import { STEPS, isVariant, type Variant } from '@/config/testSteps'
import { useI18n } from '@/i18n'
import { getHobbies } from '@/api'
import { useTestSession } from '@/store/testSession'

interface BasicFormState {
  age: number | null
  mode: '' | '3+3' | '3+1+2'
  hobby: string
}

const ids = {
  age: 'basic-age',
  mode: 'basic-mode',
  hobby: 'basic-hobby',
}

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { state, setVariant, setBasicInfo, setCurrentStep, ensureSessionId, nextStep } = useTestSession()

const form = reactive<BasicFormState>({
  age: state.age ?? null,
  mode: state.mode ?? '',
  hobby: state.hobby ?? '',
})

const touched = reactive({
  age: false,
  mode: false,
  hobby: false,
})

const hobbyOptions = ref<string[]>([])
const submitting = ref(false)

const variant = ref<Variant>('basic')

watchEffect(() => {
  const variantParam = String(route.params.variant ?? 'basic')
  if (!isVariant(variantParam)) {
    router.replace({ path: '/test/basic/step/1' })
    return
  }
  variant.value = variantParam
  setVariant(variant.value)
  setCurrentStep(1)
})

watchEffect(() => {
  const step = Number(route.params.step ?? '1')
  if (step !== 1) {
    router.replace({ path: `/test/${variant.value}/step/1` })
  }
})

const stepItems = computed(() =>
  STEPS[variant.value].map((item) => ({ key: item.key, title: t(item.titleKey) }))
)

const isAgeValid = computed(() => form.age !== null && !Number.isNaN(form.age) && form.age >= 10 && form.age <= 99)
const isModeValid = computed(() => Boolean(form.mode))
const isHobbyValid = computed(() => Boolean(form.hobby))

const ageError = computed(() => {
  if (!touched.age && !submitting.value) return ''
  if (form.age === null || Number.isNaN(form.age)) return t('form.validation.required')
  if (!isAgeValid.value) return t('form.validation.ageRange')
  return ''
})

const modeError = computed(() => {
  if (!touched.mode && !submitting.value) return ''
  if (!isModeValid.value) return t('form.validation.required')
  return ''
})

const hobbyError = computed(() => {
  if (!touched.hobby && !submitting.value) return ''
  if (!isHobbyValid.value) return t('form.validation.required')
  return ''
})

const isFormValid = computed(() => isAgeValid.value && isModeValid.value && isHobbyValid.value)

async function loadHobbies() {
  try {
    hobbyOptions.value = await getHobbies()
  } catch (error) {
    console.warn('[TestBasicInfoView] getHobbies failed', error)
    hobbyOptions.value = []
  }
}

onMounted(() => {
  loadHobbies()
  ensureSessionId()
})

function syncTouched() {
  touched.age = true
  touched.mode = true
  touched.hobby = true
}

async function handleSubmit() {
  submitting.value = true
  syncTouched()
  if (!isFormValid.value || form.age === null) {
    submitting.value = false
    return
  }

  setBasicInfo({ age: form.age, mode: form.mode, hobby: form.hobby })
  setCurrentStep(1)
  const limit = STEPS[variant.value].length
  const next = nextStep(limit)
  await router.push({ path: `/test/${variant.value}/step/${next}` })
  submitting.value = false
}
</script>

<style scoped>
.basic-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.basic-form__title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
}

.basic-form__field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.basic-form__field label {
  font-weight: 500;
  color: #111827;
}

.basic-form__field input,
.basic-form__field select {
  padding: 12px 14px;
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.6);
  font-size: 16px;
  color: #111827;
  background-color: #f9fafb;
}

.basic-form__field input:focus,
.basic-form__field select:focus {
  outline: none;
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
  background-color: #ffffff;
}

.basic-form__error {
  color: #dc2626;
  font-size: 13px;
  margin: 0;
}

.basic-form__actions {
  display: flex;
  justify-content: flex-end;
}

.basic-form__submit {
  min-width: 160px;
  padding: 12px 20px;
  border-radius: 999px;
  border: none;
  cursor: pointer;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  font-size: 16px;
  font-weight: 600;
  transition: opacity 0.2s ease;
}

.basic-form__submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .basic-form__title {
    font-size: 20px;
  }

  .basic-form__submit {
    width: 100%;
  }
}
</style>
