import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useGlobalLoading } from '@/controller/useGlobalLoading'
import { useTestSession } from '@/controller/testSession'
import { StageReport, pushStageRoute } from '@/controller/common'

/**
 * 推荐组合中单个指标，如 A 维度 / B 维度 等
 */
export interface ComboMetric {
    label: string
    value: number
}

/**
 * 报告里的一档 / 二档 / 三档组合
 */
export interface ReportCombo {
    rankLabel: string
    name: string
    score: string
    theme: string
    metrics: ComboMetric[]
    factorExplain: string[]
    recommendExplain: string[]
}

/**
 * 底部总结卡片
 */
export interface SummaryCard {
    title: string
    content: string
}

export function useReportPage() {
    const { showLoading, hideLoading } = useGlobalLoading()

    const route = useRoute()
    const router = useRouter()
    const { state } = useTestSession()

    // 当前业务类型：优先用路由 :typ，其次用全局 store
    const businessType = computed(() => {
        return String(route.params.typ ?? state.businessType ?? '')
    })

    // 步骤条：文案列表（“基础信息 / 兴趣测试 / 能力测试 / 报告”）
    const stepItems = computed(() => {
        const flow = state.testRoutes ?? []
        return flow.map(step => step.title)
    })

    // 当前步骤在流程中的 index（report 所在的位置）
    const currentStepIndex = computed(() => {
        const flow = state.testRoutes ?? []
        const idx = flow.findIndex(step => step.stage === StageReport)
        return idx >= 0 ? idx : 0
    })

    /**
     * 点击步骤条的某一步时的行为：
     * - 只允许回退（targetIndex <= currentStepIndex）
     * - 当前已经是报告页，再点“报告”就不动
     * - 其它阶段统一用 pushStageRoute 跳转
     */
    function handleStepClick(targetIndex: number) {
        const flow = state.testRoutes ?? []
        if (!flow.length) return

        const curIdx = currentStepIndex.value
        // 不允许点到还没解锁的未来步骤
        if (targetIndex > curIdx) return

        const target = flow[targetIndex]
        if (!target) return

        // 已经在报告页，再点“报告”就不跳了
        if (target.stage === StageReport) {
            return
        }

        const biz = businessType.value
        if (!biz) return

        pushStageRoute(router, biz, target.stage)
    }

    // 之后接入后端 / SSE 时再开启，这里先保持为 false
    const aiLoading = ref(false)

    // AI 生成过程中的日志（当前主要给覆盖层占位用）
    const truncatedLatestMessage = ref<string[]>([])

    // 临时静态数据：推荐组合
    const recommendedCombos = ref<ReportCombo[]>([
        {
            rankLabel: '第一档',
            name: '物理 + 化学 + 生物',
            score: '89',
            theme: 'primary', // 第一档：紫色
            metrics: [
                { label: 'A 维度', value: 33 },
                { label: 'B 维度', value: 22 },
                { label: 'C 维度', value: 2 },
                { label: 'D 维度', value: 44 },
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
                { label: 'A 维度', value: 33 },
                { label: 'B 维度', value: 22 },
                { label: 'C 维度', value: 2 },
                { label: 'D 维度', value: 44 },
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
                { label: 'A 维度', value: 30 },
                { label: 'B 维度', value: 20 },
                { label: 'C 维度', value: 5 },
                { label: 'D 维度', value: 40 },
            ],
            factorExplain: [
                '2gap 稍大，说明兴趣与能力之间存在一定差异。',
                'AI：需要在学习投入上做出更多权衡。',
            ],
            recommendExplain: ['可作为备选方案，在目标不确定时保持灵活性。'],
        },
    ])

    // 临时静态数据：报告总结卡片
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

    return {
        // 路由 & 步骤条
        route,
        stepItems,
        currentStepIndex,
        handleStepClick,

        // AI 加载状态（暂时占位）
        aiLoading,
        truncatedLatestMessage,

        // 报告内容
        recommendedCombos,
        summaryCards,
    }
}
