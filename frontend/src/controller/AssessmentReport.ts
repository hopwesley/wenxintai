import {ref, computed, onMounted, reactive} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useGlobalLoading} from '@/controller/useGlobalLoading'
import {useTestSession} from '@/controller/testSession'
import {apiRequest} from "@/api";
import {useAlert} from "@/controller/useAlert";
import {Mode312, Mode33, ModeOption, useSseLogs, useSubscriptBySSE} from "@/controller/common";

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
    factorExplain: string[]
    recommendExplain: string[]
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


export interface Mode33ChartCombo {
    subjects: string[]       // ["CHE","BIO","GEO"]
    score: number            // 后端原始 score
}

export interface Mode33ViewModel {
    overviewText: string     // mode33_overview_text
    chartCombos: Mode33ChartCombo[]
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

function isMode33Section(
    mode: string,
    section: ModeSection
): section is Mode33Section {
    return mode === '3+3' && 'mode33_overview_text' in section
}

function isMode312Section(
    mode: string,
    section: ModeSection
): section is Mode312Section {
    return mode === '3+1+2' && !('mode33_overview_text' in section)
}


// 图表 1：每个组合的一根柱子，展示 s_final_combo
export interface Mode312ComboScoreBar {
    comboKey: string        // 组合代码，例如 "PHY_CHE_BIO"
    sFinal: number          // s_final_combo
}

// 图表 2：每个组合两根柱子，展示 coverage 和 mix_penalty
export interface Mode312ComboCoverageRiskBar {
    comboKey: string        // 组合代码，例如 "PHY_CHE_BIO"
    coverage: number        // coverage
    mixPenalty: number      // mix_penalty
}

export interface Mode312OverviewStrips {
    // 文字条：来源于 AI 报告里的 mode312_PHY / mode312_HIS.overview_text
    phyOverviewText: string
    hisOverviewText: string

    // 物理组两张图的数据
    phyScoreBars: Mode312ComboScoreBar[]             // 图 1：s_final_combo 柱状图
    phyCoverageRiskBars: Mode312ComboCoverageRiskBar[] // 图 2：coverage + mix_penalty

    // 历史组两张图的数据
    hisScoreBars: Mode312ComboScoreBar[]
    hisCoverageRiskBars: Mode312ComboCoverageRiskBar[]
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

        // 只在 AI 报告里确认为 3+1+2 时才解析 mode_section
        const section = ai.mode_section
        if (!isMode312Section(ai.final_report.mode, section)) {
            return null
        }

        // 文字：来自 AI 报告的 mode312_PHY / mode312_HIS
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

        // 小工具：组合代码
        const buildComboKey = (anchor: Report312Anchor, combo: Report312Combo) =>
            `${anchor.subject}_${combo.aux1}_${combo.aux2}`

        const buildScoreBars = (anchor: Report312Anchor | null): Mode312ComboScoreBar[] => {
            if (!anchor || !anchor.combos?.length) return []
            return anchor.combos.map(c => ({
                comboKey: buildComboKey(anchor, c),
                sFinal: c.s_final_combo,
            }))
        }

        const buildCoverageRiskBars = (
            anchor: Report312Anchor | null,
        ): Mode312ComboCoverageRiskBar[] => {
            if (!anchor || !anchor.combos?.length) return []
            return anchor.combos.map(c => ({
                comboKey: buildComboKey(anchor, c),
                coverage: c.coverage,
                mixPenalty: c.mix_penalty,
            }))
        }

