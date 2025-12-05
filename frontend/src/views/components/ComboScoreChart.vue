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
import {computed} from 'vue'
import type {ComboChartItem} from '@/controller/report_manager'
import {subjectLabelMap} from '@/controller/common'

const props = defineProps<{
  combos: ComboChartItem[] | null
}>()

const option = computed(() => {
  const list = props.combos ?? []
  if (!list.length) return null

  const metricCount = list[0].metrics.length
  if (metricCount === 0) return null

  // x 轴：把 comboKey 拆成学科代码再映射成中文
  const xLabels = list.map(item => {
    const codes = item.comboKey.split('_')
    const names = codes.map(c => subjectLabelMap[c] ?? c)
    return names.join(' + ')
  })

  // 收集所有值（允许为 0）
  const allValues: number[] = []
  list.forEach(item => {
    item.metrics.slice(0, metricCount).forEach(m => allValues.push(m.value))
  })

  // 非 0 的最大值，用来判断范围；如果全是 0，则 maxVal = 0
  const nonZero = allValues.filter(v => v !== 0)
  const maxVal = nonZero.length ? Math.max(...nonZero) : 0
  const yMax = maxVal === 0 ? 1 : maxVal * 1.1

  // legend 文案：直接用 metric 的 label
  const legendData = list[0].metrics.slice(0, metricCount).map(m => m.label)

  // 小工具：格式化数值（用于 y 轴 / label / tooltip）
  const formatValue = (v: number) => {
    if (!isFinite(v)) return ''
    if (v === 0) return '0'
    const absMax = Math.max(Math.abs(maxVal), Math.max(...allValues.map(Math.abs)))
    if (absMax < 1) return v.toFixed(3)
    if (absMax < 10) return v.toFixed(2)
    return v.toFixed(1)
  }

  const series = list[0].metrics.slice(0, metricCount).map((metric, idx) => {
    return {
      name: metric.label,
      type: 'bar',
      data: list.map(item => {
        return item.metrics[idx]?.value ?? 0
      }),
      barWidth: metricCount === 1 ? '45%' : '32%',
      itemStyle: {
        borderRadius: [6, 6, 2, 2],
      },
      // ✅ 在柱子顶端显示数值，打印 PDF 时也能看到
      label: {
        show: true,
        position: 'top',
        fontSize: 10,
        color: '#374151',
        formatter(params: any) {
          const v = typeof params.value === 'number'
              ? params.value
              : Number(params.value)
          return formatValue(v)
        },
      },
    }
  })

  return {

    animation: false,
    animationDuration: 0,
    animationDurationUpdate: 0,

    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter(params: any) {
        const arr = Array.isArray(params) ? params : [params]
        const idx = arr[0].dataIndex
        const label = xLabels[idx]

        const lines = arr.map(p => {
          const v = typeof p.data === 'number' ? p.data : Number(p.data)
          return `${p.seriesName}：${formatValue(v)}`
        })

        return `${label}<br/>${lines.join('<br/>')}`
      },
    },
    legend: {
      data: legendData,
      bottom: 4,
    },
    grid: {
      top: 24,
      left: 16,
      right: 16,
      bottom: 46,  // 给 legend 留出空间
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
      max: yMax,               // ✅ 全 0 时固定到 1，避免图直接消失
      splitNumber: 5,
      axisLabel: {
        formatter: (val: number) => formatValue(val),
        fontSize: 11,
      },
      axisLine: { show: true },
      splitLine: { show: true },
    },
    series,
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
