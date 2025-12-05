import { Ref, computed, onMounted, onUnmounted, reactive, ref, unref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useGlobalLoading } from '@/controller/useGlobalLoading'
import { type TestRecordDTO, useTestSession } from '@/controller/testSession'
import { API_PATHS, apiRequest } from '@/api'
import { useAlert } from '@/controller/useAlert'
import {
    Mode312,
    Mode33,
    ModeOption,
    PlanInfo,
    subjectLabelMap,
    TestTypeBasic,
    useSseLogs,
    useSubscriptBySSE,
} from '@/controller/common'

/* ====================== 类型定义 ====================== */

export interface ComboMetric {
    label: string
    value: number
}

export interface ReportCombo {
    rankLabel: string
    name: string
    score: string
    theme: string
    metrics: ComboMetric[]
    recommendAdvice: string
    recommendExplain: string
}

export interface ReportSubjectScore {
    subject: string
    interest_z: number
    ability_z: number
    zgap: number
    ability_share: number
    fit: number
    fit_score?: number
}

export interface ReportCommonBlock {
    global_cosine: number
    quality_score: number
    global_cosine_score?: number
    quality_score_score?: number
    subjects: ReportSubjectScore[]
}

export interface ReportRadarBlock {
    subjects: string[]
    interest_pct: number[]
    ability_pct: number[]
}

export interface ReportCommonScore {
    common: ReportCommonBlock
    radar: ReportRadarBlock
}

export interface Report312Combo {
    aux1: string
    aux2: string
    avg_fit: number
    min_fit: number
    combo_cos: number
    auxAbility: number
    coverage: number
    mix_penalty: number
    s23: number
    s_final_combo: number
    combo_score?: number
}

export interface Report312Anchor {
    subject: string
    fit: number
    ability_norm: number
    term_fit: number
    term_ability: number
    term_coverage: number
    s1: number
    combos: Report312Combo[]
    s_final: number
    s_final_score?: number
}

export type ReportRecommend312 = Record<string, Report312Anchor>

export interface Recommend33Combo {
    subjects: string[]          // ["CHE","BIO","GEO"]
    avg_fit: number             // 平均匹配度
    min_ability: number         // 最低能力等级
    rarity: number              // 稀有度
    risk_penalty: number        // 风险惩罚
    score: number               // 综合推荐得分（原始）
    combo_cosine: number        // 兴趣/能力方向一致性
    recommend_score?: number    // 用于排序/展示的最终分数
}

export interface ReportRecommend33 {
    top_combinations: Recommend33Combo[]
}

export interface ComboChartMetric {
    key: string        // 内部字段名，例如 'score' / 'coverage' / 'mix_penalty'
    label: string      // 展示名
    value: number
}

export interface ComboChartItem {
    comboKey: string        // 组合编码，例如 "PHY_CHE_BIO"
    metrics: ComboChartMetric[]
}

export interface ReportRawData {
    uid: string
    nick_name?: string
    avatar_url?: string
    study_id?: string
    school_name?: string
    province?: string
    city?: string

    mode: ModeOption
    generated_at: string
    expired_at: string

    common_score: ReportCommonScore
    recommend_33: ReportRecommend33 | null
    recommend_312: ReportRecommend312 | null
    ai_content: string | null
}

export interface ReportOverviewInfo {
    mode: ModeOption
    studentLocation: string
    generateDate: string
    expireDate: string
    studentNo: string
    schoolName: string
    account: string
}

export interface Mode33ViewModel {
    overviewText: string
    chartCombos: ComboChartItem[]
    rarityRiskPairs: ComboChartItem[]
    topCombos: ReportCombo[]
}

export interface ComboDetail {
    combo_name: string
    combo_description: string
    combo_advice: string
}

export interface ModeSectionItem {
    overview_text: string
    combo_details: ComboDetail[]
}

export interface Mode312Section {
    [modeKey: string]: ModeSectionItem
}

export interface Mode33ComboExplain {
    combo_description: string
    combo_advice: string
}

