// 单个测试步骤
import {apiRequest} from "@/api";
import {ref, onMounted, onBeforeUnmount, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useTestSession} from '@/controller/testSession'
import {useAlert} from '@/controller/useAlert'
import {VerifyInviteResponse} from "@/controller/InviteCode";
import {useAuthStore} from '@/controller/wx_auth'
import {
    StageBasic,
    TestTypeBasic,
    type TestFlowStep,
    pushStageRoute, PlanKey, TestTypePro, TestTypeAdv, TestTypeSchool,
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
    const route = useRoute()
    const {state, setPublicID, setBusinessType, setTestFlow, setNextRouteItem, resetSession} = useTestSession()

    const authStore = useAuthStore()
    const inviteModalOpen = ref(false)

    function openLogin() {
        authStore.startWeChatLogin().then()
    }

    const currentPlan = ref<PlanInfo | null>(null)

    function startTest(typ: PlanKey) {
        setBusinessType(typ)
        const plan = planMap[typ]
        if (plan) {
            currentPlan.value = plan
        } else {
            currentPlan.value = null
        }
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

        if (isUserMenuOpen.value) {
            isUserMenuOpen.value = false
        }
    }

    const isUserMenuOpen = ref(false)
    const userMenuWrapperRef = ref<HTMLElement | null>(null)

    function handleUserClick(event?: MouseEvent) {
        // 防止点击头像时触发 document 的点击监听，导致立刻关闭
        if (event) {
            event.stopPropagation()
        }
        isUserMenuOpen.value = !isUserMenuOpen.value
    }

    // “我的测试”
    function handleGoMyTests() {
        isUserMenuOpen.value = false
        router.push({ name: 'my-tests' })
    }

    // “退出登录”
    function handleLogout() {
        isUserMenuOpen.value = false
        authStore.logout().then(()=>{
            resetSession();
        })
        console.log('[HomeView] logout clicked')
    }

    function handleGlobalClick(e: MouseEvent) {
        if (!isUserMenuOpen.value) return

        const rootEl = userMenuWrapperRef.value
        if (!rootEl) return

        const target = e.target as Node | null
        if (target && rootEl.contains(target)) {
            // 点击在头像/菜单区域内，不关闭
            return
        }
        // 点击在外面，关闭菜单
        isUserMenuOpen.value = false
    }

    watch(
        () => route.fullPath,
        () => {
            if (isUserMenuOpen.value) {
                isUserMenuOpen.value = false
            }
        }
    )

    onMounted(() => {
        window.addEventListener('scroll', handleScroll)
        document.addEventListener('click', handleGlobalClick)
        authStore.fetchSignInStatus().then().catch(err => {
            console.error('[HomeView] fetchSignInStatus failed', err)
        })
    })

    onBeforeUnmount(() => {
        window.removeEventListener('scroll', handleScroll)
        document.removeEventListener('click', handleGlobalClick)
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
        } finally {
            hideLoading()
        }
    }

    async function handleWeChatPay() {
        try {
            // 1. 调用后端创建订单，拿到微信支付参数
            // const res = await apiRequest('/api/pay/wechat/create-order', { ... })

            // 2. 在微信内环境调起支付（WeixinJSBridge 或 JSSDK）
            // await callWeChatPay(res.data)

            // 3. 支付成功后：关闭弹窗 + 进入测试（跟邀请码成功后的流程类似）
            inviteModalOpen.value = false
            // 这里可以直接 push 到测试页，或者调用你已有的 handlePaySuccess 之类
        } catch (e) {
            // 支付失败 / 取消：你可以选择：
            // - 提示错误，但不关弹窗
            // - 或者关闭弹窗，让用户重新点击支付
            console.error(e)
            // 如果你希望允许重试，可以让弹窗重新打开一次
            // inviteModalOpen.value = true
        }
    }


    interface PlanInfo {
        key: PlanKey
        name: string
        price: number       // 单位元；如果你用分自己改成 number of cents
        desc: string
        tag?: string        // 如果某些卡片有“推荐”“热门”之类的小标签可以放这里
    }

// 1) 把你当前 <section> 里写死的 4 个产品信息拷到这里来：名称、价格、简介
    const basicPlan: PlanInfo = {
        key: TestTypeBasic,
        name: '基础版',
        price: 29.9,
        desc: '组合推荐 + 学科优势评估',
    }

    const proPlan: PlanInfo = {
        key: TestTypePro,
        name: '专业版',
        price: 49.9,
        desc: '基础版+更加全面的参数解读',
        tag: '推荐',
    }

    const advPlan: PlanInfo = {
        key: TestTypeAdv,
        name: '增强版',
        price: 79.9,
        desc: '专业版 +专业选择推荐+职业规划建议',
    }

    const schoolPlan: PlanInfo = {
        key: TestTypeSchool,
        name: '校本定制版',
        price: 59.9,
        desc: '结合校园真是数据，精准报告，多维对比',
    }

    const planMap: Record<PlanKey, PlanInfo> = {
        [TestTypeBasic]: basicPlan,
        [TestTypePro]: proPlan,
        [TestTypeAdv]: advPlan,
        [TestTypeSchool]: schoolPlan,
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
        isUserMenuOpen,
        userMenuWrapperRef,
        handleGoMyTests,
        handleLogout,
        handleWeChatPay,
        planMap,
        currentPlan,
    }
}