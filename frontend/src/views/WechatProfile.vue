<template>
  <div class="my-tests-page home">
    <!-- 顶部：个人档案卡片 + 返回首页 -->
    <header class="my-tests-header container">
      <div class="my-tests-profile-card">
        <div class="my-tests-profile-avatar">
          <img
              v-if="profile?.avatar_url"
              :src="profile.avatar_url"
              alt="avatar"
          />
          <span v-else>{{ getAvatarInitial() }}</span>
        </div>

        <div class="my-tests-profile-info">
          <div class="my-tests-profile-title-row">
            <h1>{{ renderProfileTitle() }}</h1>
            <span class="my-tests-profile-tag">我的测试</span>
          </div>

          <p class="my-tests-profile-sub">
            {{ renderProfileSub() || '欢迎回来，继续你的探索之旅。' }}
          </p>

          <div class="my-tests-profile-meta">
            <span>测评次数：{{ list.length }}</span>
            <span>已生成报告：{{ completedCount }}</span>
          </div>
        </div>
      </div>

      <button type="button" class="btn btn-ghost" @click="handleBackHome">
        返回首页
      </button>
    </header>

    <main class="my-tests-main container">
      <!-- 进行中的测试 -->
      <section v-if="ongoingList.length" class="my-tests-section">
        <h2 class="my-tests-section-title">进行中的测试</h2>
        <div class="my-tests-list">
          <article
              v-for="item in ongoingList"
              :key="item.public_id"
              class="my-tests-card my-tests-card--ongoing"
          >
            <div class="my-tests-card-main">
              <h3 class="my-tests-card-title">
                {{ renderTitle(item) }}
              </h3>
              <p class="my-tests-card-sub">
                创建时间：{{ formatDateTime(item.created_at) }}
              </p>
              <span class="my-tests-status my-tests-status--ongoing">
                {{ renderStatusText(item) }}
              </span>
            </div>

            <button
                type="button"
                class="btn my-tests-card-btn"
                @click="handleContinueTest(item)"
            >
              继续测试
            </button>
          </article>
        </div>
      </section>

      <!-- 已完成的测试 / 报告 -->
      <section v-if="completedList.length" class="my-tests-section">
        <h2 class="my-tests-section-title">已完成的测试与报告</h2>
        <div class="my-tests-list">
          <article
              v-for="item in completedList"
              :key="item.public_id"
              class="my-tests-card"
          >
            <div class="my-tests-card-main">
              <h3 class="my-tests-card-title">
                {{ renderTitle(item) }}
              </h3>
              <p class="my-tests-card-sub">
                完成时间：
                {{ item.completed_at ? formatDateTime(item.completed_at) : '—' }}
              </p>
              <span
                  class="my-tests-status"
                  :class="{
                  'my-tests-status--done': item.status === 'COMPLETED_WITH_REPORT',
                  'my-tests-status--pending': item.status === 'COMPLETED_NO_REPORT'
                }"
              >
                {{ renderStatusText(item) }}
              </span>
            </div>

            <button
                v-if="item.status === 'COMPLETED_WITH_REPORT'"
                type="button"
                class="btn my-tests-card-btn"
                @click="handleOpenReport(item)"
            >
              查看报告
            </button>

            <button
                v-else
                type="button"
                class="btn my-tests-card-btn my-tests-card-btn--disabled"
                @click="handleClickCompletedNoReport(item)"
            >
              报告生成中
            </button>
          </article>
        </div>
      </section>

      <!-- 空状态 -->
      <section
          v-if="!loading && !ongoingList.length && !completedList.length"
          class="my-tests-empty"
      >
        <h2>还没有测试记录</h2>
        <p>可以先回到首页，选择一套适合你的测评开始体验。</p>
        <button type="button" class="btn" @click="handleBackHome">
          去开始测试
        </button>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import {useWechatProfile} from '@/controller/WechatProfile'

const {
  loading,
  profile,
  list,
  ongoingList,
  completedList,
  completedCount,
  renderTitle,
  renderStatusText,
  renderProfileTitle,
  renderProfileSub,
  getAvatarInitial,
  formatDateTime,
  handleContinueTest,
  handleOpenReport,
  handleClickCompletedNoReport,
  handleBackHome,
} = useWechatProfile()
</script>

<style scoped src="@/styles/WechatProfile.css"></style>