export interface Mode33Section {
    mode33_overview_text: string
    mode33_combo_details: Record<string, Mode33ComboExplain>
}

export type ModeSection = Mode312Section | Mode33Section

export interface CommonSection {
    report_validity_text: string
    subjects_summary_text: string
}

export interface FinalAIReport {
    mode: ModeOption
    report_validity: string
    core_trends: string
    mode_strategy: string
    student_view: string
    parent_view: string
    risk_diagnosis: string
    strategic_conclusion: string
}

export interface AIReportPayload {
    common_section: CommonSection
    mode_section: ModeSection
    final_report: FinalAIReport
}

export interface Mode312OverviewStrips {
    phyOverviewText: string
    hisOverviewText: string

    phyScoreBars: ComboChartItem[]
    phyCoverageRiskBars: ComboChartItem[]

    hisScoreBars: ComboChartItem[]
    hisCoverageRiskBars: ComboChartItem[]

    phyTopCombos: ReportCombo[]
    hisTopCombos: ReportCombo[]

    phyS1: number
    hisS1: number
}

/* ====================== 模块级共享状态（单例） ====================== */

// 这个本来就是模块级导出，保持不变
export const aiReportData = ref<AIReportPayload | null>(null)

// 报告原始数据
const rawReportData = ref<ReportRawData | null>(null)

// 抬头信息
const overview = reactive<ReportOverviewInfo>({
    mode: Mode33,
    studentLocation: '',
    generateDate: '',
    expireDate: '',
    studentNo: '',
    schoolName: '',
    account: '',
})

// 页面根元素，用于打印区域等
const reportPageRoot = ref<HTMLElement | null>(null)

// 控制状态
const aiLoading = ref(false)
const showFinishLetter = ref(false)
const paymentDialogShow = ref(false)
const currentPlan = ref<PlanInfo | null>(null)

/* ====================== 工具函数 ====================== */

function formatComboName33(subjects: string[]): string {
    return subjects.map(s => subjectLabelMap[s] ?? s).join(' + ')
}

function rankLabelFor33(index: number): string {
    const preset = ['第一档', '第二档', '第三档']
    return preset[index] ?? `第${index + 1}档`
}

function themeFor33(index: number): string {
    if (index === 0) return 'primary'
    if (index === 1) return 'blue'
    return 'yellow'
}

function formatDate(dateStr?: string | null): string {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    if (Number.isNaN(d.getTime())) return ''
    const y = d.getFullYear()
    const m = `${d.getMonth() + 1}`.padStart(2, '0')
    const day = `${d.getDate()}`.padStart(2, '0')
    return `${y}-${m}-${day}`
}

function applyReportOverview(data: ReportRawData) {
    overview.mode = data.mode
    overview.account = data.nick_name || data.uid || ''

    overview.generateDate = formatDate(data.generated_at)
    overview.expireDate = formatDate(data.expired_at)

    // 原代码写成了省+省，这里顺便修正为 省+市
    const prov = data.province || ''
    const city = data.city || ''
    overview.studentLocation = `${prov ? prov + '省' : ''}${city ? city + '市' : ''}`

    overview.studentNo = data.study_id || ''
    overview.schoolName = data.school_name || ''
}

/* ====================== 依赖共享状态的纯 computed ====================== */

const subjectRadar = computed<ReportRadarBlock | null>(() => {
    const r = rawReportData.value?.common_score?.radar
    if (!r || !r.subjects || !r.subjects.length) return null
    return r
})

const isMode33 = computed(() => overview.mode === Mode33)
const isMode312 = computed(() => overview.mode === Mode312)

