// 单个测试步骤
import {apiRequest} from "@/api";
import {onBeforeUnmount, onMounted, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useTestSession} from '@/controller/testSession'
import {useAlert} from '@/controller/useAlert'
import {VerifyInviteResponse} from "@/controller/InviteCode";
import {useAuthStore} from '@/controller/wx_auth'
import {
    PlanKey,
    pushStageRoute,
    StageBasic,
    type TestFlowStep,
    TestTypeAdv,
    TestTypeBasic,
    TestTypePro,
    TestTypeSchool,
} from "@/controller/common";
import {useGlobalLoading} from "@/controller/useGlobalLoading";
import {createNativeOrder, NativeCreateOrderResponse, queryOrderStatus} from "@/controller/WeChatNativePay";

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
        if (!authStore.isLoggedIn) {
            openLogin()
            return
        }

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



    const paying = ref(false)              // 点击“微信支付”后，按钮 loading
    const payOrder = ref<NativeCreateOrderResponse | null>(null)  // 当前订单信息（包括 code_url）
    const payPollingTimer = ref<number | null>(null)              // 轮询定时器 id
    const paySucceeded = ref(false)
    async function handleWeChatPay() {
        if (!currentPlan.value) {
            showAlert('请选择测试方案')
            return
        }

        if (paying.value) {
            return
        }

        try {
            paying.value = true
            paySucceeded.value = false

            // 1. 调后端创建 native 订单
            const order = await createNativeOrder(currentPlan.value.key)

            payOrder.value = order

            // 2. 不要立刻关弹窗，而是在弹窗里展示二维码
            //   -> InviteCodeModal 需要根据是否有 payOrder 来切换 UI（输入邀请码 vs 显示二维码）
            //   -> 这里只负责把 payOrder 设置好

            // 3. 开始轮询订单状态
            startPayPolling(order.order_id)
        } catch (err: any) {
            console.error('[Pay] handleWeChatPay failed', err)
            showAlert(err?.message || '创建支付订单失败，请稍后再试')
        } finally {
            paying.value = false
        }
    }

    function startPayPolling(orderId: string) {
        stopPayPolling() // 防止重复

        // 每 2 秒查一次支付状态，视情况调节
        payPollingTimer.value = window.setInterval(async () => {
            try {
                const status = await queryOrderStatus(orderId)
                if (status.paid) {
                    paySucceeded.value = true
                    stopPayPolling()

                    // 支付成功：关闭弹窗 + 进入测试
                    inviteModalOpen.value = false

                    // TODO: 在这里触发你原来进入测试的逻辑
                    // 例如：router.push(...) 或者调用 handlePaySuccess()
                }
            } catch (err) {
                console.error('[Pay] queryOrderStatus error', err)
                // 轮询失败可以先不打断，下次继续查
            }
        }, 2000)
    }

    function stopPayPolling() {
        if (payPollingTimer.value !== null) {
            window.clearInterval(payPollingTimer.value)
            payPollingTimer.value = null
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
        payOrder,
        paying,
    }
}