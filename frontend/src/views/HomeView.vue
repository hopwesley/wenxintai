<template>
  <div class="home">
    <section class="hero">
      <div class="site-header">
        <div class="header-container">
          <div style="display: flex; justify-content: center;">
            <div class="logo-dot" aria-label="问心台">
            </div>
            <h2 style="font-weight: 600;font-size: 20px;color: var(--brand)">智择未来</h2>
            <p style="font-weight: 400;font-size: 14px;color: var(--text-third)">新高考AI选科智能决策系统</p>

          </div>
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
          <div class="header-right">
            <button
                v-if="!isLoggedIn"
                class="btn btn-ghost login-btn"
                type="button"
                @click="openLogin"
            >
              微信登录
            </button>
            <div
                v-else
                class="home-user-wrapper"
                ref="userMenuWrapperRef"
            >
              <!-- 已登录：用 signInStatus 里的头像 & 昵称 -->
              <button
                  type="button"
                  class="home-user"
                  @click="handleUserClick"
              >
                <img
                    v-if="signInStatus.avatar_url"
                    :src="signInStatus.avatar_url"
                    alt="微信头像"
                    class="home-user__avatar"
                />
                <span
                    v-if="signInStatus.nick_name"
                    class="home-user__name"
                >
                {{ signInStatus.nick_name }}
              </span>
              </button>
              <!-- 下拉菜单 -->
              <div
                  v-if="isUserMenuOpen"
                  class="home-user-menu"
              >
                <button
                    type="button"
                    class="home-user-menu__item"
                    @click="handleGoMyTests"
                >
                  我的测试
                </button>
                <button
                    type="button"
                    class="home-user-menu__item home-user-menu__item--danger"
                    @click="handleLogout"
                >
                  退出登录
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="hero-inner container">
        <div class="hero-copy">
          <h1>AI生成探索题目，千人千面</h1>
          <p class="hero-desc">
            基于 RIASEC + ASC，结合能力权重、兴趣偏好与组合覆盖，自动生成个性化问卷
            <br>
            通过精准能力画像与价值观探索，帮助孩子发现“天赋赛道”
          </p>
        </div>
      </div>
    </section>
    <section class="plans container" id="section-start-test">
      <div
          class="plan-card plan-a"
          :class="{ 'is-active': activePlan === TestTypeBasic }"
          @click="activePlan = TestTypeBasic"
      >
        <div class="planA-head"> {{ currentProductsMap[TestTypeBasic].name }}</div>
        <div class="plan-card-content">
          <div class="plan-head">
            <div class="price"><span class="currency">¥</span>{{ currentProductsMap[TestTypeBasic].price }}</div>
          </div>
          <ul class="plan-features">
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
              <div class="list-intro">综合建议</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg"
                >
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
          </ul>
          <RouterLink to="/login" class="btn btn-primary w-full" @click.prevent="startTest(TestTypeBasic)">开始测试
          </RouterLink>
        </div>
        <p class="plan-tip">邀请码免费测试</p>
      </div>
      <div
          class="plan-card plan-pro"
          :class="{ 'is-active': activePlan === TestTypePro, disabled: true }"
          @click="activePlan = TestTypePro"
          aria-disabled="true"
      >
        <div class="planA-head">{{ currentProductsMap[TestTypePro].name }}</div>
        <div class="plan-card-content">
          <div class="plan-head">
            <div class="price"><span class="currency">¥</span>{{ currentProductsMap[TestTypePro].price }}</div>
          </div>
          <ul class="plan-features">
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
              <div class="list-intro">更全面的纬度对比</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
              <div class="list-intro">学习力系统性评价</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
              <div class="list-intro">导出PDF报告</div>
            </li>
          </ul>
          <div class="btn w-full" aria-disabled="true" @click.prevent="startTest(TestTypePro)">开始测试</div>
        </div>
        <p class="plan-tip">邀请码免费测试</p>
      </div>
      <div
          class="plan-card plan-pro"
          :class="{ 'is-active': activePlan === TestTypeAdv, disabled: true }"
          @click="activePlan = TestTypeAdv"
          aria-disabled="true"
      >
        <div class="planA-head">{{ currentProductsMap[TestTypeAdv].name }}</div>
        <div class="plan-card-content">
          <div class="plan-head">
            <div class="price price-gray"><span class="currency">¥</span>{{ currentProductsMap[TestTypeAdv].price }}
            </div>
          </div>
          <ul class="plan-features">
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
              <div class="list-intro">加入价值观纬度</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
              <div class="list-intro">糅合职业前景分析</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
              <div class="list-intro">发展趋势前瞻报告</div>
            </li>
          </ul>
          <div class="btn btn-disabled w-full" aria-disabled="true">敬请期待</div>
        </div>
        <p class="plan-tip">暂未开放</p>
      </div>
      <div
          class="plan-card plan-school"
          :class="{ 'is-active': activePlan === TestTypeSchool, disabled: true }"
          @click="activePlan = TestTypeSchool"
          aria-disabled="true"
      >
        <div class="planA-head">{{ currentProductsMap[TestTypeSchool].name }}</div>
        <div class="plan-card-content">
          <div class="plan-head">
            <div class="price price-gray"><span class="currency">¥</span>{{ currentProductsMap[TestTypeSchool].price }}
            </div>
          </div>
          <ul class="plan-features">
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
                <svg width="10px" height="10px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
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
      </div>
    </section>
    <!-- ① 选科引擎核心价值阐述（新模块） -->
    <section id="section-product-intro" class="home-section value-section">
      <div class="container">
        <header class="section-header">
          <h2>智择未来核心价值阐述</h2>
          <p>
            新高考背景下，选科不只是简单的课程选择，而是一次影响孩子全面发展的战略决策。选科失误，可能让
            孩子与心仪的专业失之交臂；盲目从众，更会埋没其独特天赋与潜能。智择未来应运而生，致力于以科学
            与智能，终结选科焦虑，为每一个家庭、每一个孩子拓宽理想之路。
          </p>
        </header>
        <div class="value-grid">
          <article class="value-card">
            <div class="value-illus value-illus-eval">
            </div>
            <h3>科学评估，洞见真实自我</h3>
            <p>
              我们深度融合霍兰德职业兴趣理论与大五人格模型等经典心理学理论，不止于分析“能学什么”，更
              深度探测“喜欢什么”、“适合什么”。通过精准的能力画像与价值观探索，我们帮助孩子发现与其特
              质高度契合的“天赋赛道”。
            </p>
          </article>

          <article class="value-card">
            <div class="value-illus value-illus-ai"></div>
            <h3>AI赋能，决策精准前瞻</h3>
            <p>
              我们凭借核心算法实时同步全国高校选科要求与职业发展大数据，通过动态决策模型，交叉分析个人
              特质与外部世界，生成文、理、工等2-3种最优选科组合及适配度评分，并提供对应的大学专业群与
              职业方向预览，让孩子的未来清晰可见。
            </p>
          </article>

          <article class="value-card">
            <div class="value-illus value-illus-family"></div>
            <h3>化解分歧，凝聚家庭共识</h3>
            <p>
              选科决策中，父母与孩子因视角不同，常面临理想与现实的碰撞，使选科成为家庭矛盾的导火索。
              为此，我们用算法和客观数据清晰呈现孩子特质，让主观猜测让位于科学洞察，致力于将可能的决策矛盾，
              转化为共同规划未来的宝贵契机，助力每个家庭在理解与信任中，携手走向明朗的未来。
            </p>
          </article>
        </div>
      </div>
    </section>
    <!-- ② 科学与人文的双重考量（新模块） -->
    <section class="home-section theory-section">
      <div class="container">
        <header class="section-header">
          <h2>智择未来算法逻辑：科学与人文的双重考量</h2>
        </header>
        <!-- 第一行：圆环图 + 教育发展 & 心理学理论 -->
        <div class="theory-row">
          <div class="theory-illus theory-illus-donut"></div>
          <div class="theory-text">
            <h3>教育发展理论支撑</h3>
            <div class="theory-point">
              <h4>加德纳的“多元智能理论”</h4>
              <p>
                我们相信，智能是多元的。孩子的优势可能体现在语言、逻辑、数学、空间、人际、自省等多种智能的不同组合上。选科，正是为了聚焦长处的同时，避其短板。</p>
            </div>
            <div class="theory-point">
              <h4>“T型人才”与“π型人才”</h4>
              <p>
                未来社会需要既具备某一领域的深度（“T”的竖），又拥有广博知识面与跨界合作能力（“T”的横）的人才，甚至是拥有两项专业深度的“π型人才”。我们的评估，正是为了识别并培育这种深度与广度的最佳结合点。</p>
            </div>
          </div>
        </div>
        <!-- 第二行：圆环图 + 教育发展 & 心理学理论 -->
        <div class="theory-row theory-row-right">
          <div class="theory-illus theory-illus-psych"></div>
          <div class="theory-text">
            <h3>心理学理论支撑</h3>
            <div class="theory-point">
              <h4>霍兰德职业兴趣理论</h4>
              <p>
                通过评估孩子在现实型、研究型、艺术型、社会型、企业型、常规型六大类型的倾向，将其兴趣与未来大学专业和职业环境进行匹配，确保内在驱动与外部选择的和谐统一。</p>
            </div>
            <div class="theory-point">
              <h4>自我决定理论</h4>
              <p>
                我们关注孩子的自主感（是否自己认同）、胜任感（是否觉得自己能学好）和归属感（是否觉得该选择被重要他人支持），因为这是激发持续学习动力的心理基石。</p>
            </div>
          </div>
        </div>
      </div>
    </section>
    <!-- 算法核心考量维度（放在“智择未来算法逻辑”板块后面） -->
    <section class="home-section theory-section theory-dimensions">
      <div class="container">
        <div class="theory-dimensions__inner">
          <h2 class="theory-dimensions__title">
            所以，我们的智能算法引擎是一个多维度的动态模型，核心考量包括：
          </h2>

          <div class="theory-dimensions__grid">
            <!-- 1 学科能力现状与潜能 -->
            <div class="dimension-card">
              <div class="dimension-icon dimension-icon-ability"></div>
              <h3>学科能力现状与潜能</h3>
              <p>
                不仅看当前成绩，更关注成绩走势、在群体中的相对位置，
                以及学习过程中展现出的思维方式与发展潜力。
              </p>
            </div>

            <!-- 2 稳定的兴趣倾向 -->
            <div class="dimension-card">
              <div class="dimension-icon dimension-icon-interest"></div>
              <h3>稳定的兴趣倾向</h3>
              <p>
                通过多次、多情境的评估，剥离临时情绪与偶然因素，
                找到孩子真正愿意长期投入的热爱所在。
              </p>
            </div>

            <!-- 3 认知风格与人格特质 -->
            <div class="dimension-card">
              <div class="dimension-icon dimension-icon-cognitive"></div>
              <h3>认知风格与人格特质</h3>
              <p>
                孩子是偏向抽象思维还是具象感知？是细致严谨还是开拓创新？
                这些都会影响不同学科的学习适配度与发展路径。
              </p>
            </div>

            <!-- 4 外部环境匹配度 -->
            <div class="dimension-card">
              <div class="dimension-icon dimension-icon-context"></div>
              <h3>外部环境匹配度</h3>
              <p>
                结合本省高考政策、高校选考要求大数据，以及区域与社会发展的长期趋势，
                评估学科组合与外部环境的整体匹配度。
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
    <!-- ③ 致亲爱的学生和家长（仅信件内容，带背景图） -->
    <section id="section-parent-letter" class="home-section letter-section">
      <div class="letter-bg">
        <div class="container">
          <div class="letter-layout">
            <div class="letter-main">
              <header class="section-header section-header--letter">
                <h2>致亲爱的学生和家长</h2>
              </header>

              <div class="letter-content">
                <p>
                  选科，是一次基于孩子独特天赋与未来社会需求的早期战略规划。我们深知其分量，因此，我们调动了最前沿的教育理念、心理学研究与大数据技术，旨在为您提供一份不负所托的决策参考。
                  它明确解释推荐理由，坦诚提示发展风险，让孩子在扬长避短中最大化高考竞争力，并为未来的专业与职业发展保留最大主动权。
                  选择智择未来，就是选择以专业和远见，将当下的选择，变为对未来最成功的投资。我们的相遇源于您对孩子教育的深切关注与巨大付出，更源于您对孩子人生道路的尊重与信任。您最关心的不仅是学科前景，更是孩子未来的幸福与发展。
                  我们坚信，每一个孩子都是一颗独一无二的星辰，我们的使命，是帮助他找到最属于自己的轨道，从而熠熠生辉。
                  诚邀您立即启程，为孩子的未来锁定最优赛道！
                </p>
                <p>
                  亲爱的同学：你正在做出的，是一次关于“我想成为什么样的人”的选择。
                  选科引擎不会替你做决定，而是帮你看见更多可能，
                  让每一次决定都建立在对自我、对世界更清晰的理解之上。
                </p>
              </div>

            </div>
          </div>
        </div>
      </div>
    </section>
    <!-- ④ 关于数据隐私的严正声明（独立板块） -->
    <section id="section-privacy" class="home-section privacy-section value-section">
      <div class="container">
        <div class="privacy-block">
          <h2 class="report-section__title">关于数据隐私的严正声明</h2>
          <p>
            我们承诺，所有收集到的个人数据都将经过严格的匿名化与脱敏处理。任何可以识别具体个人身份的信息
            （如姓名、身份证号、学校、班级等） 均会被分离并加密存储，且数据仅用于教学研究、模型优化和公益
            报告等。 我们严格遵守相关法律法规，构建完善的数据安全保护体系。
          </p>
        </div>
      </div>
    </section>
    <!-- ⑤ 开始你的测试（独立 CTA 板块） -->
    <section id="section-start-test" class=" cta-section value-section" style="padding-top: 0">
      <div class="container">
        <div class="cta-start-test">
          <div class="cta-inner">
            <div class="cta-text">
              <h2>开始你的测试</h2>
              <p>用10~20分钟，获取专属于你的《智择未来 · AI选科全景分析报告》。 </p>
            </div>
            <button class="btn btn-primary" @click="scrollToStartTest">
              开始测试
            </button>
          </div>
        </div>
      </div>
    </section>
    <section id="icp-area">
      <div
          style="margin: 0 auto;color: #b0b1b3; text-align: center;padding: 16px;border-top: 1px solid #ECECEE; font-size: 14px">
        域世安（北京）科技有限公司 | 京ICP备2025150532号-1
      </div>
    </section>
    <NewUserInfoDialog v-model:open="newUserDialogOpen"/>
    <TestDisclaimerDialog
        :visible="showDisclaimer"
        @confirm="handleDisclaimerConfirm"
        @cancel="handleDisclaimerCancel"
    />
  </div>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useHomeView} from '@/controller/HomeView'

