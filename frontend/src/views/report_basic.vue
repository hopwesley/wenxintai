<template>
  <TestLayout :key="route.fullPath">
    <template #header>
      <StepIndicator/>
    </template>

    <main class="report-page" ref="reportPageRoot">
      <!-- 顶部：精简版报告抬头 + 基础信息 -->
      <section class="report-card report-card--overview">
        <header class="report-card__header">
          <div class="report-card__title-row">
            <h1 class="report-card__title">智能选科精简报告</h1>
            <span v-if="overview.mode" class="report-card__mode-pill">
              {{ overview.mode }}
            </span>
          </div>
        </header>

        <div class="report-card__divider"></div>

        <section class="report-section report-section--profile">
          <h2 class="report-section__title">学生基础信息</h2>

          <div class="report-profile">
            <div class="report-profile-grid report-profile-grid--compact">
              <div class="report-field">
                <span class="report-field__label">问心台账号</span>
                <span class="report-field__value">
                  {{ overview.account || '——' }}
                </span>
              </div>
              <div class="report-field">
                <span class="report-field__label">学生号</span>
                <span class="report-field__value">
                  {{ overview.studentNo || '——' }}
                </span>
              </div>
              <div class="report-field">
                <span class="report-field__label">学校</span>
                <span class="report-field__value">
                  {{ overview.schoolName || '——' }}
                </span>
              </div>
              <div class="report-field">
                <span class="report-field__label">所在地区</span>
                <span class="report-field__value">
                  {{ overview.studentLocation || '——' }}
                </span>
              </div>

              <div class="report-field">
                <span class="report-field__label">生成日期</span>
                <span class="report-field__value">
                  {{ overview.generateDate || '——' }}
                </span>
              </div>
              <div class="report-field">
                <span class="report-field__label">有效期至</span>
                <span class="report-field__value">
                  {{ overview.expireDate || '——' }}
                </span>
              </div>
            </div>
          </div>
        </section>
      </section>

      <!-- 核心选科建议（精简版） -->
      <section class="report-card report-card--recommendation">
        <header class="report-card__header">
          <div class="report-card__title-row">
            <h2 class="report-card__title">核心选科建议（精简版）</h2>
          </div>
        </header>

        <div class="report-card__divider"></div>

        <!-- 3+3 模式：只展示前两档组合 + 简短 AI 文案 -->
        <section v-if="isMode33" class="report-section report-section--combos">
          <div v-if="primaryCombo33" class="combo-block">
            <div class="combo-rank-strip combo-rank-strip--primary">
              <span class="combo-rank-strip__label">
                首选组合：{{ primaryCombo33.name }}
              </span>
              <span class="combo-rank-strip__score">
                综合得分：{{ primaryCombo33.score }}
              </span>
            </div>

            <div class="combo-panel combo-panel--compact">
              <section class="combo-panel__section">
                <h5 class="combo-panel__subtitle">组合整体说明</h5>
                <div class="combo-panel__ai-block">
                  <p class="combo-panel__text-line">
                    {{ primaryComboExplainText }}
                  </p>
                </div>
              </section>
            </div>
          </div>

          <div
              v-if="secondaryCombo33"
              class="combo-block combo-block--secondary"
          >
            <div class="combo-rank-strip combo-rank-strip--blue">
              <span class="combo-rank-strip__label">
                备选组合：{{ secondaryCombo33.name }}
              </span>
              <span class="combo-rank-strip__score">
                综合得分：{{ secondaryCombo33.score }}
              </span>
            </div>

            <div class="combo-panel combo-panel--compact">
              <section class="combo-panel__section">
                <h5 class="combo-panel__subtitle">组合整体说明</h5>
                <div class="combo-panel__ai-block">
                  <p class="combo-panel__text-line">
                    {{ secondaryComboExplainText }}
                  </p>
                </div>
              </section>
            </div>
          </div>

          <div v-if="finalReport" class="recommend-main-strip recommend-main-strip--compact">
            <span class="recommend-main-strip__label">
              {{ finalReport.strategic_conclusion }}
            </span>
          </div>
        </section>

        <!-- 3+1+2 模式：给出物理组 / 历史组的整体建议 + 最推荐组合 -->
        <section v-else-if="isMode312" class="report-section report-section--combos">
          <div v-if="bestCombo312" class="combo-block">
            <div class="combo-rank-strip combo-rank-strip--primary">
              <span class="combo-rank-strip__label">
                首选方向：{{ bestCombo312.groupLabel }} · {{ bestCombo312.combo.name }}
              </span>
              <span class="combo-rank-strip__score">
                综合得分：{{ bestCombo312.combo.score }}
              </span>
            </div>

            <div class="combo-panel combo-panel--compact">
              <section class="combo-panel__section">
                <h5 class="combo-panel__subtitle">方向整体说明</h5>
                <div class="combo-panel__ai-block">
                  <p class="combo-panel__text-line">
                    {{ bestCombo312.overviewText }}
                  </p>
                </div>
              </section>

              <section class="combo-panel__section">
                <h5 class="combo-panel__subtitle">AI 选科建议</h5>
                <div class="combo-panel__ai-block">
                  <p class="combo-panel__text-line">
                    {{ bestCombo312AdviceText }}
                  </p>
                </div>
              </section>
            </div>
          </div>

          <div v-if="finalReport" class="recommend-main-strip recommend-main-strip--compact">
            <span class="recommend-main-strip__label">
              {{ finalReport.strategic_conclusion }}
            </span>
          </div>
        </section>

        <section v-else class="report-section">
          <p>当前报告模式暂不支持精简版核心建议展示。</p>
        </section>
      </section>

      <!-- 学科兴趣与能力概览（精简版） -->
      <section class="report-card">
        <header class="report-card__header">
          <div class="report-card__title-row">
            <h2 class="report-card__title">学科兴趣与能力概览（精简版）</h2>
          </div>
        </header>

        <div class="report-card__divider"></div>

        <section class="report-section report-section--basic-analysis">
          <article class="analysis-interpretation analysis-interpretation--compact">
            <div class="analysis-interpretation__header">
              <span class="analysis-interpretation__title">整体匹配度与数据质量</span>
            </div>
            <p class="ai-text-block__line">
              <span class="ai-text-block__label">整体匹配度：</span>
              <span class="ai-text-block__value">
                {{ formatZ(globalCosine) }}
              </span>
              <span class="ai-text-block__hint">
                （-1.0 表示完全不匹配，1.0 表示完全匹配）
              </span>
            </p>
            <p class="ai-text-block__line">
              <span class="ai-text-block__label">数据质量评分：</span>
              <span class="ai-text-block__value">
                {{ formatZ(qualityScore) }}
              </span>
              <span class="ai-text-block__hint">
                （高于 0.4 表示答题可信）
              </span>
            </p>
            <p class="analysis-interpretation__text">
              {{ commonSectionText }}
            </p>
          </article>

          <div class="basic-analysis-layout basic-analysis-layout--compact">
            <div class="basic-analysis-layout__chart basic-analysis-layout__chart--radar">
              <SubjectRadarChart
                  v-if="subjectRadar"
                  :radar="subjectRadar"
              />
              <div v-else class="chart-placeholder chart-placeholder--radar">
                雷达图暂无数据
              </div>
            </div>

            <div class="basic-analysis-layout__table">
              <h3 class="report-section__subtitle">六科综合适配度（精简版）</h3>
              <div class="report-table-wrapper">
                <table class="report-table">
                  <thead>
                  <tr>
                    <th class="report-table__cell report-table__cell--head report-table__cell--subject">
                      学科
                    </th>
                    <th class="report-table__cell report-table__cell--head">
                      兴趣 z
                    </th>
                    <th class="report-table__cell report-table__cell--head">
                      能力 z
                    </th>
                    <th class="report-table__cell report-table__cell--head">
                      匹配度 fit
                    </th>
                  </tr>
                  </thead>
                  <tbody v-if="subjectRows.length">
                  <tr v-for="row in subjectRows" :key="row.code">
                    <td class="report-table__cell report-table__cell--subject">
                      {{ row.label }}
                    </td>
                    <td class="report-table__cell">
                      {{ formatZ(row.interest_z) }}
                    </td>
                    <td class="report-table__cell">
                      {{ formatZ(row.ability_z) }}
                    </td>
                    <td class="report-table__cell">
                      {{ formatZ(row.fit) }}
                    </td>
                  </tr>
                  </tbody>
                  <tbody v-else>
                  <tr>
                    <td class="report-table__cell" colspan="4">
                      暂无基础参数数据
                    </td>
                  </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>
      </section>

      <!-- 报告摘要（精简版） -->
      <section class="report-card">
        <header class="report-card__header">
          <div class="report-card__title-row">
            <h2 class="report-card__title">报告摘要（精简版）</h2>
          </div>
        </header>

        <div class="report-card__divider"></div>

        <section class="report-section report-section--summary">
          <div v-if="finalReport" class="summary-grid">
            <article class="summary-card">
              <h4 class="summary-card__title">总体结论</h4>
              <p>{{ finalReport.core_trends }}</p>
            </article>

            <article class="summary-card">
              <h4 class="summary-card__title">学习提醒</h4>
              <p>{{ finalReport.risk_diagnosis }}</p>
            </article>

            <article class="summary-card">
              <h4 class="summary-card__title">未来方向</h4>
              <p>{{ finalReport.mode_strategy }}</p>
            </article>
          </div>

          <div v-else class="summary-grid">
            <article class="summary-card">
              <p>AI 摘要尚未生成或加载失败。</p>
            </article>
          </div>
        </section>
      </section>
    </main>

    <div class="report-page__actions">
      <button
          class="btn btn-secondary report-page__action"
          @click="handleBackToHome"
      >
        返回测试首页
      </button>

      <button class="btn btn-primary report-page__action" @click="handleExportPdf">
        导出 PDF
      </button>
    </div>

    <AiGeneratingOverlay
        v-if="aiLoading"
        title="AI 正在为你生成精简报告…"
        subtitle="正在分析你的测试各项参数，为你生成精简版选科报告"
        :log-lines="truncatedLatestMessage"
        :meta="{
        mode: overview.mode || '',
        grade: state.grade || '',
        stage: '选科报告（精简版）'
      }"
    />
  </TestLayout>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import StepIndicator from '@/views/components/StepIndicator.vue'
