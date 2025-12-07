import {onBeforeUnmount, onMounted, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {TestRecordDTO, useTestSession} from '@/controller/testSession'
import {useAuthStore} from '@/controller/wx_auth'
import {
    loadProducts,
    PlanKey,
    type TestFlowStep,
} from "@/controller/common";
import {useGlobalLoading} from "@/controller/useGlobalLoading";
import {useTestLauncher} from "@/controller/useTestLauncher";

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
    const router = useRouter()
    const authStore = useAuthStore()
    const {showLoading, hideLoading} = useGlobalLoading()
    const activeTab = ref<TabKey>('start')
    const scrollY = ref(0)
    const {launchTest} = useTestLauncher()
    const showDisclaimer = ref(false)
    const pendingPlanKey = ref<PlanKey | null>(null)

    onMounted(() => {
        window.addEventListener('scroll', handleScroll)
        authStore.fetchSignInStatus().then().catch(err => {
            console.error('[HomeView] fetchSignInStatus failed', err)
        })
        loadProducts().then()
    })

    function openLogin() {
        authStore.startWeChatLogin().then()
    }

    async function startTest(typ: PlanKey) {
        if (!authStore.isLoggedIn) {
            openLogin()
            return
        }
        pendingPlanKey.value = typ
        showDisclaimer.value = true
    }

    function handleDisclaimerCancel() {
        showDisclaimer.value = false
        pendingPlanKey.value = null
    }

    async function handleDisclaimerConfirm() {
        if (!pendingPlanKey.value) {
            showDisclaimer.value = false
            return
        }
        const typ = pendingPlanKey.value

        showDisclaimer.value = false
        pendingPlanKey.value = null

        await launchTest({
            businessType: typ,
            loadingText: '进入测试环节',
        })
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
    }

    async function handleGoMyTests() {
        showLoading()
        router.push({name: 'my-tests'}).finally(() => {
            hideLoading()
        })
    }

    onBeforeUnmount(() => {
        window.removeEventListener('scroll', handleScroll)
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
        handleGoMyTests,
        showDisclaimer,
        handleDisclaimerCancel,
        handleDisclaimerConfirm,
    }
}