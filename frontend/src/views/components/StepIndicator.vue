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
import {computed} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useTestSession} from "@/controller/testSession";
import {pushStageRoute} from "@/controller/common";

const route = useRoute()
const router = useRouter()
const {state} = useTestSession()

// 展示用的标题列表（"基础信息" / "兴趣测试" / ...）
const steps = computed(() => state.testRoutes ?? [])

// 当前所在步骤的下标（还是沿用 nextRouteItem 的逻辑）
const current = computed(() => {
  const stageKey = String(route.params.testStage ?? '')
  return state.nextRouteItem?.[stageKey] ?? 0
})

function stepClass(index: number) {
  if (index < current.value) return 'is-complete'
  if (index === current.value) return 'is-current'
  return 'is-upcoming'
}

/**
 * 点击步骤：
 * - 只允许点击“已完成步骤 + 当前步骤”
 * - 根据 index 在 testFlowSteps 里找到对应的 stage
 * - 使用 pushStageRoute 跳转
 */
function handleStepClick(index: number) {
  const currentIndex = current.value

  // 不允许点未来的步骤
  if (index > currentIndex) {
    return
  }

  const flow = state.testFlowSteps ?? []
  const target = flow[index]
  if (!target) return

  const biz =
      (route.params.businessType as string | undefined) ||
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
