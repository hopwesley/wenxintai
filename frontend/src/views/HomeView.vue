<template>
  <div class="home">
    <section class="hero">
      <div class="site-header">
        <div class="header-container">
          <div class="logo-dot" aria-label="问心台">
          </div>
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
    <section>
      <div
          style="margin: 0 auto;color: #b0b1b3; text-align: center;padding: 16px;border-top: 1px solid #ECECEE; font-size: 14px">
        域世安（北京）科技有限公司 | 京ICP备2025150532号-1
      </div>
    </section>
    <WeChatLoginDialog v-model:open="showLogin"/>
    <InviteCodeModal v-model:open="inviteModalOpen" @success="handleInviteSuccess"/>
  </div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import '@/styles/home.css'
import WeChatLoginDialog from '@/components/WeChatLoginDialog.vue'
import InviteCodeModal from '@/components/InviteCodeModal.vue'
import {useRouter} from 'vue-router'
import {useTestSession} from '@/store/testSession'
import {fetchTestFlow, type TestRouteDef, type NextRouteInfo} from '@/api'
import { useAlert } from '@/logic/useAlert'

const { showAlert } = useAlert()
const showLogin = ref(false)
const inviteModalOpen = ref(false)
const router = useRouter()
const {state, setInviteCode, setTestType, setTestRoutes} = useTestSession()

function startTest(typ: string) {
  setTestType(typ)
  inviteModalOpen.value = true
}

function openLogin() {
  showLogin.value = true
  console.log('[HomeView] dialogOpen ->', showLogin.value)
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
  await router.push(`/test/${typ}/${targetRouter}`)
}

type PlanKey = 'public' | 'pro' | 'school'
const activePlan = ref<PlanKey>('public')

</script>
