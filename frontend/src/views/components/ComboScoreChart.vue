<template>
  <div class="combo-score-chart-wrapper">
    <VChart
        v-if="option"
        class="combo-score-chart"
        :option="option"
        autoresize
    />
    <div v-else class="chart-placeholder">
      推荐组合整体分布概览暂无数据
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Mode33ChartCombo } from '@/controller/AssessmentReport'
import {subjectLabelMap} from "@/controller/common";

const props = defineProps<{
  combos: Mode33ChartCombo[] | null
}>()

const option = computed(() => {
  const list = props.combos ?? []
  if (!list.length) return null

  const xLabels = list.map(c =>
      c.subjects.map(s => subjectLabelMap[s] ?? s).join(' + '),
  )

  // ✅ 使用原始分数，不再乘以 100
  const scores = list.map(c => c.score)

  // ✅ 动态计算最大值，让 y 轴范围更贴合数据
  const maxScore = Math.max(...scores)
  const yMax = maxScore === 0 ? 1 : maxScore * 1.1  // 稍微留一点顶部空间

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter(params: any) {
        const p = Array.isArray(params) ? params[0] : params
        const idx = p.dataIndex
        const label = xLabels[idx]
        const score = scores[idx]
        // ✅ 显示原始值，这里我用 3 位小数，你可以自己调
        return `${label}<br/>综合得分：${score.toFixed(3)}`
      },
    },
    grid: {
      top: 24,
      left: 16,
      right: 16,
      bottom: 40,
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: xLabels,
      axisLabel: {
        interval: 0,
        fontSize: 11,
      },
      axisTick: { show: false },
      axisLine: { lineStyle: { color: '#9ca3af' } },
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: yMax,          // ✅ 用原始值范围
      splitNumber: 5,
      axisLabel: {
        // ✅ 按需要格式化显示，比如两位小数
        formatter: (val: number) => val.toFixed(2),
        fontSize: 11,
      },
      axisLine: { show: true },
      splitLine: { show: true },
    },
    series: [
      {
        type: 'bar',
        data: scores,
        barWidth: '40%',
        itemStyle: {
          borderRadius: [6, 6, 2, 2],
        },
      },
    ],
  }
})
</script>


<style scoped>
.combo-score-chart-wrapper {
  width: 100%;
}

.combo-score-chart {
  width: 100%;
  height: 200px;
}
</style>
