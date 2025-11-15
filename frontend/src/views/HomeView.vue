<template>
  <div class="home">
    <section class="hero">
      <div class="site-header">
        <div class="header-container">
          <div class="logo-dot" aria-label="问心台">
          </div>

          <!-- 新增：主导航 tab -->
          <nav class="main-nav-tabs">
            <button
                v-for="tab in tabDefs"
                :key="tab.key"
                type="button"
                class="nav-tab"
                :class="{ 'is-active': activeTab === tab.key }"
                @click="handleTabClick(tab)"
            >
              {{ tab.label }}
            </button>
          </nav>
          <button class="btn btn-ghost login-btn" @click="openLogin">登录</button>
        </div>
      </div>

      <div class="hero-inner container">
        <div class="hero-copy">
          <h1>AI生成题目,千人千面</h1>
          <p class="hero-desc">
            基于 RIASEC + ASC，结合能力权重与组合覆盖，自动生成个性化问卷与报告。
          </p>
          <div class="hero-cta">
            <RouterLink to="/login" class="btn btn-primary">立即开始</RouterLink>
          </div>
        </div>
      </div>
    </section>
    <section class="plans container">
      <button
          class="plan-card plan-a"
          :class="{ 'is-active': activePlan === 'public' }"
          @click="activePlan = 'public'"
          type="button"
      >
        <div class="planA-head">基础版</div>
        <div class="plan-card-content">
          <div class="plan-head">
            <div class="price"><span class="currency">¥</span>19.9</div>
          </div>
          <ul class="plan-features">
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">基础得分</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">组合推荐</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">注意问题</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">总结推荐</div>
            </li>
          </ul>
          <RouterLink to="/login" class="btn btn-primary w-full" @click.prevent="startTest('basic')">开始测试
          </RouterLink>
        </div>
        <p class="plan-tip">邀请码免费测试</p>
      </button>
      <button
          class="plan-card plan-pro"
          :class="{ 'is-active': activePlan === 'pro', disabled: true }"
          @click="activePlan = 'pro'"
          type="button"
          aria-disabled="true"
      >
        <div class="planA-head">专业版</div>
        <div class="plan-card-content">
          <div class="plan-head">
            <div class="price price-gray"><span class="currency">¥</span>79.9</div>
          </div>
          <ul class="plan-features">
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">更全面的维度对比</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">深度解释与策略</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">历史记录与对比</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">导出 PDF 报告</div>
            </li>
          </ul>
          <div class="btn btn-disabled w-full" aria-disabled="true">敬请期待</div>
        </div>
        <p class="plan-tip">需邀请</p>
      </button>
      <button
          class="plan-card plan-school"
          :class="{ 'is-active': activePlan === 'school', disabled: true }"
          @click="activePlan = 'school'"
          type="button"
          aria-disabled="true"
      >
        <div class="planA-head">学校版</div>
        <div class="plan-card-content">
          <div class="plan-head">
            <div class="price price-gray"><span class="currency">¥</span>49.9</div>
          </div>
          <ul class="plan-features">
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">班级/年级对比</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">批量生成报告</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">匿名分析与画像</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg"
                     xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path
                          d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z"
                          id="形状结合"
                          transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">导出数据与看板</div>
            </li>
          </ul>
          <div class="btn btn-disabled w-full" aria-disabled="true">敬请期待</div>
        </div>
        <p class="plan-tip">需签约</p>
      </button>
    </section>

    <!-- ① 选科引擎核心价值阐述（新模块） -->
    <section id="section-product-intro" class="home-section value-section">
      <div class="container">
        <header class="section-header">
          <h2>选科引擎核心价值阐述</h2>
          <p>
            亲爱的家长、同学：新高考背景下，选科不再是简单的课程选择，而是一次影响孩子全面发展的战略决策。
            选科失误，可能让孩子与心仪的专业失之交臂；盲目从众，更会埋没其独特天赋与潜能。
            选科引擎应运而生，致力于以科学与智能，终结选科焦虑，为每一个家庭、每一个孩子拓宽理想之路。
          </p>
        </header>

        <div class="value-grid">
          <article class="value-card">
            <div class="value-illus value-illus-eval">
            </div>
            <h3>科学评估，洞见真实自我</h3>
            <p>
              深度融合霍兰德职业兴趣理论与大五人格等经典心理学模型，不仅分析“擅长学什么”，
              更洞察“更适合成为什么样的人”。通过精准的能力图谱与价值观探索，帮助孩子发现与其特质高度契合的“天赋赛道”。
            </p>
          </article>

          <article class="value-card">
            <div class="value-illus value-illus-ai"></div>
            <h3>AI赋能，决策精准前瞻</h3>
            <p>
              核心算法实时对接全国高校选科要求与职业发展大数据，综合多维指标生成多套精选选科组合，
              并附带大学专业群与职业方向预览，让孩子和家长更清楚地看见不同选择背后的长期影响。
            </p>
          </article>

          <article class="value-card">
            <div class="value-illus value-illus-family"></div>
            <h3>化解分歧，凝聚家庭共识</h3>
            <p>
              测评结果同时面向学生与家长展示，用客观数据揭示共识与差异，
              为家庭对话提供理性基准，将一次紧张的“选科大战”变成一起规划未来的契机。
            </p>
          </article>
        </div>
      </div>
    </section>
    <!-- ② 科学与人文的双重考量（新模块） -->
    <section class="home-section theory-section">
      <div class="container">
        <header class="section-header">
          <h2>科学与人文的双重考量</h2>
        </header>
        <section class="home-section theory-section">
          <div class="container">
            <header class="section-header">
              <h2>科学与人文的双重考量</h2>
            </header>
            <!-- 第一行：圆环图 + 教育发展 & 心理学理论 -->
            <div class="theory-row">
              <div class="theory-illus theory-illus-donut"></div>
              <div class="theory-text">
                <h3>教育发展理论支撑</h3>
                <p>
                  我们相信，智能是多元的。孩子的优势可能体现在语言、逻辑、空间、动手、人际、自省等不同维度。
                  在题目设计与报告解读中，我们综合参考教育发展理论与新高考政策要求，
                  既关注长期发展潜力，也兼顾当前阶段的学习负担与适应性。
                </p>

                <h3>心理学理论支撑</h3>
                <p>
                  · 霍兰德职业兴趣理论：评估孩子在现实型、研究型、艺术型、社会型、企业型、常规型六大类型的倾向，
                  把兴趣与未来大学专业及职业环境进行匹配。<br />
                  · 自我决定理论：关注孩子的自主感、胜任感与关系感（是否得到重要他人的支持），
                  这是激发持续学习动力的重要心理基础。
                </p>
              </div>
            </div>
            <!-- 第二行：圆环图 + 教育发展 & 心理学理论 -->
            <div class="theory-row theory-row-right">
              <div class="theory-illus theory-illus-psych"></div>
              <div class="theory-text">
                <h3>心理学理论支撑</h3>
                <p>
                  · 霍兰德职业兴趣理论：评估孩子在现实型、研究型、艺术型、社会型、企业型、常规型六大类型的倾向，
                  把兴趣与未来大学专业及职业环境进行匹配。<br />
                  · 自我决定理论：关注孩子的自主感、胜任感与关系感（是否得到重要他人的支持），
                  这是激发持续学习动力的重要心理基础。
                </p>
              </div>
            </div>


            <!-- 第二行：阶梯图 + 基本算法考量 -->
            <div class="theory-row">
              <div class="theory-illus theory-illus-steps"></div>
              <div class="theory-text">
                <h3>基本算法考量</h3>
                <p>
                  选科引擎采用多指标动态加权模型，而非简单的“分数相加”。
                  综合考虑兴趣匹配度、学科能力、组合覆盖率与罕见度约束等因素，
                  推荐结果不仅给出“选什么”，也解释“为什么这么选”，让家长与学生更容易理解和信任。
                </p>
              </div>
            </div>
          </div>
        </section>
      </div>
    </section>
    <!-- ③ 致家长的一封信 + 数据隐私 + CTA（新模块） -->
    <section id="section-parent-letter" class="home-section letter-section">
      <div class="container">
        <!-- 上方双 Tab -->
        <div class="letter-switch">
          <button
              type="button"
              class="letter-tab"
              :class="{ 'is-active': activeLetterTab === 'parent' }"
              @click="activeLetterTab = 'parent'"
          >
            致家长的一封信
          </button>
          <button
              type="button"
              class="letter-tab"
              :class="{ 'is-active': activeLetterTab === 'student' }"
              @click="activeLetterTab = 'student'"
          >
            致学生的一封信
          </button>
        </div>

        <div class="letter-content">
          <p v-if="activeLetterTab === 'parent'">
            亲爱的家长：选科不仅关系到孩子未来三年的学习安排，更影响接下来十年以上的人生走向。
            我们希望用一套严谨但不冰冷的工具，帮助您在理性判断与情感关怀之间找到平衡，
            与孩子一起看见真正适合 TA 的方向，而不是被一时的分数和外界声音裹挟。
          </p>
          <p v-else>
            亲爱的同学：你正在做出的，是一次关于“我想成为什么样的人”的选择。
            选科引擎不会替你做决定，而是帮你看见更多可能，
            让每一次决定都建立在对自我、对世界更清晰的理解之上。
          </p>
        </div>

        <!-- 数据隐私声明 -->
        <div class="privacy-block">
          <h3>关于数据隐私的严正声明</h3>
          <p>
            我们承诺，所有收集到的个人数据都将经过严格的匿名化与脱敏处理。
            任何可以识别具体个人身份的信息（如姓名、身份证号、学校、班级等）
            均会被分离并加密存储，分析只使用无法追溯到个人的统计数据。
            这些数据仅用于教学研究、模型优化和公益报告等，不会用于任何商业获利行为。
            我们严格遵守相关法律法规，构建完善的数据安全保护体系。
          </p>
        </div>

        <!-- CTA：开始你的测试 -->
        <div class="cta-start-test">
          <div class="cta-inner">
            <div class="cta-text">
              <h3>开始你的测试</h3>
              <p>用大约 20–30 分钟，完成一份专属于你的选科测评报告。</p>
            </div>
            <button class="btn btn-primary" @click="startTest('basic')">
              开始测试
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- 原来的公司信息 section 保持在最后 -->
    <section>
      <div
          style="margin: 0 auto;color: #b0b1b3; text-align: center;padding: 16px;border-top: 1px solid #ECECEE; font-size: 14px">
        域世安（北京）科技有限公司 | 京ICP备2025150532号-1
      </div>
    </section>


    <section>
      <div
          style="margin: 0 auto;color: #b0b1b3; text-align: center;padding: 16px;border-top: 1px solid #ECECEE; font-size: 14px">
        域世安（北京）科技有限公司 | 京ICP备2025150532号-1
      </div>
    </section>
    <InviteCodeModal v-model:open="inviteModalOpen" @success="handleInviteSuccess"/>
  </div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import InviteCodeModal from '@/views/components/InviteCodeModal.vue'
