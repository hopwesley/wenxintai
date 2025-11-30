// 单个测试步骤
import {apiRequest} from "@/api";
import {ref, onMounted, onBeforeUnmount} from 'vue'
import {useRouter} from 'vue-router'
import {useTestSession} from '@/controller/testSession'
import {useAlert} from '@/controller/useAlert'
import {VerifyInviteResponse} from "@/controller/InviteCode";
import {useAuthStore} from '@/controller/wx_auth'
import {
    StageBasic,
    TestTypeBasic,
    type TestFlowStep,
    pushStageRoute, PlanKey,
} from "@/controller/common";
import {useGlobalLoading} from "@/controller/useGlobalLoading";


export interface FetchTestFlowRequest {
    public_id: string;
}

export interface FetchTestFlowResponse {
    public_id: string
    business_type: PlanKey

    // 完整流程
    steps: TestFlowStep[]

    // 当前要进入的阶段（例如 "basic-info" / "riasec"）
    current_stage: string
    current_index: number
}


export async function fetchTestFlow(payload: FetchTestFlowRequest) {
    return apiRequest<FetchTestFlowResponse>('/api/test_flow', {
        method: 'POST',
        body: payload,
    });
}

export function useHomeView() {

    const inviteStatus = ref<'idle' | 'success' | 'error'>('idle')

    function handleFlowError(msg: string) {
        console.error('[HomeView] flow error:', msg)
        inviteStatus.value = 'error'
        showAlert(msg)
    }

    const tabDefs = [
        {key: 'start', label: '开始测试', targetId: 'section-start-test'},
        {key: 'intro', label: '产品介绍', targetId: 'section-product-intro'},
        {key: 'letter', label: '致家长的一封信', targetId: 'section-parent-letter'},
    ] as const

    const activePlan = ref<PlanKey>('basic')
    type TabKey = (typeof tabDefs)[number]['key']

    const {showAlert} = useAlert()
    const router = useRouter()
    const {state, setPublicID, setBusinessType, setTestFlow, setNextRouteItem} = useTestSession()

    const authStore = useAuthStore()
    const inviteModalOpen = ref(false)

    function openLogin() {
        authStore.startWeChatLogin().then()
    }

    function startTest(typ: PlanKey) {
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
        authStore.fetchSignInStatus().then().catch(err => {
            console.error('[HomeView] fetchSignInStatus failed', err)
        })
    })

    onBeforeUnmount(() => {
        window.removeEventListener('scroll', handleScroll)
    })
    const {showLoading, hideLoading} = useGlobalLoading()
    async function handleInviteSuccess(payload: VerifyInviteResponse) {
        // 当前业务类型（兜底用）
        const fallbackBusinessType = state.businessType || TestTypeBasic

        // 先把 public_id 存一下
        setPublicID(payload.public_id)

        const req = {public_id: payload.public_id}

        showLoading("正在进入问卷测试环节......")

        try {
            const resp = await fetchTestFlow(req)

            const steps = resp.steps || []
            if (!steps.length) {
                handleFlowError('测试流程异常，未配置任何测试阶段')
                return
            }

            const businessType = resp.business_type || fallbackBusinessType
            if (!businessType) {
                handleFlowError('测试流程异常，未找到测评类型')
                return
            }

            const currentStage = resp.current_stage || StageBasic
            const currentIndex = resp.current_index

            // 同步到全局 store
            setBusinessType(businessType)
            setTestFlow(steps)
            setNextRouteItem(currentStage, currentIndex)

            await pushStageRoute(router, businessType, currentStage)
        } catch (err) {
            console.error('fetch test flow failed:', err)
            handleFlowError('获取测试流程失败，请稍后再试')
        }finally {
            hideLoading()
        }
    }

    function handleUserClick() {
        console.log('[HomeView] user avatar clicked')
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
        handleUserClick,
    }
}