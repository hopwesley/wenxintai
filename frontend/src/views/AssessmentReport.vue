<template>
  <TestLayout :key="route.fullPath">
    <!-- 顶部步骤条 -->
    <template #header>
      <StepIndicator/>
    </template>

    <!-- 报告主体 -->
    <main class="report-page">
      <!-- ===== 报告基础信息卡片 ===== -->
      <section class="report-card report-card--overview">
        <!-- 标题行 -->
        <header class="report-card__header">
          <div class="report-card__title-row">
            <h1 class="report-card__title">问心台选课报告</h1>
            <span class="report-card__mode-pill">{{ overview.mode || '—' }}</span>
          </div>
        </header>

        <div class="report-card__divider"></div>

        <!-- 个人资料 -->
        <section class="report-section report-section--profile">
          <div class="report-profile">
            <!-- 左侧标题，单独一行 -->
            <div class="report-field report-field--title">
              <span class="report-field__label">个人资料</span>
            </div>

            <!-- 右侧 4×2 网格区域 -->
            <div class="report-profile-grid">
              <!-- 第一行 -->
              <div class="report-field">
                <span class="report-field__label">模式</span>
                <span class="report-field__value">{{ overview.mode || '—' }}</span>
              </div>
              <div class="report-field">
                <span class="report-field__label">学生所在地</span>
                <span class="report-field__value">{{ overview.studentLocation || '—' }}</span>
              </div>
              <div class="report-field">
                <span class="report-field__label">报告日期</span>
                <span class="report-field__value">{{ overview.generateDate || '—' }}</span>
              </div>
              <div class="report-field">
                <span class="report-field__label">失效日期</span>
                <span class="report-field__value">{{ overview.expireDate || '—' }}</span>
              </div>

              <!-- 第二行 -->
              <div class="report-field">
                <span class="report-field__label">学生号</span>
                <span class="report-field__value">{{ overview.studentNo || '—' }}</span>
              </div>
              <div class="report-field">
                <span class="report-field__label">学校名称</span>
                <span class="report-field__value">{{ overview.schoolName || '—' }}</span>
              </div>
              <div class="report-field">
                <span class="report-field__label">问心台账号</span>
                <span class="report-field__value">{{ overview.account || '—' }}</span>
              </div>
              <!-- 右下角留空，对齐设计稿 -->
              <div class="report-field report-field--placeholder"></div>
            </div>
          </div>
        </section>

        <!-- 基础分析 -->
        <section class="report-section report-section--basic-analysis">
          <h2 class="report-section__title">基础分析</h2>

          <div class="basic-analysis-layout">
            <div class="basic-analysis-layout__chart basic-analysis-layout__chart--bar">
              <!-- TODO: 替换为真实柱状图组件 -->
              <div class="chart-placeholder">基础能力柱状图</div>
            </div>

            <div class="basic-analysis-layout__chart basic-analysis-layout__chart--radar">
              <!-- TODO: 替换为真实雷达图组件 -->
              <div class="chart-placeholder chart-placeholder--radar">
                科目能力雷达图
              </div>
            </div>
          </div>

          <div class="analysis-interpretation-row">
            <article class="analysis-interpretation">
              <div class="analysis-interpretation__header">
                <span class="analysis-interpretation__title">解读</span>
                <span class="analysis-interpretation__tag">COS</span>
              </div>
              <p class="analysis-interpretation__text">
                这里显示 AI 对基础能力与兴趣匹配程度的文字解读。
              </p>
            </article>

            <article class="analysis-interpretation">
              <div class="analysis-interpretation__header">
                <span class="analysis-interpretation__title">解读</span>
                <span class="analysis-interpretation__tag">COS</span>
              </div>
              <p class="analysis-interpretation__text">
                这里显示 AI 对学科组合整体结构的补充说明。
              </p>
            </article>
          </div>
        </section>

        <!-- 基础概念 -->
        <section class="report-section report-section--concepts">
          <h2 class="report-section__title">基础概念</h2>

          <div class="ai-text-block">
            <p class="ai-text-block__line">
              2gap 表示兴趣与能力之间的差异。
            </p>
            <p class="ai-text-block__line">
              <span class="ai-text-block__label">AI：</span>
              <span class="ai-text-block__content">
                2gap 数值越大，说明兴趣与能力差异越明显，需要在选科时更加谨慎。
              </span>
            </p>
            <p class="ai-text-block__line">
              <span class="ai-text-block__label">AI：</span>
              <span class="ai-text-block__content">
                当 gap 较小时，说明兴趣与能力比较一致，更适合作为优先考虑方向。
              </span>
            </p>
          </div>
        </section>

        <!-- 基础参数 -->
        <section class="report-section report-section--parameters">
          <h2 class="report-section__title">基础参数</h2>

          <table class="report-table">
            <thead>
            <tr>
              <th class="report-table__cell report-table__cell--head">学科</th>
              <th class="report-table__cell report-table__cell--head">兴趣</th>
              <th class="report-table__cell report-table__cell--head">能力</th>
              <th class="report-table__cell report-table__cell--head">gap</th>
              <th class="report-table__cell report-table__cell--head">综合</th>
            </tr>
            </thead>
            <tbody>
            <tr>
              <th class="report-table__cell report-table__cell--subject">物理</th>
              <td class="report-table__cell">XXX</td>
              <td class="report-table__cell">XXX</td>
              <td class="report-table__cell">XXX</td>
              <td class="report-table__cell">XXX</td>
            </tr>
            <tr>
              <th class="report-table__cell report-table__cell--subject">化学</th>
              <td class="report-table__cell">XXX</td>
              <td class="report-table__cell">XXX</td>
              <td class="report-table__cell">XXX</td>
              <td class="report-table__cell">XXX</td>
            </tr>
            <tr>
              <th class="report-table__cell report-table__cell--subject">地理</th>
              <td class="report-table__cell">XXX</td>
              <td class="report-table__cell">XXX</td>
              <td class="report-table__cell">XXX</td>
              <td class="report-table__cell">XXX</td>
            </tr>
            <!-- 其它学科按需补充 -->
            </tbody>
          </table>
        </section>

        <!-- 报告解读（AI 生成） -->
        <section class="report-section report-section--overview-ai">
          <h2 class="report-section__title">报告解读（AI 生成）</h2>

          <div class="ai-text-block">
            <p class="ai-text-block__line">
              本报告基于你的兴趣、能力与新高考政策要求，给出选科建议。
            </p>
            <p class="ai-text-block__line">
              <span class="ai-text-block__label">AI：</span>
              <span class="ai-text-block__content">
                物理与化学在能力和兴趣上都处于较优水平，建议优先考虑。
              </span>
            </p>
            <p class="ai-text-block__line">
              <span class="ai-text-block__label">AI：</span>
              <span class="ai-text-block__content">
                地理兴趣较高但能力略弱，可以通过后续学习进行弥补。
              </span>
            </p>
          </div>
        </section>
      </section>

      <!-- ===== 推荐概述卡片 ===== -->
      <section class="report-card report-card--recommendation">
        <header class="report-card__header">
          <div class="report-card__title-row">
            <h2 class="report-card__title">推荐概述</h2>
            <span class="report-card__mode-pill">3+3</span>
          </div>
        </header>

        <!-- 推荐基础分析（上面两张小图） -->
        <section class="report-section report-section--recommend-analysis">
          <h3 class="report-section__title">基础分析</h3>

          <div class="recommend-analysis-layout">
            <div class="recommend-analysis__chart">
              <div class="chart-placeholder">各档位推荐强度</div>
            </div>
            <div class="recommend-analysis__chart">
              <div class="chart-placeholder">组合整体能力</div>
            </div>
          </div>
        </section>

        <!-- 推荐组合列表：第一档 + 第二档 + 第三档，都在这里 -->
        <section class="report-section report-section--combos">
          <div
              v-for="(combo, index) in recommendedCombos"
              :key="index"
              class="combo-block"
          >
            <!-- 彩色条：第一档/第二档/第三档 -->
            <div
                class="combo-rank-strip"
                :class="`combo-rank-strip--${combo.theme}`"
            >
      <span class="combo-rank-strip__label">
        {{ combo.rankLabel }}：{{ combo.name }}
      </span>
              <span class="combo-rank-strip__score">
        得分：{{ combo.score }}
      </span>
            </div>

            <!-- 下方白色内容卡片 -->
            <article class="combo-panel">
              <!-- 推荐因数 -->
              <header class="combo-panel__header">
                <h4 class="combo-panel__title">推荐因数</h4>
                <dl class="combo-metrics">
                  <div
                      v-for="metric in combo.metrics"
                      :key="metric.label"
                      class="combo-metrics__item"
                  >
                    <dt class="combo-metrics__label">{{ metric.label }}</dt>
                    <dd class="combo-metrics__value">{{ metric.value }}</dd>
                  </div>
                </dl>
              </header>

              <!-- 因数解读 -->
              <section class="combo-panel__section">
                <h5 class="combo-panel__subtitle">因数解读（AI 生成）</h5>
                <div class="combo-panel__ai-block">
                  <p
                      v-for="(line, i) in combo.factorExplain"
                      :key="i"
                      class="combo-panel__text-line"
                  >
                    {{ line }}
                  </p>
                </div>
              </section>

              <!-- 推荐语 -->
              <section class="combo-panel__section">
                <h5 class="combo-panel__subtitle">推荐语（AI 生成）</h5>
                <div class="combo-panel__ai-block">
                  <p
                      v-for="(line, i) in combo.recommendExplain"
                      :key="i"
                      class="combo-panel__text-line"
                  >
                    {{ line }}
                  </p>
                </div>
              </section>
            </article>
          </div>
        </section>

        <!-- 总结（AI 生成） -->
        <section class="report-section report-section--summary">
          <h3 class="report-section__title">总结（AI 生成）</h3>
          <div class="summary-grid">
            <article
                v-for="(card, index) in summaryCards"
                :key="index"
                class="summary-card"
            >
              <h4 class="summary-card__title">{{ card.title }}</h4>
              <p class="summary-card__content">
                {{ card.content }}
              </p>
            </article>
          </div>
        </section>
      </section>

      <!-- 底部按钮 -->
      <footer class="report-page__actions">
        <button type="button" class="btn btn-secondary report-page__action">
          完成
        </button>
        <button type="button" class="btn btn-primary report-page__action">
          下载 PDF
        </button>
      </footer>
    </main>

    <!-- AI 生成中遮罩 -->
    <AiGeneratingOverlay
        v-if="aiLoading"
        title="AI 正在为你生成专属报告…"
        subtitle="正在分析你的测试各项参数，为您全面展示智能分析结果"
        :log-lines="truncatedLatestMessage"
    />
  </TestLayout>
</template>

<script setup lang="ts">
import StepIndicator from '@/views/components/StepIndicator.vue'
import TestLayout from '@/views/components/TestLayout.vue'
import AiGeneratingOverlay from '@/views/components/AiGeneratingOverlay.vue'
import {useReportPage} from '@/controller/AssessmentReport'

const {
  route,
  overview,
  aiLoading,
  truncatedLatestMessage,
  recommendedCombos,
  summaryCards,
} = useReportPage()
</script>

<style scoped src="@/styles/assessment-report.css"></style>
