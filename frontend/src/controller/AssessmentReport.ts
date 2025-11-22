import {ref, computed, onMounted} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useGlobalLoading} from '@/controller/useGlobalLoading'
import {useTestSession} from '@/controller/testSession'
import {apiRequest} from "@/api";
import {useAlert} from "@/controller/useAlert";

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

// 整体返回结构，对应 Go: EngineResult
export interface ReportResponse {
    sub_basic: CommonSection | null;
    radar: RadarData | null;
    recommend_33: Mode33Section | null;
    recommend_312: Mode312Section | null;
}

// ====== 通用部分：兴趣-能力整体特征 ======

export interface CommonSection {
    global_cosine: number;
    quality_score: number;
    subjects: SubjectProfileData[];
}

export interface SubjectProfileData {
    subject: string;        // "PHY" | "CHE" | "BIO" | "GEO" | "HIS" | "POL" ...
    interest_z: number;
    ability_z: number;
    zgap: number;
    ability_share: number;
    fit: number;
}

// ====== 雷达图数据 ======

export interface RadarData {
    subjects: string[];      // ["PHY","CHE","BIO","GEO","HIS","POL"]
    interest_pct: number[];  // [61, 60, 58, 55, 40, 39]
    ability_pct: number[];   // [100, 100, 100, 50, 44, 44]
}

// ====== 3+3 推荐部分 ======

export interface Mode33Section {
    top_combinations: Combo33CoreData[];
}

export interface Combo33CoreData {
    subjects: string[];   // 实际长度为 3
    avg_fit: number;
    min_ability: number;
    rarity: number;
    risk_penalty: number;
    score: number;
    combo_cosine: number;
}

// ====== 3+1+2 推荐部分 ======

export interface Mode312Section {
    anchor_phy: AnchorCoreData;
    anchor_his: AnchorCoreData;
}

export interface AnchorCoreData {
    subject: string;      // "PHY" 或 "HIS"
    fit: number;
    ability_norm: number;
    term_fit: number;
    term_ability: number;
    term_coverage: number;
    s1: number;
    combos: ComboCoreData[];
    s_final: number;
}

export interface ComboCoreData {
    aux1: string;
    aux2: string;
    avg_fit: number;
    min_fit: number;
    combo_cos: number;
    auxAbility: number;   // 注意：后端 JSON 字段是 "auxAbility"
    coverage: number;
    mix_penalty: number;
    s23: number;
    s_final_combo: number;
}


function getAiReportParam(publicID: string, businessTyp: string) {
    return apiRequest<ReportResponse>('/api/generate_report', {
        method: 'POST',
        body: {
            public_id: publicID,
            business_type: businessTyp
        },
    })
}


export function useReportPage() {
    const {showLoading, hideLoading} = useGlobalLoading()
    const route = useRoute()
    const {state} = useTestSession()
    const {showAlert} = useAlert()
    const router = useRouter()

    // 备用：当前业务类型（暂时模板里没用，将来如果需要接后端可以直接用）
    const businessType = computed(() => {
        return String(route.params.typ ?? state.businessType ?? '')
    })

    // 之后接入后端 / SSE 时再开启，这里先保持为 false
    const aiLoading = ref(false)
    const truncatedLatestMessage = ref<string[]>([])

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

    onMounted(async () => {
        showLoading("正在准备智能分析参数", 20_000)
        const public_id = state.recordPublicID
        if (!public_id) {
            showAlert('未找到试卷编号', () => {
                router.replace('/').then()
            })
            return
        }
        try {
            const resp = await getAiReportParam(public_id, businessType.value)
            console.log("------>>>resp data:", resp)
        } catch (e) {
            showAlert("生成 AI 参数失败:"+e);
        } finally {
            hideLoading();
        }
    })

    return {
        route,
        businessType,
        aiLoading,
        truncatedLatestMessage,
        recommendedCombos,
        summaryCards,
    }
}
