<template>
  <div class="subject-radar-wrapper">
    <VChart
        v-if="option"
        class="subject-radar-chart"
        :option="option"
        autoresize
    />
    <div v-else class="chart-placeholder chart-placeholder--radar">
      暂无数据
    </div>

    <!-- 打印 & 精读用的数据表，保证数值不会丢 -->
    <table v-if="rows.length" class="subject-radar-table">
      <thead>
      <tr>
        <th>科目</th>
        <th>兴趣（%）</th>
        <th>能力（%）</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="row in rows" :key="row.label">
        <td>{{ row.label }}</td>
        <td>{{ formatPct(row.interest) }}</td>
        <td>{{ formatPct(row.ability) }}</td>
      </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ReportRadarBlock } from '@/controller/AssessmentReport'

// props：直接吃 ReportRadarBlock
const props = defineProps<{
  radar: ReportRadarBlock | null
}>()

// 学科编码 -> 中文标签
const subjectLabelMap: Record<string, string> = {
  PHY: '物理',
  CHE: '化学',
  BIO: '生物',
  GEO: '地理',
  HIS: '历史',
  POL: '政治',
}

// 打印用：把雷达里的数组拆成表格行
const rows = computed(() => {
  const r = props.radar
  if (!r || !r.subjects || !r.subjects.length) return []

  return r.subjects.map((sub, idx) => ({
    label: subjectLabelMap[sub] ?? sub,
    interest: r.interest_pct[idx] ?? 0,
    ability: r.ability_pct[idx] ?? 0,
  }))
})

function formatPct(v?: number): string {
  if (v == null || Number.isNaN(v)) return '--'
  return `${v.toFixed(1)}%`
}

// ECharts 配置
const option = computed(() => {
  const r = props.radar
  if (!r || !r.subjects || !r.subjects.length) return null

  return {
    // 备用色盘（可以不写，只用下面的显式颜色）
    // color: ['#1d9bf0', '#a855f7'],

    tooltip: {
      trigger: 'item',
    },
    legend: {
      data: ['兴趣', '能力'],
      top: 0,
      left: 'center',
    },
    radar: {
      // 每个维度
      indicator: r.subjects.map((sub: string) => ({
        name: subjectLabelMap[sub] ?? sub,
        max: 100,
        min: 0,
      })),
      splitNumber: 4,
      radius: '70%',            // 放大雷达主体
      center: ['50%', '55%'],   // 稍微往下挪，给 legend 腾空间
      axisName: {
        fontSize: 11,
      },
    },
    series: [
      {
        type: 'radar',
        data: [
          {
            value: r.interest_pct,
            name: '兴趣',
            lineStyle: {
              width: 2,
              color: '#1d9bf0', // 兴趣：主题蓝
            },
            areaStyle: {
              opacity: 0.25,
              color: 'rgba(29,155,240,0.25)',
            },
            itemStyle: {
              color: '#1d9bf0',
            },
          },
          {
            value: r.ability_pct,
            name: '能力',
            lineStyle: {
              width: 2,
              color: '#a855f7', // 能力：对比紫
            },
            areaStyle: {
              opacity: 0.18,
              color: 'rgba(168,85,247,0.18)',
            },
            itemStyle: {
              color: '#a855f7',
            },
            // 如果你希望图上也常显数值，可以打开这段 label
            // label: {
            //   show: true,
            //   formatter: (params: any) => {
            //     const { value, dataIndex } = params
            //     if (!Array.isArray(value)) return ''
            //     const v = value[dataIndex]
            //     if (v == null || Number.isNaN(v)) return ''
            //     return `${v.toFixed(0)}`
            //   },
            //   fontSize: 10,
            // },
          },
        ],
      },
    ],
  }
})
</script>

<style scoped>
.subject-radar-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* 把图放大一点，打印时更清晰 */
.subject-radar-chart {
  width: 100%;
  height: 320px;
}

/* 表格：紧凑一点，适合 A4 打印 */
.subject-radar-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  line-height: 1.4;
  margin-top: 4px;
}

.subject-radar-table th,
.subject-radar-table td {
  padding: 4px 6px;
  border-top: 1px solid #e5e7eb;
  text-align: right;
}

.subject-radar-table th:first-child,
.subject-radar-table td:first-child {
  text-align: left;
}

.subject-radar-table thead th {
  font-weight: 500;
  color: #6b7280;
  border-bottom: 1px solid #e5e7eb;
}
</style>
