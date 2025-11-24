<template>
  <TestLayout :key="route.fullPath">
    <template #header>
      <StepIndicator/>
    </template>
    <main class="report-page">
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

              <!-- 占位单元，让 4x2 网格视觉更均衡 -->
              <div class="report-field report-field--placeholder"></div>
              <div class="report-field report-field--placeholder"></div>
            </div>
          </div>
        </section>

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
                <span class="ai-text-block__value"> {{ rawReportData?.common_score.common.global_cosine }}  </span>
                <span>(-1.0表示兴趣与能力完全不匹配，1.0表示兴趣与能力完全匹配)</span>
              </p>
              <p class="ai-text-block__line">
                <span class="ai-text-block__label">数据质量评分：</span>
                <span class="ai-text-block__value"> {{ rawReportData?.common_score.common.quality_score }} </span>
                <span>(高于 0.4表示答题可信，低于 0.4表示答题内容不可信)</span>
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

        <section class="report-section report-section--concepts">
          <h2 class="report-section__title">核心指标说明</h2>
          <p class="report-section__intro">
            下表对本报告中涉及的关键指标进行简要说明，建议在阅读图表和文字解读前先浏览一遍，
            方便理解各学科在“兴趣”“能力”“匹配度”等维度上的含义。
          </p>

          <div class="field-definitions">
            <table class="field-definitions__table">
              <thead>
              <tr>
                <th class="field-definitions__cell field-definitions__cell--head field-definitions__cell--key">
                  字段
                </th>
                <th class="field-definitions__cell field-definitions__cell--head">
                  含义
                </th>
              </tr>
              </thead>
              <tbody>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">
                  兴趣z
                </td>
                <td class="field-definitions__cell">
                  兴趣强度的标准化值，表示该学科的内在动机水平（数值越高，代表兴趣驱动越强）。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">
                  能力z
                </td>
                <td class="field-definitions__cell">
                  能力强度的标准化值，表示该学科的自我效能感水平（数值越高，代表学习信心越强）。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">
                  fit
                </td>
                <td class="field-definitions__cell">
                  单科兴趣–能力匹配度，数值越高，说明兴趣与能力越协调，学习往往更顺畅且持续性更好。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">
                  zgap
                </td>
                <td class="field-definitions__cell">
                  兴趣与能力的差距（兴趣z − 能力z）。正值表示能力领先兴趣，负值表示兴趣主导能力。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">
                  能力占比
                </td>
                <td class="field-definitions__cell">
                  各学科能力在整体中的占比，反映学习信心与精力投入的重心（所有学科之和约为 1）。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">
                  整体匹配度
                </td>
                <td class="field-definitions__cell">
                  兴趣–能力总体方向一致性。数值越高，说明自我认同越清晰、整体发展方向越稳定。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">
                  数据质量评分
                </td>
                <td class="field-definitions__cell">
                  本次测评数据的可信度指标，数值越高，代表答题过程越稳定、报告结论的可靠性越高。
                </td>
              </tr>
              </tbody>
            </table>
          </div>

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
                  zGap
                </th>
                <th class="report-table__cell report-table__cell--head">
                  能力占比
                </th>
                <th class="report-table__cell report-table__cell--head">
                  fit
                </th>
              </tr>
              </thead>

              <tbody v-if="rawReportData && rawReportData.common_score && rawReportData.common_score.common">
              <tr
                  v-for="sub in rawReportData.common_score.common.subjects"
                  :key="sub.subject"
              >
                <td class="report-table__cell report-table__cell--subject">
                  {{ subjectLabelMap[sub.subject] ?? sub.subject }}
                </td>
                <td class="report-table__cell">
                  {{ formatZ(sub.interest_z) }}
                </td>
                <td class="report-table__cell">
                  {{ formatZ(sub.ability_z) }}
                </td>
                <td class="report-table__cell">
                  {{ formatZ(sub.zgap) }}
                </td>
                <td class="report-table__cell">
                  {{ formatPercent(sub.ability_share) }}
                </td>
                <td class="report-table__cell">
                  {{ formatZ(sub.fit) }}
                </td>
              </tr>
              </tbody>

              <!-- 没数据时的兜底行（可选） -->
              <tbody v-else>
              <tr>
                <td class="report-table__cell" colspan="6">
                  暂无基础参数数据
                </td>
              </tr>
              </tbody>
            </table>
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
      <!-- 下半部分：选科组合推荐卡片 -->

      <section class="report-card report-card--recommendation">
        <header class="report-card__header">
          <div class="report-card__title-row">
            <h2 class="report-card__title">选科组合推荐</h2>
          </div>
        </header>
        <!-- 3+3 模式参数说明：字体偏小、不占太多空间 -->
        <div class="report-card__divider"></div>
        <!-- ===================== 3+3 模式：沿用现有结构 ===================== -->
        <div v-if="isMode33">
          <div class="field-definitions field-definitions--compact">
            <table class="field-definitions__table">
              <thead>
              <tr>
                <th class="field-definitions__cell field-definitions__cell--head">参数名称</th>
                <th class="field-definitions__cell field-definitions__cell--head">对选科的影响说明</th>
              </tr>
              </thead>
              <tbody>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">平均匹配度</td>
                <td class="field-definitions__cell">
                  反映孩子在这三门科目上兴趣与能力的整体协调程度。数值越高，说明既有兴趣又有相对优势，
                  更利于长期投入和稳定发挥，是判断“这三科是否适合长期学下去”的核心指标之一。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">最低能力等级</td>
                <td class="field-definitions__cell">
                  三门科目中当前能力相对最弱的一门。数值越低，短板越明显，选为组合时需要在备考中为这门科目
                  预留更多时间和支持，否则整体成绩容易被这一门拖累。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">方向协同性</td>
                <td class="field-definitions__cell">
                  衡量三门科目的学习方式、思维特点是否接近。协同性越高，科目之间切换成本越低，
                  孩子在时间与精力分配上会更顺畅，不容易出现“每天都在完全不同类型学科之间来回换挡”的疲惫感。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">综合得分</td>
                <td class="field-definitions__cell">
                  对该组合整体适配度的综合评价，在匹配度、能力基础和风险等多个维度之间做平衡。
                  分数越高，整体越适合作为优先考虑的选科方案。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">稀有度</td>
                <td class="field-definitions__cell">
                  表示在当前地区报考中，该组合被选择的多少程度。数值越高越少见，可能带来竞争对手较少、
                  但课程资源、志愿填报参考信息相对不足等双重影响，一般需要家长和学生额外关注对应院校的选科要求。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">风险惩罚</td>
                <td class="field-definitions__cell">
                  综合考虑学科短板、兴趣冲突和组合稀有带来的不稳定因素。值越高，说明在时间精力分配、
                  成绩波动或志愿填报上需要更谨慎，更适合作为备选方案，而不是唯一依赖的组合。
                </td>
              </tr>
              </tbody>
            </table>
          </div>

          <!-- 5. 推荐概览（将来可以放两张小图） -->
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
          <!-- 全局说明：为什么不直接替你选物理 / 历史 -->
          <div class="field-definitions field-definitions--compact">
            <table class="field-definitions__table">
              <thead>
              <tr>
                <th class="field-definitions__cell field-definitions__cell--head">参数名称</th>
                <th class="field-definitions__cell field-definitions__cell--head">对选科的影响说明</th>
              </tr>
              </thead>
              <tbody>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">综合得分</td>
                <td class="field-definitions__cell">
                  综合反映主干科目与两门辅科的整体适配情况，结合兴趣、能力和风险等多维因素得出。
                  分数越高，说明该三科组合更稳、更契合学生当前的兴趣和学习实力，适合作为优先考虑方案。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">专业覆盖率</td>
                <td class="field-definitions__cell">
                  表示该三科组合在当前省份规则下能覆盖的专业范围。覆盖率越高，说明未来志愿选择空间更大，
                  可报考专业更丰富，升学灵活度更高。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">组合风险</td>
                <td class="field-definitions__cell">
                  衡量该组合在文理跨度、学习方式冲突或短板科目影响下的潜在不稳定性。
                  数值偏高时，说明需要在时间分配和备考节奏上更谨慎，建议作为备选组合。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">辅科平均匹配度</td>
                <td class="field-definitions__cell">
                  反映两门辅科在兴趣与能力上的整体协调性。数值越高，说明这两门课更贴合学生的学习风格，
                  同时具备一定的兴趣驱动和学习信心。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">辅科平均能力值</td>
                <td class="field-definitions__cell">
                  表示学生在两门辅科上的平均能力水平。数值越高，说明在非主干科目上学习基础更扎实，
                  具备较好的扩展潜力和学习稳健度。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">辅科最低匹配度</td>
                <td class="field-definitions__cell">
                  两门辅科中相对较弱的一门在兴趣与能力上的匹配度，用来识别潜在短板。
                  如果该值明显偏低，需要在后续学习中对该科目给予更多关注与支持。
                </td>
              </tr>
              <tr>
                <td class="field-definitions__cell field-definitions__cell--key">辅科一致性</td>
                <td class="field-definitions__cell">
                  衡量两门辅科在学习方式与思维风格上的一致程度。数值越高，说明两门课在复习节奏、
                  思维模式上更协调，切换成本更低，整体学习更顺畅。
                </td>
              </tr>
              </tbody>
            </table>
          </div>

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
      </section>

      <section class="report-section report-section--summary">
        <h3 class="report-section__title">报告摘要</h3>

        <!-- 优先使用 AI 的 final_report（3+3 / 3+1+2 都走这里） -->
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
          <article  class="summary-card"  >
            <p>无总结报告可显示</p>
          </article>
        </div>
      </section>


    </main>

    <div class="report-page__actions">
      <!-- TODO：绑定具体路由或打印逻辑 -->
      <button class="btn btn-secondary report-page__action">
        返回测试首页
      </button>
      <button class="btn btn-primary report-page__action">
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
import {subjectLabelMap} from "@/controller/common";

const {
  state,
  route,
  overview,
  aiLoading,
  truncatedLatestMessage,
  summaryCards,
  subjectRadar,
  rawReportData,
  isMode33,
  isMode312,
  mode33View,
  mode312OverviewStrips,
  finalReport,
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
