// 单个测试步骤
import {apiRequest} from "@/api";
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useTestSession } from '@/store/testSession'
import { useAlert } from '@/logic/useAlert'
import {VerifyInviteResponse} from "@/controller/InviteCode";
import { useAuthStore } from '@/store/auth'
import {TestTypeBasic, TestTypePro, TestTypeSchool} from "@/controller/common";

export interface TestRouteDef {
    router: string; // 英文路由名，例如 'basic-info' | 'riasec' | 'asc' | 'report'
    desc: string;   // 中文描述，例如 '基本信息' | '兴趣测试' | '能力测试' | '测试报告'
}

export interface FetchTestFlowRequest {
    test_type: string;
    public_id?: string | null;
    invite_code?: string;
    wechat_openid?: string;
}

export interface FetchTestFlowResponse {
    public_id: string;
    routes: TestRouteDef[];
    nextRoute?: string | null;
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
        { key: 'start', label: '开始测试', targetId: 'section-start-test' },
        { key: 'intro', label: '产品介绍', targetId: 'section-product-intro' },
        { key: 'letter', label: '致家长的一封信', targetId: 'section-parent-letter' },
    ] as const

    const activePlan = ref<PlanKey>('basic')
    type TabKey = (typeof tabDefs)[number]['key']

    const { showAlert } = useAlert()
    const router = useRouter()
    const { state, setPublicID, setTestType, setTestRoutes } = useTestSession()
    const authStore = useAuthStore()
    const inviteModalOpen = ref(false)

    function openLogin() {
        authStore.openWeChatLogin()
        console.log('[HomeView] dialogOpen ->', authStore.wechatLoginOpen)
    }

    function startTest(typ: string) {
        setTestType(typ)
        inviteModalOpen.value = true
    }

    const activeTab = ref<TabKey>('start')
    const scrollY = ref(0)

    function handleTabClick(tab: typeof tabDefs[number]) {
        activeTab.value = tab.key

        const el = document.getElementById(tab.targetId)
        if (el) {
            el.scrollIntoView({ behavior: 'smooth', block: 'start' })
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
        const typ = state.testType || TestTypeBasic
        const req = {
            test_type: typ,
            public_id:payload.public_id,
            invite_code: state.inviteCode as string | undefined,
            wechat_openid: state.wechatOpenId as string | undefined,
        }

        let routes: TestRouteDef[] = []
        let nextRoute: string | null | undefined

        try {
            const resp = await fetchTestFlow(req)
            routes = resp.routes || []
            nextRoute = resp.nextRoute ?? null

            setPublicID(resp.public_id)
            setTestRoutes(routes)
        } catch (e) {
            console.error('[handleInviteSuccess] fetchTestFlow failed', e)
            showAlert('获取测试流程失败，请稍后再试:' + e)
            return
        }

        const targetRouter =   nextRoute || (routes.length > 0 ? routes[0].router : null)

        if (!targetRouter) {
            console.warn('[handleInviteSuccess] no target router found')
            showAlert('测试流程配置异常，请稍后再试或联系管理员')
            return
        }

        await router.push(`/assessment/${typ}/${targetRouter}`)
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