        return {
            phyOverviewText: phyText,
            hisOverviewText: hisText,
            phyScoreBars: buildScoreBars(phyAnchor),
            hisScoreBars: buildScoreBars(hisAnchor),
            phyCoverageRiskBars: buildCoverageRiskBars(phyAnchor),
            hisCoverageRiskBars: buildCoverageRiskBars(hisAnchor),
        }
    })

    const mode33OverviewText = computed(() => {
        if (!isMode33.value) return ''

        const section = aiReportData.value?.mode_section as any
        // 后面我们会把 ModeSection 换成 union 类型，这里先用 any 顶一下
        return section?.mode33_overview_text ?? ''
    })

    // 备用：当前业务类型（暂时模板里没用，将来如果需要接后端可以直接用）
    const businessType = computed(() => {
        return String(route.params.typ ?? state.businessType ?? '')
    })
    // 之后接入后端 / SSE 时再开启，这里先保持为 false
    const aiLoading = ref(false)


    // 3+3 模式统一视图：overview 文本 + 三个组合的原始 score
    const mode33View = computed<Mode33ViewModel | null>(() => {
        const raw = rawReportData.value
        const ai = aiReportData.value
        if (!raw || !ai) return null

        const recommend = raw.recommend_33
        const section: any = ai.mode_section

        const overviewText: string = section?.mode33_overview_text ?? ''
        const combosRaw = recommend?.top_combinations ?? []

        const chartCombos: Mode33ChartCombo[] = combosRaw.map(c => ({
            subjects: c.subjects,
            score: c.score,
        }))

        return {
            overviewText,
            chartCombos,
        }
    })

    const recommendedCombos = ref<ReportCombo[]>([
        {
            rankLabel: '第一档',
            name: '物理 + 化学 + 生物',
            score: '89',
            theme: 'primary',
            metrics: [
                {label: 'A 维度', value: 33},
                {label: 'B 维度', value: 22},
                {label: 'C 维度', value: 2},
                {label: 'D 维度', value: 44},
            ],
            factorExplain: [
                '2gap 表示兴趣与能力之间的差异。',
                'AI：该组合在能力上较为均衡，兴趣略有倾向。',
                'AI：适合作为首选或备选方向。',
            ],
            recommendExplain: [
                '该组合有利于理科方向的长期发展。',
                '如果未来考虑工科、医科等方向，此组合较为匹配。',
            ],
        },
        {
            rankLabel: '第二档',
            name: '物理 + 化学 + 生物',
            score: '82',
            theme: 'blue',
            metrics: [
                {label: 'A 维度', value: 33},
                {label: 'B 维度', value: 22},
                {label: 'C 维度', value: 2},
                {label: 'D 维度', value: 44},
            ],
            factorExplain: [
                '2gap 表示兴趣与能力之间的差异。',
                'AI：该组合在能力上较为均衡，兴趣略有倾向。',
                'AI：适合作为首选或备选方向。',
            ],
            recommendExplain: [
                '该组合有利于理科方向的长期发展。',
                '如果未来考虑工科、医科等方向，此组合较为匹配。',
            ],
        },
        {
            rankLabel: '第三档',
            name: '物理 + 化学 + 生物',
            score: '78',
            theme: 'yellow',
            metrics: [
                {label: 'A 维度', value: 30},
                {label: 'B 维度', value: 20},
                {label: 'C 维度', value: 5},
                {label: 'D 维度', value: 40},
            ],
            factorExplain: [
                '2gap 稍大，说明兴趣与能力之间存在一定差异。',
                'AI：需要在学习投入上做出更多权衡。',
            ],
            recommendExplain: ['可作为备选方案，在目标不确定时保持灵活性。'],
        },
    ])

    // 3+1+2 模式下：以物理为主的 3 个组合（假数据，占位）
    const combos312Phy = ref<ReportCombo[]>([
        {
            rankLabel: '物理组·首选',
            name: '物理 + 化学 + 生物',
            score: '0.51',
            theme: 'primary',
            metrics: [
                {label: '主干阶段 S1', value: 0.48},
                {label: '辅科阶段 S23', value: 0.55},
                {label: '综合得分 S_final', value: 0.51},
                {label: '专业覆盖率', value: 0.96},
            ],
            factorExplain: [
                '示例：物理作为主干科目，兴趣与能力整合度较高，方向较为稳定。',
                '示例：化学、生物与物理在学习风格上协同性强，整体结构扎实。',
            ],
            recommendExplain: [
                '示例：适合作为 3+1+2 模式下以物理为主的首选组合，适合理科倾向明显的学生。',
            ],
        },
        {
            rankLabel: '物理组·备选一',
            name: '物理 + 化学 + 政治',
            score: '0.49',
            theme: 'blue',
            metrics: [
                {label: '主干阶段 S1', value: 0.48},
                {label: '辅科阶段 S23', value: 0.50},
                {label: '综合得分 S_final', value: 0.49},
                {label: '专业覆盖率', value: 0.99},
            ],
            factorExplain: [
                '示例：该组合在覆盖理工与政法相关专业方面更广，但辅科匹配度略有分化。',
            ],
            recommendExplain: [
                '示例：适合在理科倾向明确、同时希望保留部分文法类方向的学生作为备选方案。',
            ],
        },
        {
            rankLabel: '物理组·备选二',
            name: '物理 + 生物 + 政治',
            score: '0.48',
            theme: 'yellow',
            metrics: [
                {label: '主干阶段 S1', value: 0.48},
                {label: '辅科阶段 S23', value: 0.48},
                {label: '综合得分 S_final', value: 0.48},
                {label: '专业覆盖率', value: 0.85},
            ],
            factorExplain: [
                '示例：协同性较好，但辅科能力存在一定分化，对学习节奏要求更高。',
            ],
            recommendExplain: [
                '示例：适合作为灵活度较高的备选组合，需要在时间管理和科目平衡上多加注意。',
            ],
        },
    ])

    // 3+1+2 模式下：以历史为主的 3 个组合（假数据，占位）
    const combos312His = ref<ReportCombo[]>([
        {
            rankLabel: '历史组·首选',
            name: '历史 + 化学 + 生物',
            score: '0.37',
            theme: 'primary',
            metrics: [
                {label: '主干阶段 S1', value: 0.33},
                {label: '辅科阶段 S23', value: 0.42},
                {label: '综合得分 S_final', value: 0.37},
                {label: '专业覆盖率', value: 0.46},
            ],
            factorExplain: [
                '示例：历史作为主干科目，兴趣与能力整合度一般，需要一定补强。',
                '示例：化学、生物作为辅科在能力上有优势，但整体覆盖略窄。',
            ],
            recommendExplain: [
                '示例：适合作为偏向文科但仍希望保留部分理科基础时的首选方案。',
            ],
        },
        {
            rankLabel: '历史组·备选一',
            name: '历史 + 政治 + 生物',
            score: '0.35',
            theme: 'blue',
            metrics: [
                {label: '主干阶段 S1', value: 0.33},
                {label: '辅科阶段 S23', value: 0.38},
                {label: '综合得分 S_final', value: 0.35},
                {label: '专业覆盖率', value: 0.46},
            ],
            factorExplain: [
                '示例：文科协同性更好，但整体专业覆盖范围相对较集中。',
            ],
            recommendExplain: [
                '示例：适合明确文科兴趣、目标集中在部分人文社科方向的学生。',
            ],
        },
        {
            rankLabel: '历史组·备选二',
            name: '历史 + 化学 + 政治',
            score: '0.34',
            theme: 'yellow',
            metrics: [
                {label: '主干阶段 S1', value: 0.33},
                {label: '辅科阶段 S23', value: 0.37},
                {label: '综合得分 S_final', value: 0.34},
                {label: '专业覆盖率', value: 0.44},
            ],
            factorExplain: [
                '示例：辅科能力较均衡，但整体结构略松散，覆盖有限。',
            ],
            recommendExplain: [
                '示例：适合作为风险更可控的保守型文科组合备选。',
            ],
        },
    ])


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


    return {
        state,
        route,
        overview,
        businessType,
        aiLoading,
        truncatedLatestMessage,
        recommendedCombos,
        summaryCards,
        rawReportData,
        subjectRadar,
        isMode33,
        isMode312,
        combos312Phy,
        combos312His,
        mode33OverviewText,
        mode33View,
        mode312OverviewStrips,
    }
}
