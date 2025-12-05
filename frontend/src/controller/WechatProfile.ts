import {ref, computed, onMounted, watch} from 'vue'
import { useRouter } from 'vue-router'
import { API_PATHS, apiRequest } from '@/api'
import { useGlobalLoading } from '@/controller/useGlobalLoading'
import { useAlert } from '@/controller/useAlert'
import {chinaProvinces} from "@/controller/chinaRegions";
import {isValidChinaMobile} from "@/controller/common";

export interface MyTestItem {
    public_id: string
    business_type: string
    mode: string
    create_at: string
    report_status?: number | null
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
    const profile = ref<UserProfile | null>(null)
    const list = ref<MyTestItem[]>([])
    const editingExtra = ref(false)
    const extraForm = ref({
        mobile: '',
        study_id: '',
        school_name: '',
        province: '',
        city: '',
    })
    const ongoingList = computed(() =>
        list.value.filter(
            (item) => item.report_status === undefined || item.report_status === 0,
        ),
    )
    const completedList = computed(() =>
        list.value.filter((item) => item.report_status === 1),
    )
    const completedCount = computed(() => completedList.value.length)
    const latestCompleted = computed(() => completedList.value[0] || null)
    const activeTab = ref<'ongoing' | 'completed'>('ongoing')

    // 省市下拉
    const provinces = chinaProvinces
    const selectedProvince = ref<string>('')
    const selectedCity = ref<string>('')

    const currentCities = computed(() => {
        const prov = provinces.find(p => p.name === selectedProvince.value)
        return prov ? prov.cities : []
    })

    watch(selectedProvince, () => {
        selectedCity.value = ''
    })



    function setActiveTab(tab: 'ongoing' | 'completed') {
        activeTab.value = tab
    }

    function renderTitle(item: MyTestItem): string {
        const base = businessTypeLabelMap[item.business_type] || '我的测试'
        const mode = item.mode ? `（${item.mode}）` : ''
        return base + mode
    }

    function renderStatusText(item: MyTestItem): string {
        switch (item.report_status) {
            case 1:
                return '完成'
            default:
                return '进行中'
        }
    }

    function renderProfileTitle(): string {
        if (!profile.value) return ''
        return profile.value.nick_name || '同学'
    }


    function getAvatarInitial(): string {
        if (!profile.value?.nick_name) return '同'
        return profile.value.nick_name[0] || '同'
    }

    async function fetchMyTests() {
        showLoading('正在加载你的测评记录…')
        try {
            // 后端实现：GET /api/tests/my -> MyTestsResponse
            const resp = await apiRequest<MyTestsResponse>(API_PATHS.WECHAT_MY_PROFILE)
            if (resp) {
                profile.value = resp.profile
                list.value = resp.tests || []

                extraForm.value = {
                    mobile: resp.profile.mobile || '',
                    study_id: resp.profile.study_id || '',
                    school_name: resp.profile.school_name || '',
                    province: resp.profile.province || '',
                    city: resp.profile.city || '',
                }

            } else {
                profile.value = null
                list.value = []
            }
        } catch (e) {
            console.error('[MyTests] fetchMyTests failed', e)
            showAlert('加载测评记录失败，请稍后重试')
        } finally {
            hideLoading()
        }
    }

    function startEditExtra() {
        if (profile.value) {
            extraForm.value = {
                mobile: profile.value.mobile || '',
                study_id: profile.value.study_id || '',
                school_name: profile.value.school_name || '',
                province: profile.value.province || '',
                city: profile.value.city || '',
            }

            selectedProvince.value = profile.value.province || ''
            selectedCity.value = profile.value.city || ''
        }
        editingExtra.value = true
    }


    function cancelEditExtra() {
        editingExtra.value = false
    }

    async function saveExtra() {
        if (!profile.value) return

        const originalMobile = (profile.value.mobile || '').trim()
        const currentMobile = (extraForm.value.mobile || '').trim()
        const mobileChanged = currentMobile !== originalMobile

        // 要发给后端的 body
        const body: any = {
            study_id: extraForm.value.study_id || undefined,
            school_name: extraForm.value.school_name || undefined,
            province: selectedProvince.value || undefined,
            city: selectedCity.value || undefined,
        }

        if (mobileChanged) {
            // 用户确实对手机号做了修改
            if (currentMobile !== '') {
                // 填了非空 → 必须是合法手机号
                if (!isValidChinaMobile(currentMobile)) {
                    showAlert('请输入有效的中国大陆手机号码')
                    return
                }
                body.mobile = currentMobile       // 新手机号
            } else {
                // 用户明确把手机号删空：允许清空
                body.mobile = ''                  // 会被后端当成“清空手机号”
            }
        }

        showLoading('正在保存你的资料…')
        try {
            await apiRequest(API_PATHS.WECHAT_UPDATE_PROFILE, {
                method: 'POST',
                body,
            })

            await fetchMyTests()   // 再次拿到脱敏后的 profile（比如 138****0000）
            editingExtra.value = false
            showAlert('资料已保存')
        } catch (e) {
            console.error('[MyTests] saveExtra failed', e)
            showAlert('保存资料失败，请稍后重试')
        } finally {
            hideLoading()
        }
    }


    async function handleContinueTest(item: MyTestItem) {
        showLoading('正在为你恢复测试进度…')
        try {

        } catch (e) {
            console.error('[MyTests] handleContinueTest failed', e)
            showAlert('恢复测试失败，请稍后重试')
        } finally {
            hideLoading()
        }
    }

    const reportPreviewVisible = ref(false)
    const reportPreviewTarget = ref<MyTestItem | null>(null)

    function openReportPreview(target?: MyTestItem) {
        const chosen = target || latestCompleted.value
        if (!chosen) {
            showAlert('暂时没有可预览的报告')
            return
        }
        reportPreviewTarget.value = chosen
        reportPreviewVisible.value = true
    }

    function closeReportPreview() {
        reportPreviewVisible.value = false
    }

    function handleBackHome() {
        showLoading("返回首页......")
        router.push({ name: 'home' }).finally(()=>hideLoading())
    }

    onMounted(() => {
        fetchMyTests().then()
    })

    return {
        profile,
        list,
        ongoingList,
        completedList,
        completedCount,
        latestCompleted,

        renderTitle,
        renderStatusText,
        renderProfileTitle,
        getAvatarInitial,
        formatDateTime,

        handleContinueTest,
        handleBackHome,

        reportPreviewVisible,
        reportPreviewTarget,
        openReportPreview,
        closeReportPreview,

        editingExtra,
        extraForm,
        startEditExtra,
        cancelEditExtra,
        saveExtra,
        activeTab,
        setActiveTab,

        provinces,
        selectedProvince,
        selectedCity,
        currentCities,
    }
}
