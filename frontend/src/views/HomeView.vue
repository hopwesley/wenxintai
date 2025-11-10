<template>
  <div class="home">
    <!-- 顶部横幅 / Hero -->
    <section class="hero">
      <!-- 顶部导航条：左 logo 右 登录 -->
      <div class="site-header container">
        <div class="logo-dot" aria-label="问心台">
          <!-- 用图片当背景，或把这个 div 换成 <img src="/img/logo.png"> -->
        </div>
        <button class="btn btn-ghost login-btn" @click="openLogin">登录</button>
      </div>

      <div class="hero-inner container">
        <div class="hero-copy">
          <h1>AI生成题目</h1>
          <h2>千人千面</h2>
          <p class="hero-desc">
            基于 RIASEC + ASC，结合能力权重与组合覆盖，自动生成个性化问卷与报告。
          </p>
          <div class="hero-cta">
            <RouterLink to="/login" class="btn btn-primary">立即开始</RouterLink>
          </div>
        </div>

        <div class="hero-visual">
          <!-- 先用 logo 占位 -->
          <img src="/img/logo.png" alt="hero" />
        </div>
      </div>
    </section>

    <!-- 方案卡片 / Plans（可切换） -->
    <section class="plans container">
      <button
          class="plan-card plan-a"
          :class="{ 'is-active': activePlan === 'public' }"
          @click="activePlan = 'public'"
          type="button"
      >
        <div class="plan-head">
          <h3>基础版</h3>
          <div class="price"><span class="currency">¥</span>19.9</div>
        </div>
        <ul class="plan-features">
          <li>基础得分</li>
          <li>组合推荐</li>
          <li>注意问题</li>
          <li>总结推荐</li>
        </ul>
        <RouterLink to="/login" class="btn btn-primary w-full">开始测试</RouterLink>
        <p class="plan-tip">需登录</p>
      </button>

      <button
          class="plan-card plan-pro"
          :class="{ 'is-active': activePlan === 'pro', disabled: true }"
          @click="activePlan = 'pro'"
          type="button"
          aria-disabled="true"
      >
        <div class="plan-head">
          <h3>专业版</h3>
          <div class="price"><span class="currency">¥</span>79.9</div>
        </div>
        <ul class="plan-features">
          <li>更全面的维度对比</li>
          <li>深度解释与策略</li>
          <li>历史记录与对比</li>
          <li>导出 PDF 报告</li>
        </ul>
        <div class="btn btn-disabled w-full" aria-disabled="true">敬请期待</div>
        <p class="plan-tip">需邀请</p>
      </button>

      <button
          class="plan-card plan-school"
          :class="{ 'is-active': activePlan === 'school', disabled: true }"
          @click="activePlan = 'school'"
          type="button"
          aria-disabled="true"
      >
        <div class="plan-head">
          <h3>学校版</h3>
          <div class="price"><span class="currency">¥</span>19.9</div>
        </div>
        <ul class="plan-features">
          <li>班级/年级对比</li>
          <li>批量生成报告</li>
          <li>匿名分析与画像</li>
          <li>导出数据与看板</li>
        </ul>
        <div class="btn btn-disabled w-full" aria-disabled="true">敬请期待</div>
        <p class="plan-tip">需签约</p>
      </button>
    </section>

    <!-- 说明模块 / Feature strip（随选择切换） -->
    <section class="summary container">
      <div class="summary-inner">
        <div class="summary-item" v-for="item in summaryItems" :key="item.title">
          <h4>{{ item.title }}</h4>
          <p>{{ item.text }}</p>
        </div>
      </div>

      <div class="summary-cta">
        <button class="btn btn-primary" type="button" @click="startTest">开始测试</button>
        <p class="summary-hint">{{ t('invite.freeHint') }}</p>
      </div>
    </section>
    <!-- 登录弹窗：双向绑定 -->
    <WeChatLoginDialog v-model:open="showLogin" />
    <InviteCodeModal v-model:open="inviteModalOpen" @success="handleInviteSuccess" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import '@/styles/home.css'
import WeChatLoginDialog from '@/components/WeChatLoginDialog.vue'
import InviteCodeModal from '@/components/InviteCodeModal.vue'
import { useRouter } from 'vue-router'
import { useI18n } from '@/i18n'

const showLogin = ref(false)

const inviteModalOpen = ref(false)
const router = useRouter()
const { t } = useI18n()

function startTest() {
  inviteModalOpen.value = true
}

function openLogin() {
  showLogin.value = true
  console.log('[HomeView] dialogOpen ->', showLogin.value)
}

function handleInviteSuccess() {
  router.push('/test/basic/step/1')
}

type PlanKey = 'public' | 'pro' | 'school'

/** 默认选中“基础版” */
const activePlan = ref<PlanKey>('public')

/** 三个版本对应的说明文案（你可以随时改成自己的最终文案） */
const PLAN_DETAILS: Record<PlanKey, { items: { title: string; text: string }[] }> = {
  public: {
    items: [
      { title: '基础得分', text: '汇总兴趣与学科自概念，提供清晰的基础画像，便于家长与学生快速理解。' },
      { title: '组合推荐', text: '结合人岗匹配、能力加权与覆盖矩阵，从多维度筛选更稳妥的高考选科组合。' },
      { title: '注意问题', text: '自动识别潜在风险与矛盾点，提示学习分配、备考节奏与资源优先级。' },
      { title: '总结推荐', text: '提供可执行的阶段建议，支持与班主任、家长沟通，统一决策语言。' },
    ],
  },
  pro: {
    items: [
      { title: '维度对比', text: '更多性格/能力/学科子维度对比，可视化定位优势短板，提供个体曲线。' },
      { title: '深度解释', text: '对冲分与冲突点专项说明，给出学习节奏与资源调度策略建议。' },
      { title: '历史对比', text: '支持多次测试的时序对比，观察兴趣/能力趋势与干预效果。' },
      { title: '导出报告', text: '一键导出 PDF 报告，适配打印与留存。' },
    ],
  },
  school: {
    items: [
      { title: '班级/年级对比', text: '快速对比班级与年级画像，定位班群体差异与共性。' },
      { title: '批量报告', text: '批量生成学生个体报告，并支持匿名化汇总。' },
      { title: '匿名分析', text: '生成年级/班级的匿名分析图像，辅助教学与选科宣讲。' },
      { title: '数据看板', text: '导出数据用于校内看板或家校沟通。' },
    ],
  },
}

const summaryItems = computed(() => PLAN_DETAILS[activePlan.value].items)
</script>