import TestLayout from '@/views/components/TestLayout.vue'
import AiGeneratingOverlay from '@/views/components/AiGeneratingOverlay.vue'
import SubjectRadarChart from '@/views/components/SubjectRadarChart.vue'
import {useReportPage, ReportSubjectScore} from '@/controller/AssessmentReport'
import {aiReportData} from '@/controller/AssessmentReport'
import {subjectLabelMap} from '@/controller/common'
import html2pdf from 'html2pdf.js'

const {
  state,
  route,
  overview,
  aiLoading,
  truncatedLatestMessage,
  subjectRadar,
  rawReportData,
  isMode33,
  isMode312,
  mode33View,
  mode312OverviewStrips,
  finalReport,
  handleBackToHome,
  reportPageRoot,
  handleExportPdf,
} = useReportPage()

// --------- 3+3：首选 & 备选组合（只取前两档） ----------
const primaryCombo33 = computed(() => {
  if (!isMode33.value || !mode33View.value) return null
  return mode33View.value.topCombos[0] || null
})

const secondaryCombo33 = computed(() => {
  if (!isMode33.value || !mode33View.value) return null
  return mode33View.value.topCombos[1] || null
})

const primaryComboExplainText = computed(() => {
  if (!primaryCombo33.value) return '暂无组合说明。'
  return (
      primaryCombo33.value.recommendAdvice ||
      primaryCombo33.value.recommendExplain ||
      '该组合在兴趣、能力与风险之间取得了较好的平衡，适合作为当前阶段的首选方案。'
  )
})

