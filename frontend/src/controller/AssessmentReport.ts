import {ref, computed, onMounted, reactive} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useGlobalLoading} from '@/controller/useGlobalLoading'
import {useTestSession} from '@/controller/testSession'
import {apiRequest} from "@/api";
import {useAlert} from "@/controller/useAlert";
import {Mode312, Mode33, ModeOption, subjectLabelMap, useSseLogs, useSubscriptBySSE} from "@/controller/common";

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

export interface SummaryCard {
    title: string
    content: string
}


// ===== 报告接口类型 =====
// 单科 common_score.common.subjects 里的条目
export interface ReportSubjectScore {
    subject: string
    interest_z: number
    ability_z: number
    zgap: number
    ability_share: number
    fit: number
}

// common_score.common
export interface ReportCommonBlock {
    global_cosine: number
    quality_score: number
    subjects: ReportSubjectScore[]
}

// common_score.radar
export interface ReportRadarBlock {
    subjects: string[]
    interest_pct: number[]
    ability_pct: number[]
}

// common_score 整体
export interface ReportCommonScore {
    common: ReportCommonBlock
    radar: ReportRadarBlock
}

// recommend_312[*].combos 里的条目
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
}

// recommend_312 里每个 anchor_* 块
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
}

// recommend_312 顶层：key 例如 "anchor_phy" / "anchor_his"…
export type ReportRecommend312 = Record<string, Report312Anchor>


// 3+3 推荐组合的原始数值
export interface Recommend33Combo {
    subjects: string[]          // ["CHE","BIO","GEO"]
    avg_fit: number             // 平均匹配度
    min_ability: number         // 最低能力等级（原 JSON 里的 min_ability）
    rarity: number              // 稀有度
    risk_penalty: number        // 风险惩罚
    score: number               // 综合推荐得分
    combo_cosine: number        // 兴趣/能力方向一致性
}

export interface ReportRecommend33 {
    top_combinations: Recommend33Combo[]
}

// 通用：一个组合 + 若干数值（这里我们约定长度为 1 或 2）
export interface ComboChartMetric {
    key: string        // 内部字段名，例如 'score' / 's_final' / 'coverage' / 'mix_penalty'
    label: string      // 展示用名称，例如 '综合得分' / '专业覆盖率'
    value: number
}

// 通用：图表用的“组合 + 指标”
export interface ComboChartItem {
    comboKey: string        // 组合编码，例如 "PHY_CHE_BIO"
    metrics: ComboChartMetric[]   // 长度 1 = 单指标；长度 2 = 双指标
}


// 整体响应
export interface ReportRawData {
    uid: string
    mode: ModeOption
    generate_at: string
    expired_at: string
    common_score: ReportCommonScore
    recommend_33: ReportRecommend33 | null
    recommend_312: ReportRecommend312 | null
}

function getAiReportParam(publicID: string, businessTyp: string) {
    return apiRequest<ReportRawData>('/api/generate_report', {
        method: 'POST',
        body: {
            public_id: publicID,
            business_type: businessTyp
        },
    })
}


export interface ReportOverviewInfo {
    mode: ModeOption;            // 模式：3+3 / 3+1+2
    studentLocation: string; // 学生所在地
    generateDate: string;
    expireDate: string;
    studentNo: string;       // 学生号
    schoolName: string;      // 学校名称
    account: string;         // 问心台账号
}



export interface Mode33ViewModel {
    overviewText: string     // mode33_overview_text
    chartCombos: ComboChartItem[]
    rarityRiskPairs: ComboChartItem[]
    topCombos: ReportCombo[]
}

// ==== 报告概览卡片（顶部那张）用到的字段 ====

// 简单日期格式化：2025-11-22
function formatDate(dateStr?: string | null): string {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    if (Number.isNaN(d.getTime())) return ''
    const y = d.getFullYear()
    const m = `${d.getMonth() + 1}`.padStart(2, '0')
    const day = `${d.getDate()}`.padStart(2, '0')
    return `${y}-${m}-${day}`
}


// 一条组合详情
export interface ComboDetail {
    combo_name: string
    combo_description: string
    combo_advice: string
}

// 某一个“模式小节”（例如 mode312_PHY、mode312_HIS）
export interface ModeSectionItem {
    overview_text: string
    combo_details: ComboDetail[]
}

// 整个 mode_section：key 是 "mode312_PHY" / "mode312_HIS" 之类
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

// 最终：AIReportPayload 里的 mode_section 允许两种形状
export type ModeSection = Mode312Section | Mode33Section

// common_section
export interface CommonSection {
    report_validity_text: string
    subjects_summary_text: string
}

// final_report 部分
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

// 整个 AI 报告 payload
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

export const aiReportData = ref<AIReportPayload | null>(null)

