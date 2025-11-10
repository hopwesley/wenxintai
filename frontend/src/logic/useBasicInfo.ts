import { computed, onMounted, reactive, ref, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTestSession } from '@/store/testSession'
import { STEPS, type Variant, isVariant } from '@/config/testSteps'
import { getHobbies } from '@/api'

// 与后端/Store 的真实定义保持一致：仅本文件内使用，避免编译冲突
type ModeOption = '3+3' | '3+1+2'

type StepLite = { key: string; titleKey?: string }

interface BasicFormState {
    age: number | null
    mode: '' | ModeOption
    hobby: string
}

export function useBasicInfo() {
    const route = useRoute()
    const router = useRouter()

    const {
        setVariant, setCurrentStep, setBasicInfo, nextStep,
    } = useTestSession()

    const variant = ref<Variant>('basic')
    const submitting = ref(false)
    const navigating = ref(false)
    const hobbies = ref<string[]>([])

    const form = reactive<BasicFormState>({
        age: null,
        mode: '',
        hobby: '',
    })
    const touched = reactive({ age: false, mode: false, hobby: false })

    const ids = { age: 'age-input', mode: 'mode-select', hobby: 'hobby-select' }

    // 首次拉兴趣列表失败时降级为空数组
    onMounted(async () => {
        try {
            const list = await getHobbies()
            hobbies.value = Array.isArray(list) ? list.map(String) : []
        } catch {
            hobbies.value = []
        }
    })

    // 路由中的 variant/step 约束到类型
    watchEffect(() => {
        const v = String(route.params.variant ?? 'basic')
        if (isVariant(v)) {
            variant.value = v
            setVariant(v)
        } else {
            router.replace({ path: '/test/basic/step/1' })
            return
        }

        // 基础资料页固定 step=1
        const stepNum = Number(route.params.step ?? '1')
        if (stepNum !== 1) {
            router.replace({ path: `/test/${variant.value}/step/1` })
            return
        }
        setCurrentStep(1)
    })

    // 明确收敛 STEPS 的元素类型，避免 never[]
    const stepItems = computed(() => {
        const arr = (STEPS as Record<Variant, readonly StepLite[]>)[variant.value] ?? []
        // 此处不依赖 i18n，先展示 titleKey 或 key，确保能编译运行
        return arr.map(it => ({ key: it.key, title: it.titleKey ?? it.key }))
    })

    // 基本校验
    const isAgeValid = computed(() => form.age !== null && !Number.isNaN(form.age) && form.age >= 10 && form.age <= 99)
    const isModeValid = computed(() => form.mode === '3+3' || form.mode === '3+1+2')
    const isHobbyValid = computed(() => !!form.hobby)
    const isFormValid = computed(() => isAgeValid.value && isModeValid.value && isHobbyValid.value)

    const ageError = computed(() => {
        if (!touched.age && !submitting.value) return ''
        if (form.age == null) return '必填'
        if (!isAgeValid.value) return '年龄范围 10-99'
        return ''
    })
    const modeError = computed(() => {
        if (!touched.mode && !submitting.value) return ''
        if (!isModeValid.value) return '必选'
        return ''
    })
    const hobbyError = computed(() => {
        if (!touched.hobby && !submitting.value) return ''
        if (!isHobbyValid.value) return '必选'
        return ''
    })

    function touchAll() {
        touched.age = true
        touched.mode = true
        touched.hobby = true
    }

    async function handleSubmit() {
        submitting.value = true
        touchAll()

        if (!isFormValid.value || form.age == null) {
            submitting.value = false
            return
        }

        // mode 去掉 ""，收口为 ModeOption
        const mode: ModeOption = form.mode === '3+3' ? '3+3' : '3+1+2'
        setBasicInfo({ age: form.age, mode, hobby: form.hobby })

        const limit = ((STEPS as Record<Variant, readonly StepLite[]>)[variant.value] ?? []).length
        const next = nextStep(limit)
        try {
            navigating.value = true
            await router.push({ path: `/test/${variant.value}/step/${next}` })
        } finally {
            navigating.value = false
            submitting.value = false
        }
    }

    return {
        // 表单与校验
        form, touched, ids, hobbies,
        ageError, modeError, hobbyError, isFormValid,
        // 进度 & 步骤
        submitting, navigating, stepItems,
        // 事件
        handleSubmit,
    }
}