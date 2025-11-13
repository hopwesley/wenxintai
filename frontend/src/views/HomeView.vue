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
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                <title>形状结合</title>
                <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                  <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                    <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                  </g>
                </g>
              </svg>
              </div>
              <div class="list-intro">基础得分</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg></div>
              <div class="list-intro">组合推荐</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg></div>
              <div class="list-intro">注意问题</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg></div>
              <div class="list-intro">总结推荐</div>
            </li>
          </ul>
          <RouterLink to="/login" class="btn btn-primary w-full">开始测试</RouterLink>
        </div>
        <p class="plan-tip">需登录</p>
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
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">更全面的维度对比</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg></div>
              <div class="list-intro">深度解释与策略</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg></div>
              <div class="list-intro">历史记录与对比</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg></div>
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
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg>
              </div>
              <div class="list-intro">班级/年级对比</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg></div>
              <div class="list-intro">批量生成报告</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg></div>
              <div class="list-intro">匿名分析与画像</div>
            </li>
            <li class="plan-lists">
              <div class="list-icon">
                <svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                  <title>形状结合</title>
                  <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" opacity="0.300000012">
                    <g id="banner01备份" transform="translate(-324, -742)" fill="#5A60EA">
                      <path d="M337.608242,745.911968 L337.575942,745.984855 C337.119327,747.069316 336.836904,748.567485 336.836904,750.222317 C336.836904,751.769892 337.083903,753.180453 337.489381,754.243768 C337.129462,754.709541 336.710603,755.128402 336.244547,755.489232 C335.182367,755.083015 333.771411,754.835864 332.223358,754.835864 C330.530609,754.835864 329.001785,755.131376 327.912006,755.606561 C327.330883,755.182193 326.817434,754.668743 326.392425,754.086618 C326.868431,752.996904 327.163763,751.468498 327.163763,749.776268 C327.163763,748.228215 326.916611,746.81726 326.51091,745.75383 C326.871565,745.288761 327.289801,744.870524 327.755104,744.51014 C328.8183,744.915571 330.229256,745.162722 331.777309,745.162722 C333.469539,745.162722 334.997944,744.86739 336.087658,744.392462 C336.669701,744.81631 337.183199,745.329796 337.608242,745.911968 Z" id="形状结合" transform="translate(332.0003, 749.9995) rotate(-45) translate(-332.0003, -749.9995)"></path>
                    </g>
                  </g>
                </svg></div>
              <div class="list-intro">导出数据与看板</div>
            </li>
          </ul>
          <div class="btn btn-disabled w-full" aria-disabled="true">敬请期待</div>
        </div>
        <p class="plan-tip">需签约</p>
      </button>
    </section>
    <section class="summary container">
      <div class="summary-inner">
        <h3>基础版 - 详情介绍</h3>
        <p>RIASE与ASC兼顾文理协同招生范围</p>
        <div class="summary-list">
          <div class="summary-item" v-for="item in summaryItems" :key="item.title">
            <h4>{{ item.title }}</h4>
            <p>{{ item.text }}</p>
          </div>
        </div>
        <div class="summary-cta">
          <button class="btn btn-primary" type="button" @click="startTest">开始测试</button>
          <p class="summary-hint">使用邀请码免费体验</p>
        </div>
      </div>
    </section>
    <section>
      <div style="margin: 0 auto;color: #b0b1b3; text-align: center;padding: 16px;border-top: 1px solid #ECECEE; font-size: 14px">域世安（北京）科技有限公司 | 京ICP备2025150532号-1</div>
    </section>
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
import { useTestSession } from '@/store/testSession'

const showLogin = ref(false)
const inviteModalOpen = ref(false)
const router = useRouter()
const { setInviteCode } = useTestSession()

function startTest() {
  inviteModalOpen.value = true
}

function openLogin() {
  showLogin.value = true
  console.log('[HomeView] dialogOpen ->', showLogin.value)
}

function handleInviteSuccess(payload: { code: string; sessionId?: string }) {
  setInviteCode(payload.code)
  router.push('/test/basic/step/1')
}

type PlanKey = 'public' | 'pro' | 'school'
const activePlan = ref<PlanKey>('public')

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
