<template>
  <nav class="step-indicator" aria-label="progress">
    <ol>
      <li
          v-for="(step, index) in steps"
          :key="index"
          :class="stepClass(index)"
          @click="handleStepClick(index)"
      >
        <div class="step-indicator__circle" aria-hidden="true">
          <span>{{ index + 1 }}</span>
        </div>
        <div class="step-indicator__title">{{ step }}</div>
        <div v-if="index < steps.length - 1" class="step-indicator__connector" aria-hidden="true"></div>
      </li>
    </ol>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTestSession } from '@/controller/testSession'
import { pushStageRoute, StageBasic, StageReport } from '@/controller/common'

const route = useRoute()
const router = useRouter()
const { state } = useTestSession()

// 展示用的标题列表：优先用 testFlowSteps（带 stage+title），兜底 testRoutes（老的 string[]）
const steps = computed(() => {
  const flow = state.testFlowSteps ?? []
  if (flow.length) {
    return flow.map(step => step.title)
  }
  return state.testRoutes ?? []
})

/**
 * 根据当前路由，推导“所在阶段”的 stage key
 * - /assessment/:typ/basic-info       -> StageBasic
 * - /assessment/:typ/report           -> StageReport
 * - /assessment/:businessType/:testStage -> route.params.testStage
 */
function getStageFromRoute(): string {
  const name = route.name

  if (name === 'test-basic-info') {
    return StageBasic
  }
  if (name === 'test-report') {
    return StageReport
  }

  // 题目阶段：/assessment/:businessType/:testStage
  return String(route.params.testStage ?? '')
}

// 当前所在步骤 index
const current = computed(() => {
  const stage = getStageFromRoute()
  const flow = state.testFlowSteps ?? []

  // 1) 优先用 nextRouteItem 中记录的 index（这是我们在每次跳转时设进去的）
  if (stage) {
    const idxFromMap = state.nextRouteItem?.[stage]
    if (typeof idxFromMap === 'number') {
      return idxFromMap
    }
  }

  // 2) 没有 nextRouteItem 时，在 flow 里按 stage 查找
  if (flow.length && stage) {
    const idx = flow.findIndex(it => it.stage === stage)
    if (idx >= 0) {
      return idx
    }
  }

  // 3) 兜底：0（通常是“基础信息”）
  return 0
})

function stepClass(index: number) {
  if (index < current.value) return 'is-complete'
  if (index === current.value) return 'is-current'
  return 'is-upcoming'
}

/**
 * 点击步骤：
 * - 只允许点击“已完成步骤 + 当前步骤”（index <= current）
 * - 按 index 从 testFlowSteps 里找到目标 stage
 * - 根据当前路由取出 businessType，然后 pushStageRoute 跳转
 */
function handleStepClick(index: number) {
  const currentIndex = current.value
  if (index > currentIndex) return // 禁止点未来步骤

  const flow = state.testFlowSteps ?? []
  const target = flow[index]
  if (!target) return

  // businessType 可能来自两种路由形式：
  // - /assessment/:businessType/:testStage    (questions)
  // - /assessment/:typ/basic-info / report    (basic + report)
  const biz =
      (route.params.businessType as string | undefined) ||
      (route.params.typ as string | undefined) ||
      state.businessType ||
      ''

  if (!biz || !target.stage) return

  pushStageRoute(router, biz, target.stage)
}
</script>


<style scoped>
  .step-indicator {
    display: flex;
    justify-content: center;
  }

  .step-indicator ol {
    display: flex;
    list-style: none;
    padding: 0;
    margin: 0;
    gap: 16px;
    width: 100%;
  }

  .step-indicator li {
    position: relative;
    flex: 1;
    display: flex;
    align-items: center;
    gap: 12px;
    color: rgba(30, 41, 59, 0.6);
    font-weight: 500;
  }

  .step-indicator__circle {
    width: 36px;
    height: 36px;
    border-radius: 18px;
    border: 2px solid rgba(99, 102, 241, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #fff;
    transition: all 0.2s ease;
  }

  .step-indicator__title {
    font-size: 14px;
    line-height: 1.4;
  }

  .step-indicator__connector {
    position: absolute;
    right: -8px;
    top: 50%;
    transform: translateY(-50%);
    width: calc(100% - 44px);
    height: 2px;
    background: repeating-linear-gradient(
        to right,
        rgba(99, 102, 241, 0.3) 0,
        rgba(99, 102, 241, 0.3) 8px,
        transparent 8px,
        transparent 16px
    );
    pointer-events: none;
  }

  .step-indicator li.is-complete .step-indicator__circle {
    background: linear-gradient(135deg, #6366f1, #8b5cf6);
    border-color: transparent;
    color: white;
  }

  .step-indicator li.is-current .step-indicator__circle {
    border-color: #6366f1;
    box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.12);
  }

  .step-indicator li.is-upcoming .step-indicator__circle {
    border-color: rgba(148, 163, 184, 0.6);
  }

  @media (max-width: 768px) {
    .step-indicator ol {
      flex-direction: column;
      gap: 12px;
    }

    .step-indicator li {
      align-items: flex-start;
    }

    .step-indicator__connector {
      display: none;
    }
  }
</style>