const secondaryComboExplainText = computed(() => {
  if (!secondaryCombo33.value) return '暂无组合说明。'
  return (
      secondaryCombo33.value.recommendAdvice ||
      secondaryCombo33.value.recommendExplain ||
      '该组合整体结构较稳，可作为在学校开课条件与个人感受之间折中的备选方案。'
  )
})

// --------- 3+1+2：物理组 / 历史组首选组合，选一个作为“首选方向” ----------
const bestCombo312 = computed(() => {
  if (!isMode312.value || !mode312OverviewStrips.value) return null
  const strips = mode312OverviewStrips.value

  const phyCombo = strips.phyTopCombos[0]
  const hisCombo = strips.hisTopCombos[0]

  if (!phyCombo && !hisCombo) return null

  // 用 s1 作为“主干阶段得分”的比较基础；若缺失则退回首选组合分数
  const phyScore = strips.phyS1 || (phyCombo ? Number(phyCombo.score) || 0 : 0)
  const hisScore = strips.hisS1 || (hisCombo ? Number(hisCombo.score) || 0 : 0)

  const usePhy = phyScore >= hisScore

  if (usePhy && phyCombo) {
    return {
      groupLabel: '物理组',
      combo: phyCombo,
      overviewText: strips.phyOverviewText || '',
    }
  }
  if (!usePhy && hisCombo) {
    return {
      groupLabel: '历史组',
      combo: hisCombo,
      overviewText: strips.hisOverviewText || '',
    }
  }
  return null
})

