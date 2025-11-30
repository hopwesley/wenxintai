<template>
  <TestLayout :key="route.fullPath">
    <template #header>
      <StepIndicator/>
    </template>

    <main class="report-page" ref="reportPageRoot">
      <!-- 顶部：报告概览卡片 -->
      <section class="report-card report-card--overview">
        <header class="report-card__header">
          <div class="report-card__title-row">
            <h1 class="report-card__title">智能选科分析报告</h1>
            <span v-if="overview.mode" class="report-card__mode-pill">
            {{ overview.mode }}
          </span>
          </div>
          <!-- 将来可以放一个“生成新报告 / 导出 PDF”之类的小按钮 -->
        </header>
        <div class="report-card__divider"></div>
        <!-- 1. 个人资料 / 报告基础信息 -->
        <section class="report-section report-section--profile">
          <h2 class="report-section__title">学生基础信息</h2>

          <div class="report-profile">
            <div class="report-profile-grid">
              <div class="report-field">
                <span class="report-field__label">问心台账号</span>
                <span class="report-field__value">
                {{ overview.account || '——' }}
              </span>
              </div>
              <div class="report-field">
                <span class="report-field__label">学号</span>
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
              <!-- 占位单元，让 4x2 网格视觉更均衡 -->
              <div class="report-field report-field--placeholder"></div>
              <div class="report-field report-field--placeholder"></div>
            </div>
          </div>
        </section>
        <div class="report-card__divider"></div>
        <header class="report-card__header">
          <div class="report-card__title-row">
            <h2 class="report-card__title">选科组合推荐</h2>
          </div>
        </header>


        <div v-if="isMode33">
          <section class="report-section report-section--recommend-analysis">
            <h3 class="report-section__title">整体推荐概览（3+3 模式）</h3>

            <div class="recommend-analysis-layout">
              <!-- 左侧：三种组合综合得分柱状图（用 VChart） -->
              <div class="recommend-analysis__chart">
                <ComboScoreChart
                    v-if="mode33View && mode33View.chartCombos.length"
                    :combos="mode33View.chartCombos"
                />
                <div v-else class="chart-placeholder">
                  推荐组合整体分布概览暂无数据
                </div>
              </div>

              <div class="recommend-analysis__chart">
                <ComboScoreChart
                    v-if="mode33View && mode33View.rarityRiskPairs.length"
                    :combos="mode33View.rarityRiskPairs"
                />
                <div v-else class="chart-placeholder">
                  推荐组合 稀有度 &amp; 风险 暂无数据
                </div>
              </div>
            </div>

            <div v-if="mode33View && mode33View.overviewText" class="recommend-main-strip">
  <span class="recommend-main-strip__label">
    {{ mode33View.overviewText }}
  </span>
            </div>

          </section>
          <!-- 6. 分档组合列表（3+3：仍然用原来的 recommendedCombos） -->
          <section class="report-section report-section--combos">
            <h3 class="report-section__title">分档组合详情（3+3 模式）</h3>

            <div
                v-for="combo in mode33View?.topCombos"
                :key="combo.rankLabel + combo.name"
                class="combo-block"
            >
              <div
                  class="combo-rank-strip"
                  :class="{
            'combo-rank-strip--primary': combo.theme === 'primary',
            'combo-rank-strip--blue': combo.theme === 'blue',
            'combo-rank-strip--yellow': combo.theme === 'yellow',
          }"
              >
          <span class="combo-rank-strip__label">
            {{ combo.rankLabel }}：{{ combo.name }}
          </span>
                <span class="combo-rank-strip__score">
            综合得分：{{ combo.score }}
          </span>
              </div>
              <div class="combo-panel">
                <header class="combo-panel__header">
                  <h4 class="combo-panel__title">结构指标概览</h4>
                </header>

                <div class="combo-metrics">
                  <div
                      v-for="metric in combo.metrics"
                      :key="metric.label"
                      class="combo-metrics__item"
                  >
              <span class="combo-metrics__label">
                {{ metric.label }}
              </span>
                    <span class="combo-metrics__value">
                {{ metric.value }}
              </span>
                  </div>
                </div>

                <section class="combo-panel__section">
                  <h5 class="combo-panel__subtitle">AI 推荐评语</h5>
                  <div class="combo-panel__ai-block">
                    <p class="combo-panel__text-line">
                      {{ combo.recommendExplain }}
                    </p>
                  </div>
                </section>

                <section class="combo-panel__section">
                  <h5 class="combo-panel__subtitle">AI 选科建议</h5>
                  <div class="combo-panel__ai-block">
                    <p class="combo-panel__text-line">
                      {{ combo.recommendAdvice }}
                    </p>
                  </div>
                </section>
              </div>
            </div>
          </section>
        </div>
        <!-- ===================== 3+1+2 模式：物理组 + 历史组 ===================== -->
        <div v-else-if="isMode312">
          <section class="report-section report-section--mode312-intro">
            <h3 class="report-section__title">物理 / 历史方向说明</h3>
            <p class="mode312-intro__text">
              在当前 3+1+2 模式下，全国范围内理工类专业及招生计划数量通常多于文史类专业，
              因此从长期升学机会和专业选择空间来看，选择物理往往会获得更宽的专业覆盖范围。
            </p>
            <p class="mode312-intro__text mode312-intro__text--secondary">
              为了避免把「选物理还是选历史」这个关键决策完全交给算法，本系统的做法是：
              <strong>不直接替你做出文理方向的决定</strong>，
              而是分别给出以物理为主、以历史为主的 3 个代表性组合，
              由你和家长结合兴趣、目标专业和学校情况做最终选择。
            </p>
          </section>
          <!-- ===================== 3+1+2 · 物理组 ===================== -->
          <section class="report-section report-section--mode312-group">
            <!-- 物理组标题 + 概述 -->
            <header class="recommend-312-group__header">
              <h4 class="recommend-312-group__title">以物理为主的 3+1+2 组合</h4>
            </header>

            <!-- 物理组：2 张图表占位，展示 3 个组合的得分情况 -->
            <section
                class="report-section report-section--recommend-analysis report-section--mode312-analysis"
            >
              <h5 class="report-section__subtitle">物理组整体推荐概览</h5>
              <div class="recommend-analysis-layout">
                <div class="recommend-analysis__chart">
                  <ComboScoreChart
                      v-if="mode312OverviewStrips && mode312OverviewStrips.phyScoreBars.length"
                      :combos="mode312OverviewStrips.phyScoreBars"
                  />
                  <div v-else class="chart-placeholder">
                    推荐组合整体分布概览暂无数据
                  </div>
                </div>
                <div class="recommend-analysis__chart">
                  <ComboScoreChart
                      v-if="mode312OverviewStrips && mode312OverviewStrips.phyCoverageRiskBars.length"
                      :combos="mode312OverviewStrips.phyCoverageRiskBars"
                  />
                  <div v-else class="chart-placeholder">
                    推荐组合覆盖率/风险暂无数据
                  </div>
                </div>
              </div>

              <div v-if="mode312OverviewStrips" class="recommend-main-strip">
                <span class="recommend-main-strip__label">  {{ mode312OverviewStrips.phyOverviewText }}</span>
              </div>

            </section>

            <!-- 物理组：3 个组合卡片，复用原来的 combo-block 结构 -->
            <section class="report-section report-section--combos report-section--mode312-combos">
              <h5 class="report-section__subtitle">以物理为主的分档组合详情</h5>

              <div
                  v-for="combo in mode312OverviewStrips?.phyTopCombos"
                  :key="'PHY-' + combo.rankLabel + combo.name"
                  class="combo-block"
              >
                <div
                    class="combo-rank-strip"
                    :class="{
              'combo-rank-strip--primary': combo.theme === 'primary',
              'combo-rank-strip--blue': combo.theme === 'blue',
              'combo-rank-strip--yellow': combo.theme === 'yellow',
            }"
                >
            <span class="combo-rank-strip__label">
              {{ combo.rankLabel }}：{{ combo.name }}
            </span>
                  <span class="combo-rank-strip__score">
              综合得分：{{ combo.score }}
            </span>
                </div>

                <div class="combo-panel">
                  <header class="combo-panel__header">
                    <h4 class="combo-panel__title">结构指标概览</h4>
                  </header>

                  <div class="combo-metrics">
                    <div
                        v-for="metric in combo.metrics"
                        :key="metric.label"
                        class="combo-metrics__item"
                    >
                <span class="combo-metrics__label">
                  {{ metric.label }}
                </span>
                      <span class="combo-metrics__value">
                  {{ metric.value }}
                </span>
                    </div>
                  </div>

                  <section class="combo-panel__section">
                    <h5 class="combo-panel__subtitle">影响因素解读</h5>
                    <div class="combo-panel__ai-block">
                      <p class="combo-panel__text-line">
                        {{ combo.recommendExplain }}
                      </p>
                    </div>
                  </section>

                  <section class="combo-panel__section">
                    <h5 class="combo-panel__subtitle">AI 选科建议</h5>
                    <div class="combo-panel__ai-block">
                      <p class="combo-panel__text-line">
                        {{ combo.recommendAdvice }}
                      </p>
                    </div>
                  </section>
                </div>
              </div>
            </section>
          </section>
          <!-- ===================== 3+1+2 · 历史组 ===================== -->
          <section class="report-section report-section--mode312-group">
            <!-- 历史组标题 + 概述 -->
            <header class="recommend-312-group__header">
              <h4 class="recommend-312-group__title">以历史为主的 3+1+2 组合</h4>
            </header>

            <!-- 历史组：2 张图表占位 -->
            <section
                class="report-section report-section--recommend-analysis report-section--mode312-analysis"
            >
              <h5 class="report-section__subtitle">历史组整体推荐概览</h5>
              <div class="recommend-analysis-layout">
                <div class="recommend-analysis__chart">
                  <ComboScoreChart
                      v-if="mode312OverviewStrips && mode312OverviewStrips.hisScoreBars.length"
                      :combos="mode312OverviewStrips.hisScoreBars"
                  />
                  <div v-else class="chart-placeholder">
                    推荐组合整体分布概览暂无数据
                  </div>
                </div>
                <div class="recommend-analysis__chart">
                  <ComboScoreChart
                      v-if="mode312OverviewStrips && mode312OverviewStrips.hisCoverageRiskBars.length"
                      :combos="mode312OverviewStrips.hisCoverageRiskBars"
                  />
                  <div v-else class="chart-placeholder">
                    推荐组合覆盖率/风险暂无数据
                  </div>
                </div>
              </div>

              <div v-if="mode312OverviewStrips" class="recommend-main-strip">
                <span class="recommend-main-strip__label"> {{ mode312OverviewStrips.hisOverviewText }} </span>
              </div>

            </section>

            <!-- 历史组：3 个组合卡片 -->
            <section class="report-section report-section--combos report-section--mode312-combos">
              <h5 class="report-section__subtitle">以历史为主的分档组合详情</h5>

              <div
                  v-for="combo in mode312OverviewStrips?.hisTopCombos"
                  :key="'HIS-' + combo.rankLabel + combo.name"
                  class="combo-block"
              >
                <div
                    class="combo-rank-strip"
                    :class="{
              'combo-rank-strip--primary': combo.theme === 'primary',
              'combo-rank-strip--blue': combo.theme === 'blue',
              'combo-rank-strip--yellow': combo.theme === 'yellow',
            }"
                >
            <span class="combo-rank-strip__label">
              {{ combo.rankLabel }}：{{ combo.name }}
            </span>
                  <span class="combo-rank-strip__score">
              得分：{{ combo.score }}
            </span>
                </div>

                <div class="combo-panel">
                  <header class="combo-panel__header">
                    <h4 class="combo-panel__title">结构指标概览</h4>
                  </header>

                  <div class="combo-metrics">
                    <div
                        v-for="metric in combo.metrics"
                        :key="metric.label"
                        class="combo-metrics__item"
                    >
                <span class="combo-metrics__label">
                  {{ metric.label }}
                </span>
                      <span class="combo-metrics__value">
                  {{ metric.value }}
                </span>
                    </div>
                  </div>

                  <section class="combo-panel__section">
                    <h5 class="combo-panel__subtitle">影响因素解读</h5>
                    <div class="combo-panel__ai-block">
                      <p class="combo-panel__text-line">
                        {{ combo.recommendExplain }}
                      </p>
                    </div>
                  </section>

                  <section class="combo-panel__section">
                    <h5 class="combo-panel__subtitle">AI 选科建议</h5>
                    <div class="combo-panel__ai-block">
                      <p class="combo-panel__text-line">
                        {{ combo.recommendAdvice }}
                      </p>
                    </div>
                  </section>
                </div>
              </div>
            </section>
          </section>
        </div>

        <!-- 2. 基础分析：雷达 + 柱状图 + 两块简短解读 -->
        <section class="report-section report-section--basic-analysis">
          <h2 class="report-section__title">基础能力与兴趣结构</h2>

          <div>
            <article class="analysis-interpretation">
              <div class="analysis-interpretation__header">
                <span class="analysis-interpretation__title">整体匹配度解读</span>
              </div>
              <p class="ai-text-block__line">
                <span class="ai-text-block__label">整体匹配度：</span>
                <span class="ai-text-block__value"> {{ rawReportData?.common_score.common.global_cosine_score }}  </span>
                <span>(0–100 标准分，数值越高表示兴趣与能力整体方向越一致)</span>
              </p>
              <p class="ai-text-block__line">
                <span class="ai-text-block__label">数据质量评分：</span>
                <span class="ai-text-block__value"> {{ rawReportData?.common_score.common.quality_score_score }} </span>
                <span>(0–100 标准分，约高于 40 分表示本次答题质量较可信)</span>
              </p>
              <p class="analysis-interpretation__text">
                {{ aiReportData?.common_section?.report_validity_text }}
              </p>
            </article>
          </div>

          <div class="basic-analysis-layout__chart basic-analysis-layout__chart--radar">
            <SubjectRadarChart
                v-if="subjectRadar"
                :radar="subjectRadar"
            />
            <div
                v-else
                class="chart-placeholder chart-placeholder--radar"
            >
              雷达图暂无数据
            </div>
          </div>

          <div class="basic-analysis-layout__chart basic-analysis-layout__chart--bar">
            <SubjectAbilityBarChart
                v-if="rawReportData && rawReportData.common_score && rawReportData.common_score.common"
                :subjects="rawReportData.common_score.common.subjects"
            />
            <div v-else class="chart-placeholder">
              基础能力柱状图暂无数据
            </div>
          </div>
        </section>
        <section>
          <article class="analysis-interpretation">
            <div class="analysis-interpretation__header">
              <span class="analysis-interpretation__title">能力/兴趣结构综述</span>
            </div>
            <p class="analysis-interpretation__text">
              {{ aiReportData?.common_section?.subjects_summary_text }}
            </p>
          </article>
        </section>
      </section>
      <section class="report-section report-section--summary">
        <h3 class="report-section__title">报告摘要</h3>
        <div
            v-if="finalReport"
            class="summary-grid"
        >
          <article class="summary-card">
            <h4 class="summary-card__title">数据可信度与报告有效性</h4>
            <p>{{ finalReport.report_validity }}</p>
          </article>

          <article class="summary-card">
            <h4 class="summary-card__title">核心趋势概览</h4>
            <p>{{ finalReport.core_trends }}</p>
          </article>

          <article class="summary-card">
            <h4 class="summary-card__title">选科模式与组合策略</h4>
            <p>{{ finalReport.mode_strategy }}</p>
          </article>

          <article class="summary-card">
            <h4 class="summary-card__title">给学生的话</h4>
            <p>{{ finalReport.student_view }}</p>
          </article>

          <article class="summary-card">
            <h4 class="summary-card__title">给家长的建议</h4>
            <p>{{ finalReport.parent_view }}</p>
          </article>

          <article class="summary-card">
            <h4 class="summary-card__title">风险诊断与应对方向</h4>
            <p>{{ finalReport.risk_diagnosis }}</p>
          </article>

          <article class="summary-card">
            <h4 class="summary-card__title">整体选科建议</h4>
            <p>{{ finalReport.strategic_conclusion }}</p>
          </article>
        </div>
        <div v-else class="summary-grid">
          <article class="summary-card">
            <p>无总结报告可显示</p>
          </article>
        </div>
      </section>
    </main>

    <div class="report-page__actions">
      <button
          class="btn btn-secondary report-page__action"
          @click="handleBackToHome">
        返回测试首页
      </button>

      <button class="btn btn-primary report-page__action" @click="handleExportPdf">
        导出 PDF
      </button>
    </div>
    <AiGeneratingOverlay
        v-if="aiLoading"
        title="AI 正在为你生成专属报告…"
        subtitle="正在分析你的测试各项参数，为您全面展示智能分析结果"
        :log-lines="truncatedLatestMessage"
        :meta="{
    mode: overview.mode || '',
    grade: state.grade || '',
    stage: '选科报告'
  }"
    />
  </TestLayout>
</template>

<script setup lang="ts">
import StepIndicator from '@/views/components/StepIndicator.vue'
import TestLayout from '@/views/components/TestLayout.vue'
import AiGeneratingOverlay from '@/views/components/AiGeneratingOverlay.vue'
import {useReportPage} from '@/controller/AssessmentReport'
import SubjectRadarChart from "@/views/components/SubjectRadarChart.vue";
import SubjectAbilityBarChart from '@/views/components/SubjectAbilityBarChart.vue'
import ComboScoreChart from '@/views/components/ComboScoreChart.vue'
import {aiReportData} from '@/controller/AssessmentReport'

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


// z 值格式化：保留 2 位小数，空值显示 --
function formatZ(v: number | null | undefined): string {
  if (v === null || v === undefined || Number.isNaN(v)) return '--'
  return v.toFixed(2)
}

// 百分比格式化：0~1 -> 0.0%
function formatPercent(p: number | null | undefined): string {
  if (p === null || p === undefined || Number.isNaN(p)) return '--'
  return `${(p * 100).toFixed(1)}%`
}


</script>

<style scoped src="@/styles/assessment-report.css"></style>
<style scoped src="@/styles/pdf.css"></style>