const mode312OverviewStrips = computed<Mode312OverviewStrips | null>(() => {
    const raw = rawReportData.value
    const ai = aiReportData.value

    if (!raw || !ai) return null
    if (raw.mode !== '3+1+2') return null
    if (!raw.recommend_312) return null

    const section = ai.mode_section as Mode312Section
    const phyText = section['mode312_PHY']?.overview_text ?? ''
    const hisText = section['mode312_HIS']?.overview_text ?? ''

    const phyAnchor =
        raw.recommend_312['anchor_phy'] ??
        raw.recommend_312['PHY'] ??
        Object.values(raw.recommend_312).find(a => a.subject === 'PHY') ??
        null

    const hisAnchor =
        raw.recommend_312['anchor_his'] ??
        raw.recommend_312['HIS'] ??
        Object.values(raw.recommend_312).find(a => a.subject === 'HIS') ??
        null

    if (!phyAnchor && !hisAnchor) return null

    const buildComboKey = (anchor: Report312Anchor, combo: Report312Combo) =>
        `${anchor.subject}_${combo.aux1}_${combo.aux2}`

    const buildScoreItems = (anchor: Report312Anchor | null): ComboChartItem[] => {
        if (!anchor || !anchor.combos?.length) return []
        return anchor.combos.map(c => ({
            comboKey: buildComboKey(anchor, c),
            metrics: [
                {
                    key: 'score',
                    label: '综合得分',
                    value: c.combo_score!,
                },
            ],
        }))
    }

    const buildCoverageRiskItems = (anchor: Report312Anchor | null): ComboChartItem[] => {
        if (!anchor || !anchor.combos?.length) return []
        return anchor.combos.map(c => ({
            comboKey: buildComboKey(anchor, c),
            metrics: [
                {
                    key: 'coverage',
                    label: '专业覆盖率',
                    value: c.coverage,
                },
                {
                    key: 'mix_penalty',
                    label: '组合风险',
                    value: c.mix_penalty,
                },
            ],
        }))
    }

    const buildGroupCombos = (
        anchor: Report312Anchor | null,
        sec: ModeSectionItem | undefined,
    ): ReportCombo[] => {
        if (!anchor || !anchor.combos?.length) return []

        const groupLabel = `${subjectLabelMap[anchor.subject] ?? anchor.subject}组`

        const explainIndex: Record<string, ComboDetail> = {}
        if (sec?.combo_details?.length) {
            for (const detail of sec.combo_details) {
                if (detail.combo_name) {
                    explainIndex[detail.combo_name] = detail
                }
            }
        }

        const sorted = [...anchor.combos].sort((a, b) => b.s_final_combo - a.s_final_combo)

        return sorted.map((c, index) => {
            const comboKey = buildComboKey(anchor, c)
            const explain = explainIndex[comboKey]

            let rankLabel = `${groupLabel}·第${index + 1}档`
            if (index === 0) rankLabel = `${groupLabel}·首选`
            else if (index === 1) rankLabel = `${groupLabel}·备选一`
            else if (index === 2) rankLabel = `${groupLabel}·备选二`

            const theme =
                index === 0 ? 'primary'
                    : index === 1 ? 'blue'
                        : 'yellow'

            const subs = [anchor.subject, c.aux1, c.aux2]
            const name = subs.map(s => subjectLabelMap[s] ?? s).join(' + ')

            return {
                rankLabel,
                name,
                score: Math.round(c.combo_score!).toString(),
                theme,
                metrics: [
                    { label: '辅科平均匹配度', value: c.avg_fit },
                    { label: '辅科平均能力值', value: c.auxAbility },
                    { label: '辅科最低匹配度', value: c.min_fit },
                    { label: '辅科一致性', value: c.combo_cos },
                ],
                recommendExplain: explain?.combo_description ?? '',
                recommendAdvice: explain?.combo_advice ?? '',
            }
        })
    }

    const phySection = section['mode312_PHY']
    const hisSection = section['mode312_HIS']
    const phyTopCombos = buildGroupCombos(phyAnchor, phySection)
    const hisTopCombos = buildGroupCombos(hisAnchor, hisSection)

    return {
        phyOverviewText: phyText,
        hisOverviewText: hisText,
        phyScoreBars: buildScoreItems(phyAnchor),
        hisScoreBars: buildScoreItems(hisAnchor),
        phyCoverageRiskBars: buildCoverageRiskItems(phyAnchor),
        hisCoverageRiskBars: buildCoverageRiskItems(hisAnchor),
        phyTopCombos,
        hisTopCombos,
        phyS1: phyAnchor?.s1 ?? 0,
        hisS1: hisAnchor?.s1 ?? 0,
    }
})