export function useReportPage() {
    const {showLoading, hideLoading} = useGlobalLoading()
    const route = useRoute()
    const {state} = useTestSession()
    const {showAlert} = useAlert()
    const router = useRouter()

    const rawReportData = ref<ReportRawData | null>(null)
    const subjectRadar = computed<ReportRadarBlock | null>(() => {
        const r = rawReportData.value?.common_score?.radar
        if (!r || !r.subjects || !r.subjects.length) return null
        return r
    })

    const overview = reactive<ReportOverviewInfo>({
        mode: Mode33,
        studentLocation: '',
        generateDate: '',
        expireDate: '',
        studentNo: '',
        schoolName: '',
        account: '',
    })
    // 当前报告模式：3+3 / 3+1+2
    const isMode33 = computed(() => overview.mode === Mode33)
    const isMode312 = computed(() => overview.mode === Mode312)

// === 新增：3+1+2 概览 + 图表数据 ===
    const mode312OverviewStrips = computed<Mode312OverviewStrips | null>(() => {
        const raw = rawReportData.value
        const ai = aiReportData.value

        // 原始数据 / AI 数据任意一个还没到，就先不给
        if (!raw || !ai) return null
        if (raw.mode !== '3+1+2') return null
        if (!raw.recommend_312) return null

        const section = ai.mode_section as Mode312Section
        const phyText = section['mode312_PHY']?.overview_text ?? ''
        const hisText = section['mode312_HIS']?.overview_text ?? ''

        // 数值：来自原始 recommend_312 的 anchor_phy / anchor_his
        const phyAnchor =
            raw.recommend_312['anchor_phy'] ??
            raw.recommend_312['PHY'] ?? // 兜底：如果后端改 key，只按 subject 找
            Object.values(raw.recommend_312).find(a => a.subject === 'PHY') ??
            null

        const hisAnchor =
            raw.recommend_312['anchor_his'] ??
            raw.recommend_312['HIS'] ??
            Object.values(raw.recommend_312).find(a => a.subject === 'HIS') ??
            null

        const buildComboKey = (anchor: Report312Anchor, combo: Report312Combo) =>
            `${anchor.subject}_${combo.aux1}_${combo.aux2}`

        const buildScoreItems = (anchor: Report312Anchor | null): ComboChartItem[] => {
            if (!anchor || !anchor.combos?.length) return []
            return anchor.combos.map(c => ({
                comboKey: buildComboKey(anchor, c),
                metrics: [
                    {
                        key: 's_final',
                        label: '综合得分',
                        value: c.s_final_combo,
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

            // “物理组”“历史组”
            const groupLabel = `${subjectLabelMap[anchor.subject] ?? anchor.subject}组`

            // AI 文本索引：combo_name -> ComboDetail
            const explainIndex: Record<string, ComboDetail> = {}
            if (sec?.combo_details?.length) {
                for (const detail of sec.combo_details) {
                    if (detail.combo_name) {
                        explainIndex[detail.combo_name] = detail
                    }
                }
            }

            // 按 s_final_combo 从高到低排序
            const sorted = [...anchor.combos].sort((a, b) => b.s_final_combo - a.s_final_combo)

            return sorted.map((c, index) => {
                const comboKey = buildComboKey(anchor, c) // e.g. PHY_CHE_BIO
                const explain = explainIndex[comboKey]

                // 档位文案
                let rankLabel = `${groupLabel}·第${index + 1}档`
                if (index === 0) rankLabel = `${groupLabel}·首选`
                else if (index === 1) rankLabel = `${groupLabel}·备选一`
                else if (index === 2) rankLabel = `${groupLabel}·备选二`

                // 颜色主题：沿用 3+3 的规则
                const theme =
                    index === 0 ? 'primary'
                        : index === 1 ? 'blue'
                            : 'yellow'

                // 组合中文名：主干 + 辅科1 + 辅科2
                const subs = [anchor.subject, c.aux1, c.aux2]
                const name = subs.map(s => subjectLabelMap[s] ?? s).join(' + ')

                return {
                    rankLabel,
                    name,
                    score: c.s_final_combo.toFixed(3),
                    theme,
                    metrics: [
                        {label: '辅科平均匹配度', value: c.avg_fit},
                        {label: '辅科平均能力值', value: c.auxAbility},
                        {label: '辅科最低匹配度', value: c.min_fit},
                        {label: '辅科一致性', value: c.combo_cos},
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
    computed(() => {
        if (!isMode33.value) return ''

        const section = aiReportData.value?.mode_section as any
        // 后面我们会把 ModeSection 换成 union 类型，这里先用 any 顶一下
        return section?.mode33_overview_text ?? ''
    });
    const businessType = computed(() => {
        return String(route.params.typ ?? state.businessType ?? '')
    })
    // 之后接入后端 / SSE 时再开启，这里先保持为 false
    const aiLoading = ref(false)


    // 3+3：科目数组 -> “化学 + 生物 + 地理”
    function formatComboName33(subjects: string[]): string {
        return subjects.map(s => subjectLabelMap[s] ?? s).join(' + ')
    }

// 3+3：第几个组合 -> 档位文案
    function rankLabelFor33(index: number): string {
        const preset = ['第一档', '第二档', '第三档']
        return preset[index] ?? `第${index + 1}档`
    }

// 3+3：第几个组合 -> 颜色主题
    function themeFor33(index: number): string {
        if (index === 0) return 'primary' // 首选：主紫色
        if (index === 1) return 'blue'    // 备选一：蓝色
        return 'yellow'                   // 其他：黄色
    }

    // 3+3 模式统一视图：overview 文本 + 三个组合的原始 score
    const mode33View = computed<Mode33ViewModel | null>(() => {
        const raw = rawReportData.value
        const ai = aiReportData.value
        if (!raw || !ai) return null

        const recommend = raw.recommend_33
        const section: any = ai.mode_section

        const overviewText: string = section?.mode33_overview_text ?? ''
        const combosRaw = recommend?.top_combinations ?? []

        const chartCombos: ComboChartItem[] = combosRaw.map(c => ({
            comboKey: c.subjects.join('_'),  // "CHE_BIO_GEO"
            metrics: [
                {
                    key: 'score',
                    label: '综合得分',       // 图表 legend / tooltip 显示
                    value: c.score,
                },
            ],
        }))

        const rarityRiskPairs: ComboChartItem[] = combosRaw.map(c => ({
            comboKey: c.subjects.join('_'),  // "CHE_BIO_GEO"
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

        const sortedCombos = [...combosRaw].sort((a, b) => b.score - a.score)
        const topCombos: ReportCombo[] = sortedCombos.map((c, index) => {
            const comboKey = c.subjects.join('_')
            const ai_combo_detail = section?.mode33_combo_details[comboKey]
            return {
                rankLabel: rankLabelFor33(index),            // 第一档 / 第二档 / 第三档...
                name: formatComboName33(c.subjects),        // 物理 + 化学 + 生物 / 历史 + 地理 + 生物...
                score: c.score.toFixed(3),                  // 展示用分数，先保留三位小数
                theme: themeFor33(index),                   // primary / blue / yellow
                recommendExplain: ai_combo_detail?.combo_description,
                recommendAdvice:ai_combo_detail?.combo_advice,
                metrics:[{
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
                    },],
            }
        })

        return {
            overviewText,
            chartCombos,
            rarityRiskPairs,
            topCombos,
        }
    })

    const summaryCards = ref<SummaryCard[]>([
        {
            title: 'report_validity_1',
            content:
                '本报告基于当前测试结果，提供了较高可信度的选科建议，但仍需要结合学校课程安排与家庭实际情况综合考虑。',
        },
        {
            title: 'report_validity_2',
            content:
                '建议家长与学生共同阅读报告内容，重点关注兴趣与能力差异较大的科目，并适当安排后续的体验与辅助学习。',
        },
        {
            title: 'report_validity_3',
            content:
                '本报告不直接决定高考选科，仅作为重要参考工具，帮助你更系统地理解自己的优势与风险点。',
        },
    ])

    function applyReportOverview(data: ReportRawData) {
        overview.mode = data.mode
        overview.account = data.uid || ''

        overview.generateDate = formatDate(data.generate_at)
        overview.expireDate = formatDate(data.expired_at)

        overview.studentLocation = ''
        overview.studentNo = ''
        overview.schoolName = ''
    }

    function handleSseError(err: Error) {
        console.log('------>>> sse channel error:', err)
        showAlert('获取测试报告失败，请稍后再试:' + err)
        aiLoading.value = false
    }


    const {
        truncatedLatestMessage,
        handleSseMsg,
    } = useSseLogs(8, 20)

    function handleSseDone(raw: string) {
        try {
            let parsed = JSON.parse(raw) as AIReportPayload
            if (typeof parsed === 'string') {
                parsed = JSON.parse(parsed) as AIReportPayload
            }
            console.log('------>>> parsed object:', parsed)
            aiReportData.value = parsed

        } catch (e) {
            showAlert('获取测试题目失败，请稍后再试' + e)
        } finally {
            aiLoading.value = false;
        }
    }

    const sseCtrl = useSubscriptBySSE(`/api/sub/report/${state.recordPublicID!}`, {
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
    })


    onMounted(async () => {
        showLoading("正在准备智能分析参数", 20_000)
        try {
            const public_id = state.recordPublicID
            if (!public_id) {
                showAlert('未找到试卷编号', () => {
                    router.replace('/').then()
                })
                return
            }
            const resp = await getAiReportParam(public_id, businessType.value)
            rawReportData.value = resp;
            console.log("------>>>resp data:", rawReportData)
            applyReportOverview(resp);
            sseCtrl.start()
        } catch (e) {
            showAlert("生成 AI 参数失败:" + e);
        } finally {
            hideLoading();
        }
    })

    const finalReport = computed<FinalAIReport | null>(() => {
        const ai = aiReportData.value
        if (!ai || !ai.final_report) return null
        return ai.final_report
    })

    return {
        state,
        route,
        overview,
        businessType,
        aiLoading,
        truncatedLatestMessage,
        summaryCards,
        rawReportData,
        subjectRadar,
        isMode33,
        isMode312,
        mode33View,
        mode312OverviewStrips,
        finalReport,
    }
}
