import { computed, ref } from 'vue'
import { useRoute,useRouter } from 'vue-router'
import { useTestSession } from '@/store/testSession'
import {StageAsc, StageMotivation, StageOcean, StageRiasec} from "@/controller/common";
import {useAlert} from "@/controller/useAlert";

export function useQuestionsStageView() {
    const route = useRoute()
    const router = useRouter()
    const { state } = useTestSession()
    const {showAlert} = useAlert()

    const loading = ref(true)

    function showLoading() {
        loading.value = true
    }

    function hideLoading() {
        loading.value = false
    }

    const stepItems = computed(() => {
        const routes = state.testRoutes ?? []
        return routes.map(r => ({
            key: r.router,
            title: r.desc,
        }))
    })

    const currentStep = computed(() => {
        const routes = state.testRoutes ?? []
        const testStage = String(route.params.testStage ?? '')
        const idx = routes.findIndex(r => r.router === testStage)
        return idx >= 0 ? idx + 1 : 0
    })

    const currentStepTitle = computed(() => {
        const routes = state.testRoutes ?? []
        const idx = currentStep.value - 1
        if (idx >= 0 && idx < routes.length) {
            return routes[idx].desc || '正在加载…'
        }
        return '正在加载…'
    })

    function validateTestStage(testStage: string): boolean {

        const validStages = [
            StageRiasec,
            StageAsc,
            StageOcean,
            StageMotivation,
        ]

        if (!validStages.includes(testStage)) {
            showAlert( '测试流程异常，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return false
        }
        return true
    }

    return {
        validateTestStage,
        showAlert,
        route,
        loading,
        stepItems,
        currentStep,
        currentStepTitle,
        showLoading,
        hideLoading,
    }
}