import {useRouter} from 'vue-router'
import {useTestSession} from '@/store/testSession'
import {fetchTestFlow, type TestRouteDef, type NextRouteInfo} from '@/api'
import {useAlert} from '@/logic/useAlert'
import { useAuthStore } from '@/store/auth'

const {showAlert} = useAlert()
const inviteModalOpen = ref(false)
const router = useRouter()
const authStore = useAuthStore()
const {state, setInviteCode, setTestType, setTestRoutes} = useTestSession()
// 顶部导航 tab 配置
const tabDefs = [
  { key: 'start', label: '开始测试', targetId: 'section-start-test' },
  { key: 'intro', label: '产品介绍', targetId: 'section-product-intro' },
  { key: 'letter', label: '致家长的一封信', targetId: 'section-parent-letter' },
] as const

type TabKey = (typeof tabDefs)[number]['key']
const activeTab = ref<TabKey>('start')

// “致家长 / 致学生” tab
type LetterTabKey = 'parent' | 'student'
const activeLetterTab = ref<LetterTabKey>('parent')

function handleTabClick(tab: typeof tabDefs[number]) {
  activeTab.value = tab.key

  const el = document.getElementById(tab.targetId)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

function startTest(typ: string) {
  setTestType(typ)
  inviteModalOpen.value = true
}

function openLogin() {
  authStore.openWeChatLogin()
  console.log('[HomeView] dialogOpen ->', authStore.wechatLoginOpen)
}

async function handleInviteSuccess(payload: { code: string; sessionId?: string }) {
  setInviteCode(payload.code)
  const typ = state.testType || 'basic'
  const req = {
    test_type: typ,
    invite_code: payload.code,
    wechat_openid: state.wechatOpenId as string | undefined,
  }

  let routes: TestRouteDef[] = []
  let nextRoute: NextRouteInfo | null | undefined

  try {
    const resp = await fetchTestFlow(req)
    routes = resp.routes || []
    nextRoute = resp.nextRoute ?? null

    setTestRoutes(routes)

  } catch (e) {
    console.error('[handleInviteSuccess] fetchTestFlow failed', e)
    showAlert('获取测试流程失败，请稍后再试:' + e)
    return
  }

  const targetRouter =
      nextRoute?.router || (routes.length > 0 ? routes[0].router : null)

  if (!targetRouter) {
    console.warn('[handleInviteSuccess] no target router found')
    showAlert('测试流程配置异常，请稍后再试或联系管理员')
    return
  }
  await router.push(`/assessment/${typ}/${targetRouter}`)
}

type PlanKey = 'public' | 'pro' | 'school'
const activePlan = ref<PlanKey>('public')
</script>
<style scoped src="@/styles/home.css"></style>