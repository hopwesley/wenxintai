// 单个测试步骤
import {apiRequest} from "@/api";
import {ref, onMounted, onBeforeUnmount} from 'vue'
import {useRouter} from 'vue-router'
import {useTestSession} from '@/store/testSession'
import {useAlert} from '@/controller/useAlert'
import {VerifyInviteResponse} from "@/controller/InviteCode";
import {useAuthStore} from '@/store/auth'
import {StageBasic, TestTypeBasic, TestTypePro, TestTypeSchool} from "@/controller/common";


export interface FetchTestFlowRequest {
    public_id: string;
}

export interface FetchTestFlowResponse {
    public_id: string;
    routes: TestRouteDef[];
    nextRoute?: string | null;
    next_route_id?: number
}

export async function fetchTestFlow(payload: FetchTestFlowRequest) {
    return apiRequest<FetchTestFlowResponse>('/api/test_flow', {
        method: 'POST',
        body: payload,
    });
}

export function useHomeView() {

    type PlanKey = typeof TestTypeBasic | typeof TestTypePro | typeof TestTypeSchool

    const tabDefs = [
        {key: 'start', label: '开始测试', targetId: 'section-start-test'},
        {key: 'intro', label: '产品介绍', targetId: 'section-product-intro'},
        {key: 'letter', label: '致家长的一封信', targetId: 'section-parent-letter'},
    ] as const

    const activePlan = ref<PlanKey>('basic')
    type TabKey = (typeof tabDefs)[number]['key']

    const {showAlert} = useAlert()
    const router = useRouter()
    const {state, setPublicID, setBusinessType, setTestRoutes, setNextRouteItem} = useTestSession()
    const authStore = useAuthStore()
    const inviteModalOpen = ref(false)

    function openLogin() {
        authStore.openWeChatLogin()
        console.log('[HomeView] dialogOpen ->', authStore.wechatLoginOpen)
    }

    function startTest(typ: string) {
        setBusinessType(typ)
        inviteModalOpen.value = true
    }

    const activeTab = ref<TabKey>('start')
    const scrollY = ref(0)

    function handleTabClick(tab: typeof tabDefs[number]) {
        activeTab.value = tab.key

        const el = document.getElementById(tab.targetId)
        if (el) {
            el.scrollIntoView({behavior: 'smooth', block: 'start'})
        }
    }

    function handleScroll() {
        scrollY.value = window.scrollY || window.pageYOffset || 0
    }

    onMounted(() => {
        window.addEventListener('scroll', handleScroll)
    })

    onBeforeUnmount(() => {
        window.removeEventListener('scroll', handleScroll)
    })

    async function handleInviteSuccess(payload: VerifyInviteResponse) {
        const typ = state.businessType || TestTypeBasic
        const req = {
            public_id: payload.public_id,
        }

        let routes: string[] = []
        let nextRoute: string
        let nextRouteId: number
        try {
            const resp = await fetchTestFlow(req)
            routes = resp.routes || []
            if (routes.length == 0) {
                showAlert('没有可用的测试流程，请稍后再试或联系管理员')
                return
            }

            nextRoute = resp.next_route ?? StageBasic
            nextRouteId = resp.next_route_id ?? 0
            if (nextRouteId < 0) nextRouteId = 0

            setPublicID(resp.public_id)
            setTestRoutes(routes)
            setNextRouteItem(nextRoute, nextRouteId)
        } catch (e) {
            console.error('[handleInviteSuccess] fetchTestFlow failed', e)
            showAlert('获取测试流程失败，请稍后再试:' + e)
            return
        }

        if (routes.length === 0) {
            console.warn('[handleInviteSuccess] no target router found')
            showAlert('测试流程配置异常，请稍后再试或联系管理员')
            return
        }

        await router.push(`/assessment/${typ}/${nextRoute}`)
    }

    return {
        // 状态
        activePlan,
        inviteModalOpen,
        activeTab,
        scrollY,
        tabDefs,
        // 行为
        openLogin,
        startTest,
        handleTabClick,
        handleInviteSuccess,
    }
}