const mode33View = computed<Mode33ViewModel | null>(() => {
    const raw = rawReportData.value
    const ai = aiReportData.value
    if (!raw || !ai) return null

    const recommend = raw.recommend_33
    const section = ai.mode_section as Mode33Section | any

    const overviewText: string = section?.mode33_overview_text ?? ''
    const combosRaw = recommend?.top_combinations ?? []

    const chartCombos: ComboChartItem[] = combosRaw.map(c => ({
        comboKey: c.subjects.join('_'),
        metrics: [
            {
                key: 'score',
                label: '综合得分',
                value: c.recommend_score!,
            },
        ],
    }))

    const rarityRiskPairs: ComboChartItem[] = combosRaw.map(c => ({
        comboKey: c.subjects.join('_'),
        metrics: [
            {
                key: 'coverage',
                label: '稀有度',
                value: c.rarity,
            },
            {
                key: 'mix_penalty',
                label: '风险惩罚',
                value: c.risk_penalty,
            },
        ],
    }))

    const sortedCombos = [...combosRaw].sort(
        (a, b) => (b.recommend_score ?? 0) - (a.recommend_score ?? 0),
    )

    const topCombos: ReportCombo[] = sortedCombos.map((c, index) => {
        const comboKey = c.subjects.join('_')
        const ai_combo_detail = section?.mode33_combo_details?.[comboKey]

        return {
            rankLabel: rankLabelFor33(index),
            name: formatComboName33(c.subjects),
            score: Math.round(c.recommend_score ?? 0).toString(),
            theme: themeFor33(index),
            recommendExplain: ai_combo_detail?.combo_description ?? '',
            recommendAdvice: ai_combo_detail?.combo_advice ?? '',
            metrics: [
                {
                    label: '平均匹配度',
                    value: c.avg_fit,
                },
                {
                    label: '最低能力等级',
                    value: c.min_ability,
                },
                {
                    label: '方向协同性',
                    value: c.combo_cosine,
                },
            ],
        }
    })

    return {
        overviewText,
        chartCombos,
        rarityRiskPairs,
        topCombos,
    }
})

const finalReport = computed<FinalAIReport | null>(() => {
    const ai = aiReportData.value
    if (!ai || !ai.final_report) return null
    return ai.final_report
})

/* ====================== composable A：控制逻辑（父组件用） ====================== */

interface ReportControllerOptions {
    publicId?: string | Ref<string>
    businessType?: string | Ref<string>
    autoQueryOnMounted?: boolean
}

