import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiRequest } from '@/api'
import { useGlobalLoading } from '@/controller/useGlobalLoading'
import { useAlert } from '@/controller/useAlert'

export type MyTestStatus = 'RUNNING' | 'COMPLETED_NO_REPORT' | 'COMPLETED_WITH_REPORT'

export interface MyTestItem {
    public_id: string
    business_type: string
    mode: string
    created_at: string
    status: MyTestStatus
    completed_at?: string | null
    report_generated_at: string | null
}

/**
 * 对应 app.user_profile
 */
export interface UserProfile {
    id:number
    uid: string
    nick_name: string
    avatar_url: string
    mobile: string | null
    study_id: string | null
    school_name: string | null
    province: string | null
    created_at: string
    updated_at: string
    city: string | null
    report_no: number
}

/**
 * 新增的组合接口：一个用户档案 + 多个测试记录
 */
export interface MyTestsResponse {
    profile: UserProfile
    tests: MyTestItem[]
}

const businessTypeLabelMap: Record<string, string> = {
    basic: '基础能力测评',
    pro: '进阶能力测评',
    adv: '深度选科规划',
    school: '校园合作测评',
}

function formatDateTime(iso: string | null | undefined): string {
    if (!iso) return ''
    const date = new Date(iso)
    const y = date.getFullYear()
    const m = String(date.getMonth() + 1).padStart(2, '0')
    const d = String(date.getDate()).padStart(2, '0')
    const hh = String(date.getHours()).padStart(2, '0')
    const mm = String(date.getMinutes()).padStart(2, '0')
    return `${y}-${m}-${d} ${hh}:${mm}`
}

export function useWechatProfile() {
    const router = useRouter()
    const { showLoading, hideLoading } = useGlobalLoading()
    const { showAlert } = useAlert()

    const loading = ref(false)

    // 🔹 用户档案
    const profile = ref<UserProfile | null>(null)

    // 🔹 当前用户的所有测试记录
    const list = ref<MyTestItem[]>([])

    const ongoingList = computed(() =>
        list.value.filter((item) => item.status === 'RUNNING'),
    )

    const completedList = computed(() =>
        list.value.filter((item) => item.status !== 'RUNNING'),
    )

    const completedCount = computed(() =>
        list.value.filter((item) => item.status === 'COMPLETED_WITH_REPORT').length,
    )

    function renderTitle(item: MyTestItem): string {
        const base = businessTypeLabelMap[item.business_type] || '我的测试'
        const mode = item.mode ? `（${item.mode}）` : ''
        return base + mode
    }

    function renderStatusText(item: MyTestItem): string {
        switch (item.status) {
            case 'RUNNING':
                return '进行中'
            case 'COMPLETED_NO_REPORT':
                return '已完成（报告生成中）'
            case 'COMPLETED_WITH_REPORT':
                return '已完成'
            default:
                return ''
        }
    }

    function renderProfileTitle(): string {
        if (!profile.value) return ''
        return profile.value.nick_name || '同学'
    }

    function renderProfileSub(): string {
        if (!profile.value) return ''
        const parts: string[] = []
        if (profile.value.school_name) parts.push(profile.value.school_name)
        if (profile.value.city || profile.value.province) {
            parts.push(profile.value.city || profile.value.province || '')
        }
        return parts.join(' ｜ ')
    }

    function getAvatarInitial(): string {
        if (!profile.value?.nick_name) return '同'
        return profile.value.nick_name[0] || '同'
    }

    async function fetchMyTests() {
        loading.value = true
        showLoading('正在加载你的测评记录…')
        try {
            // 后端实现：GET /api/tests/my -> MyTestsResponse
            const resp = await apiRequest<MyTestsResponse>('/api/tests/profile')
            if (resp) {
                profile.value = resp.profile
                list.value = resp.tests || []
            } else {
                profile.value = null
                list.value = []
            }
        } catch (e) {
            console.error('[MyTests] fetchMyTests failed', e)
            showAlert('加载测评记录失败，请稍后重试')
        } finally {
            loading.value = false
            hideLoading()
        }
    }

    async function handleContinueTest(item: MyTestItem) {
        showLoading('正在为你恢复测试进度…')
        try {
            // 建议后端提供：GET /api/tests/:public_id/next-route -> { path: string }
            const resp = await apiRequest<{ path: string }>(
                `/api/tests/${encodeURIComponent(item.public_id)}/next-route`,
            )
            if (resp?.path) {
                router.push(resp.path)
            } else {
                showAlert('暂时无法恢复这次测试，请稍后重试')
            }
        } catch (e) {
            console.error('[MyTests] handleContinueTest failed', e)
            showAlert('恢复测试失败，请稍后重试')
        } finally {
            hideLoading()
        }
    }

    function handleOpenReport(item: MyTestItem) {
        router.push({
            name: 'test-report',
            params: {
                typ: item.business_type,
            },
            query: {
                public_id: item.public_id,
            },
        })
    }

    function handleClickCompletedNoReport(item: MyTestItem) {
        showAlert('报告正在生成中，请稍后在此页面查看。')
    }

    function handleBackHome() {
        router.push({ name: 'home' })
    }

    onMounted(() => {
        fetchMyTests().then()
    })

    return {
        loading,
        profile,
        list,
        ongoingList,
        completedList,
        completedCount,

        renderTitle,
        renderStatusText,
        renderProfileTitle,
        renderProfileSub,
        getAvatarInitial,
        formatDateTime,

        handleContinueTest,
        handleOpenReport,
        handleClickCompletedNoReport,
        handleBackHome,
    }
}