const bestCombo312AdviceText = computed(() => {
  if (!bestCombo312.value) return '暂无选科建议。'

  const combo = bestCombo312.value.combo
  if (combo.recommendAdvice) return combo.recommendAdvice
  if (combo.recommendExplain) return combo.recommendExplain

  return '该方向在主干科目与两门辅科之间整体结构较为稳健，可作为当前阶段的主要发展方向。'
})

// --------- 学科表：从 rawReportData.common_score.common.subjects 精简提取 ----------
const subjectRows = computed(() => {
  const subjects = rawReportData.value?.common_score?.common?.subjects || []
  return (subjects as ReportSubjectScore[]).map((s) => ({
    code: s.subject,
    label: subjectLabelMap[s.subject] ?? s.subject,
    interest_z: s.interest_z,
    ability_z: s.ability_z,
    fit: s.fit,
  }))
})

// 整体匹配度 & 数据质量
const globalCosine = computed(
    () => rawReportData.value?.common_score?.common?.global_cosine ?? null,
)
const qualityScore = computed(
    () => rawReportData.value?.common_score?.common?.quality_score ?? null,
)

// AI common_section 文本（没有就用 finalReport 兜底）
const commonSectionText = computed(() => {
  const commonSection = aiReportData.value?.common_section
  if (commonSection?.report_validity_text) {
    return commonSection.report_validity_text
  }
  if (finalReport.value?.report_validity) {
    return finalReport.value.report_validity
  }
  return '本次测评数据可信度尚可，可作为当前阶段选科决策的重要参考。'
})

// --------- 工具函数 ----------
function formatZ(v: number | null | undefined): string {
  if (v === null || v === undefined || Number.isNaN(v)) return '--'
  return v.toFixed(3)
}

</script>

<style scoped src="@/styles/assessment-report.css"></style>

<style scoped>
.report-profile-grid--compact {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.analysis-interpretation--compact {
  margin-bottom: 12px;
}

.basic-analysis-layout--compact {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 1.2fr);
  gap: 16px;
  align-items: flex-start;
}

.basic-analysis-layout__table {
  margin-top: 4px;
}

.combo-panel--compact {
  margin-top: 8px;
}

.combo-block--secondary {
  margin-top: 16px;
}

.recommend-main-strip--compact {
  margin-top: 16px;
}

.ai-text-block__hint {
  font-size: 12px;
  color: #6b7280;
  margin-left: 4px;
}
</style>