export function useReportController(options?: ReportControllerOptions) {
    const { showLoading, hideLoading } = useGlobalLoading()
    const { state, resetSession } = useTestSession()
    const { showAlert } = useAlert()
    const route = useRoute()
    const router = useRouter()

    const record = computed<TestRecordDTO | undefined>(() => state.record)
    const routeBusinessType = computed(() => String(route.params.typ ?? ''))
    const routePublicId = computed(() => String(route.query.public_id ?? ''))
    const businessType = computed(() => {
        const manualBusinessType = options?.businessType ? unref(options.businessType) : ''
        if (manualBusinessType) return manualBusinessType
        if (record.value?.business_type) return record.value.business_type
        if (routeBusinessType.value) return routeBusinessType.value
        return TestTypeBasic
    })
    const publicId = computed(() => {
        const manualPublicId = options?.publicId ? unref(options.publicId) : ''
        if (manualPublicId) return manualPublicId
        if (record.value?.public_id) return record.value.public_id
        return routePublicId.value
    })

    const { truncatedLatestMessage, handleSseMsg } = useSseLogs(8, 20)

    function handleSseError(err: Error) {
        console.log('------>>> sse channel error:', err)
        showAlert('获取测试报告失败，请稍后再试:' + err)
        aiLoading.value = false
    }

    function handleSseDone(raw: string) {
        try {
            let parsed = JSON.parse(raw) as AIReportPayload
            console.log('------>>> parsed object:', parsed)
            aiReportData.value = parsed
        } catch (e) {
            showAlert('解析报告数据失败，请稍后再试' + e)
        } finally {
            aiLoading.value = false
        }
    }

    const sseCtrl = useSubscriptBySSE(
        `${API_PATHS.SSE_REPORT_SUB}${publicId.value}`,
        {
            autoStart: false,
            onOpen: () => {
                aiLoading.value = true
            },
            onError: handleSseError,
            onMsg: handleSseMsg,
            onClose: () => {
                aiLoading.value = false
            },
            onDone: handleSseDone,
        },
    )

    async function generateReport() {
        paymentDialogShow.value = false
        showLoading('正在准备智能分析参数', 20_000)
        try {
            if (!publicId.value) {
                showAlert('未找到试卷编号', () => {
                    router.replace('/home').then()
                })
                return
            }

            const resp = await apiRequest<ReportRawData>(API_PATHS.GENERATE_REPORT, {
                method: 'POST',
                body: {
                    public_id: publicId.value,
                    business_type: businessType.value,
                },
            })

            rawReportData.value = resp
            applyReportOverview(resp)

            if (!resp.ai_content) {
                sseCtrl.start()
            } else {
                let aiContent = JSON.parse(resp.ai_content) as AIReportPayload
                if (typeof aiContent === 'string') {
                    aiContent = JSON.parse(aiContent) as AIReportPayload
                }
                aiReportData.value = aiContent
            }
        } catch (e) {
            showAlert('生成 AI 参数失败:' + e)
        } finally {
            hideLoading()
        }
    }

    async function queryCurPlan() {
        showLoading('加载支付信息')
        try {
            currentPlan.value = await apiRequest<PlanInfo>(API_PATHS.LOAD_CUR_PRODUCT, {
                method: 'POST',
                body: {
                    public_id: publicId.value,
                },
            })

            const hasPaid = currentPlan.value?.has_paid || false
            if (hasPaid) {
                await generateReport()
            } else {
                paymentDialogShow.value = true
            }
        } catch (e) {
            showAlert('查询产品价格失败:' + e)
        } finally {
            hideLoading()
        }
    }

    const handleLetterConfirm = async () => {
        try {
            await apiRequest<ReportRawData>(API_PATHS.FINISH_REPORT, {
                method: 'POST',
                body: {
                    public_id: publicId.value,
                    business_type: businessType.value,
                },
            })
        } catch (e) {
            console.error('结束报告失败：' + e)
        } finally {
            resetSession()
            showFinishLetter.value = false
            router.replace('/').then()
        }
    }

    const handleExportPdf = () => {
        const oldTitle = document.title

        const account = overview.account || ''
        const date = overview.generateDate || ''

        document.title = `智择未来 · AI选科全景分析报告-${account || '报告'}${date ? '-' + date : ''}`

        window.print()

        setTimeout(() => {
            document.title = oldTitle
        }, 1000)
    }

    onMounted(async () => {
        if (options?.autoQueryOnMounted === false) return
        await queryCurPlan()
    })

    onUnmounted(() => {
        sseCtrl.stop()
    })

    return {
        // 基础
        route,
        state,
        businessType,
        publicId,

        // loading & SSE
        aiLoading,
        truncatedLatestMessage,

        // 流程控制
        showFinishLetter,
        handleLetterConfirm,
        handleExportPdf,

        paymentDialogShow,
        currentPlan,
        generateReport,
    }
}

/* ====================== composable B：视图模型（子组件用） ====================== */

export function useReportView() {
    return {
        overview,
        rawReportData,
        subjectRadar,
        isMode33,
        isMode312,
        mode33View,
        mode312OverviewStrips,
        finalReport,
        reportPageRoot,
    }
}
