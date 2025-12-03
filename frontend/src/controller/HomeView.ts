import {API_PATHS, apiRequest} from "@/api";
import {onBeforeUnmount, onMounted, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {TestRecordDTO, useTestSession} from '@/controller/testSession'
import {useAlert} from '@/controller/useAlert'
import {useAuthStore} from '@/controller/wx_auth'
import {
    loadProducts,
    PlanKey,
    pushStageRoute,
    StageBasic,
    type TestFlowStep,
} from "@/controller/common";
import {useGlobalLoading} from "@/controller/useGlobalLoading";

export interface FetchTestFlowResponse {
    record: TestRecordDTO
    steps: TestFlowStep[]
    current_stage: string
    current_index: number
}

export function useHomeView() {
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
    const {setTestFlow, setNextRouteItem, resetSession, setRecord} = useTestSession()
    const authStore = useAuthStore()
    const {showLoading, hideLoading} = useGlobalLoading()
    const isUserMenuOpen = ref(false)
    const userMenuWrapperRef = ref<HTMLElement | null>(null)
    const activeTab = ref<TabKey>('start')
    const scrollY = ref(0)

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
        loadProducts().then()
    })

    function handleFlowError(msg: string) {
        console.error('[HomeView] flow error:', msg)
        showAlert(msg)
    }

    function openLogin() {
        authStore.startWeChatLogin().then()
    }

    async function startTest(typ: PlanKey) {
        if (!authStore.isLoggedIn) {
            openLogin()
            return
        }

        showLoading("进入测试环节")

        try {
            const resp = await apiRequest<FetchTestFlowResponse>(API_PATHS.TEST_FLOW, {
                method: 'POST',
                body: {business_type: typ},
            });

            const steps = resp.steps || []
            if (!steps.length) {
                handleFlowError('测试流程异常，未配置任何测试阶段')
                return
            }
            const currentStage = resp.current_stage || StageBasic
            const currentIndex = resp.current_index

            setRecord(resp.record)

            setTestFlow(steps)
            setNextRouteItem(currentStage, currentIndex)
            await pushStageRoute(router, typ, currentStage)
        } catch (e) {
            showAlert("创建问卷测试失败:" + e)
        } finally {
            hideLoading()
        }
    }

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

    function handleUserClick(event?: MouseEvent) {
        // 防止点击头像时触发 document 的点击监听，导致立刻关闭
        if (event) {
            event.stopPropagation()
        }
        isUserMenuOpen.value = !isUserMenuOpen.value
    }

    // “我的测试”
    async function handleGoMyTests() {
        isUserMenuOpen.value = false
        await router.push({name: 'my-tests'})
    }

    // “退出登录”
    function handleLogout() {
        isUserMenuOpen.value = false
        authStore.logout().then(() => {
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

    onBeforeUnmount(() => {
        window.removeEventListener('scroll', handleScroll)
        document.removeEventListener('click', handleGlobalClick)
    })


    return {
        // 状态
        activePlan,
        activeTab,
        scrollY,
        tabDefs,
        // 行为
        openLogin,
        startTest,
        handleTabClick,
        handleUserClick,
        isUserMenuOpen,
        userMenuWrapperRef,
        handleGoMyTests,
        handleLogout,
    }
}