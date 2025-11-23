<template>
  <div class="subject-bar-wrapper">
    <!-- 上方：一句话说明 + 颜色图例 -->
    <div class="subject-bar-meta">
      <p class="subject-bar-caption">
        六门学科的相对强弱（Z 分视图）
      </p>
      <ul class="subject-bar-legend">
        <li class="subject-bar-legend__item">
          <span class="subject-bar-legend__dot subject-bar-legend__dot--interest"></span>
          兴趣明显高于能力
        </li>
        <li class="subject-bar-legend__item">
          <span class="subject-bar-legend__dot subject-bar-legend__dot--ability"></span>
          能力明显高于兴趣
        </li>
        <li class="subject-bar-legend__item">
          <span class="subject-bar-legend__dot subject-bar-legend__dot--balanced"></span>
          兴趣与能力大致平衡
        </li>
      </ul>
    </div>

    <VChart
        v-if="option"
        class="subject-bar-chart"
        :option="option"
        autoresize
    />
    <div v-else class="chart-placeholder">
      基础能力柱状图暂无数据
    </div>

    <!-- 下方：AI 对各学科整体格局的文字总结 -->
    <p class="subject-bar-summary">
      注：本图中的兴趣Z分、能力Z分和Z差值，表示的是你在六门学科内部的相对高低，并非雷达图中 0–100 的原始兴趣/能力百分比分数。
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ReportSubjectScore } from '@/controller/AssessmentReport'

const props = defineProps<{
  subjects: ReportSubjectScore[] | null
  summaryText?: string
}>()

const subjectLabelMap: Record<string, string> = {
  PHY: '物理',
  CHE: '化学',
  BIO: '生物',
  GEO: '地理',
  HIS: '历史',
  POL: '政治',
}

// 根据 zgap 给柱子分色：兴趣主导 / 能力主导 / 相对平衡
function zgapColor(zgap: number): string {
  if (zgap <= -1) return '#f97316' // 兴趣明显 > 能力（橙色）
  if (zgap >= 1) return '#22c55e'  // 能力明显 > 兴趣（绿色）
  return '#60a5fa'                // 基本平衡（蓝色）
}

const option = computed(() => {
  const list = props.subjects || []
  if (!list.length) return null

  const names = list.map(s => subjectLabelMap[s.subject] ?? s.subject)

  const seriesData = list.map(s => ({
    value: s.ability_z,
    itemStyle: {
      color: zgapColor(s.zgap),
    },
  }))

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter(params: any) {
        const p = Array.isArray(params) ? params[0] : params
        const s = list[p.dataIndex]
        if (!s) return ''
        return [
          `${subjectLabelMap[s.subject] ?? s.subject}`,
          `能力 z：${s.ability_z.toFixed(2)}`,
          `兴趣 z：${s.interest_z.toFixed(2)}`,
          `2gap：${s.zgap.toFixed(2)}`,
          `能力占比：${(s.ability_share * 100).toFixed(1)}%`,
          `fit：${s.fit.toFixed(2)}`
        ].join('<br/>')
      },
    },
    grid: {
      left: 40,
      right: 20,
      bottom: 30,
      top: 28, // 给上方 caption/legend 腾一点空间
    },
    xAxis: {
      type: 'category',
      data: names,
      axisTick: { alignWithLabel: true },
    },
    yAxis: {
      type: 'value',
      min: -3,
      max: 3,
      splitNumber: 6,
      axisLine: { show: true },
      splitLine: { show: true },
    },
    series: [
      {
        type: 'bar',
        data: seriesData,
        barWidth: '45%',
      },
    ],
  }
})
</script>

<style scoped>
.subject-bar-wrapper {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* 顶部说明 + 图例 */
.subject-bar-meta {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 2px;
}

.subject-bar-caption {
  font-size: 12px;
  color: #6b7280;
  line-height: 1.5;
}

.subject-bar-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  list-style: none;
  margin: 0;
  padding: 0;
}

.subject-bar-legend__item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #4b5563;
}

.subject-bar-legend__dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  display: inline-block;
}

/* 和 zgapColor 中的配色保持一致 */
.subject-bar-legend__dot--interest {
  background: #f97316;
}

.subject-bar-legend__dot--ability {
  background: #22c55e;
}

.subject-bar-legend__dot--balanced {
  background: #60a5fa;
}

/* 图本身 */
.subject-bar-chart {
  width: 100%;
  height: 260px;
}

/* 下方 AI 总结文本 */
.subject-bar-summary {
  font-size: 13px;
  line-height: 1.6;
  color: #4b5563;
}
</style>
