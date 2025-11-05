<template>
  <div class="home">
    <!-- 顶部横幅 / Hero -->
    <section class="hero">
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

    <!-- 方案卡片 / Plans -->
    <section class="plans container">
      <div class="plan-card plan-a">
        <div class="plan-head">
          <h3>公益版</h3>
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
      </div>

      <div class="plan-card plan-pro disabled">
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
        <button class="btn btn-disabled w-full" disabled>敬请期待</button>
        <p class="plan-tip">需邀请</p>
      </div>

      <div class="plan-card plan-school disabled">
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
        <button class="btn btn-disabled w-full" disabled>敬请期待</button>
        <p class="plan-tip">需签约</p>
      </div>
    </section>

    <!-- 说明模块 / Feature strip -->
    <section class="summary container">
      <div class="summary-inner">
        <div class="summary-item">
          <h4>基础得分</h4>
          <p>
            汇总兴趣与学科自概念得分，提供清晰的基础画像，便于家长与学生快速理解。
          </p>
        </div>
        <div class="summary-item">
          <h4>组合推荐</h4>
          <p>
            结合人岗匹配、能力加权与覆盖矩阵，从多维度筛选更稳妥的高考选科组合。
          </p>
        </div>
        <div class="summary-item">
          <h4>注意问题</h4>
          <p>
            自动识别潜在风险与矛盾点，提示学习分配、备考节奏与资源优先级。
          </p>
        </div>
        <div class="summary-item">
          <h4>总结推荐</h4>
          <p>
            提供可执行的阶段建议，支持与班主任、家长沟通，统一决策语言。
          </p>
        </div>
      </div>

      <div class="summary-cta">
        <RouterLink to="/login" class="btn btn-primary">开始测试</RouterLink>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
/* 纯展示页，无额外逻辑 */
</script>
