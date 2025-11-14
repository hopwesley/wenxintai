// src/features/questions-stage/view/QuestionsStageView.ts
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useTestSession } from '@/store/testSession'

// StepIndicator + 标题 + loading 的统一逻辑
export function useQuestionsStageView() {
    const route = useRoute()
    const { state } = useTestSession()

    // 1) loading：控制顶部“正在从服务器获取信息…”和全屏遮罩
    const loading = ref(true)

    function showLoading() {
        loading.value = true
    }

    function hideLoading() {
        loading.value = false
    }

    // 2) StepIndicator 的 steps：直接用后端返回的 testRoutes
    //    Go 端结构：[{ router: "basic-info", desc: "基本信息" }, ...]
    const stepItems = computed(() => {
        const routes = state.testRoutes ?? []
        return routes.map(r => ({
            key: r.router,   // 英文路由 key，用来对齐当前 router
            title: r.desc,   // 中文标题，显示在 StepIndicator 上
        }))
    })

    // 3) 当前是第几步：用当前 URL 里的 :scale 去 testRoutes 里找下标
    //    例如 /test/basic/riasec  → 找到 router === 'riasec' 的那一项
    const currentStep = computed(() => {
        const routes = state.testRoutes ?? []
        const scaleKey = String(route.params.scale ?? '')
        const idx = routes.findIndex(r => r.router === scaleKey)
        // StepIndicator 用 1-based 索引（第一个步骤是 1）
        return idx >= 0 ? idx + 1 : 0
    })

    // 4) 标题：用当前 step 对应的 desc；找不到就用“正在加载…”
    const currentStepTitle = computed(() => {
        const routes = state.testRoutes ?? []
        const idx = currentStep.value - 1
        if (idx >= 0 && idx < routes.length) {
            return routes[idx].desc || '正在加载…'
        }
        return '正在加载…'
    })

    return {
        // 给 Vue 用的
        route,
        loading,
        stepItems,
        currentStep,
        currentStepTitle,
        showLoading,
        hideLoading,
    }
}
