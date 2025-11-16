import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useTestSession } from '@/store/testSession'

export function useQuestionsStageView() {
    const route = useRoute()
    const { state } = useTestSession()

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
        const scaleKey = String(route.params.scale ?? '')
        const idx = routes.findIndex(r => r.router === scaleKey)
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

    return {
        route,
        loading,
        stepItems,
        currentStep,
        currentStepTitle,
        showLoading,
        hideLoading,
    }
}


export interface ApplyTestRequest {
    test_type: string            // basic / pro ...
    invite_code?: string         // 邀请码
    wechat_openid?: string       // 微信 openid，可选
    grade?: string               // 年级
    mode?: string                // 3+3 / 3+1+2
    hobby?: string               // 兴趣爱好
    session_id?: string          // 当前 sessionID（如果有）
}

export interface ApplyTestResponse {
    test_id: number              // 对应 tests_record.id (bigserial)
    status: number               // 对应 tests_record.status (int2)，例如 0=生成中
}

export async function applyTest(
    routeKey: string,
    payload: ApplyTestRequest,
): Promise<ApplyTestResponse> {
    const resp = await fetch(`/api/apply_test/${routeKey}`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(payload),
    })

    if (!resp.ok) {
        throw new Error(`apply_test failed: ${resp.status}`)
    }
    return await resp.json()
}