import {useAuthStore} from '@/controller/wx_auth'
import {useRouter} from "vue-router";
import {TestTypeAdv, TestTypeBasic, TestTypePro, TestTypeSchool, currentProductsMap} from "@/controller/common";
import NewUserInfoDialog from "@/views/components/NewUserInfoDialog.vue";
import TestDisclaimerDialog from "@/views/components/TestDisclaimerDialog.vue";

const {
  activePlan,
  activeTab,
  tabDefs,
  openLogin,
  startTest,
  handleTabClick,
  handleUserClick,
  isUserMenuOpen,
  userMenuWrapperRef,
  handleGoMyTests,
  handleLogout,
  showDisclaimer,
  handleDisclaimerCancel,
  handleDisclaimerConfirm,
} = useHomeView()

const authStore = useAuthStore()
const router = useRouter()

const signInStatus = computed(() => authStore.signInStatus)
const newUserDialogOpen = ref(false)
const isLoggedIn = computed(
    () => signInStatus.value.status === 'ok'
)

watch(
    () => authStore.loginStatus,
    (status) => {
      if (status !== 'success') return
      const isNew = signInStatus.value?.is_new === true

      console.log("------>>> newUserInfoDismissed value:", authStore.newUserInfoDismissed)

      if (isNew && !authStore.newUserInfoDismissed) {
        console.log('[HomeView] 微信登录成功，新用户，打开补充信息弹窗')
        newUserDialogOpen.value = true
      } else {
        console.log('[HomeView] 微信登录成功，老用户或已关闭提醒，跳回首页')
        router.push('/')
      }
    },
)

const scrollToStartTest = () => {
  const el = document.getElementById('section-start-test')
  if (!el) return
  el.scrollIntoView({behavior: 'smooth', block: 'start'})
}

</script>

<style scoped src="@/styles/home.css"></style>