<template>
  <div class="my-tests-page home">
    <div class="my-tests-header-actions">
      <button type="button" class="btn btn-ghost btn-back-home"  @click="handleBackHome">
        <svg
            class="btn-back-home__icon"
            viewBox="0 0 1024 1024"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
            focusable="false"
        >
          <path d="M29.184 464.128m48.384 0l868.608 0q48.384 0 48.384 48.384l0-0.256q0 48.384-48.384 48.384l-868.608 0q-48.384 0-48.384-48.384l0 0.256q0-48.384 48.384-48.384Z"></path>
          <path d="M1.967498 504.159242m37.242805-30.886647l387.009893-320.959644q37.242805-30.886646 68.129452 6.356159l-0.163422-0.197052q30.886646 37.242805-6.356159 68.129452l-387.009893 320.959644q-37.242805 30.886646-68.129452-6.356159l0.163422 0.197052q-30.886646-37.242805 6.356159-68.129452Z"></path>
          <path d="M2.052739 533.200328m30.886646-37.242806l-0.163421 0.197052q30.886646-37.242805 68.129451-6.356159l387.009893 320.959644q37.242805 30.886646 6.356159 68.129452l0.163422-0.197052q-30.886646 37.242805-68.129452 6.356159l-387.009893-320.959644q-37.242805-30.886646-6.356159-68.129452Z"></path>
        </svg>
        <span>返回首页</span>
      </button>
    </div>

    <div class="my-tests-shell">
      <!-- 顶部：个人档案卡片 -->
      <header class="my-tests-header container">
      <div class="my-tests-profile-card">
        <!-- 第一行：头像 + 昵称 + tag + 退出登录 -->
        <div class="my-tests-profile-row my-tests-profile-row--top">
          <!-- 左侧：头像 + 基本信息（只保留昵称 + tag） -->
          <div class="my-tests-profile-left">
            <div class="my-tests-profile-avatar">
              <img
                  v-if="profile?.avatar_url"
                  :src="profile.avatar_url"
                  alt="avatar"
              />
              <span v-else>{{ getAvatarInitial() }}</span>
            </div>

            <div class="my-tests-profile-info">
              <!-- 昵称 + tag -->
              <div class="my-tests-profile-title-row">
                <h1>{{ renderProfileTitle() }}</h1>
                <span class="my-tests-profile-tag">我的测试</span>
              </div>
            </div>
          </div>

          <!-- 右侧：退出登录 -->
          <div class="my-tests-profile-actions-top">
            <button
                type="button"
                class="my-tests-profile-action my-tests-profile-action--logout"
                @click="handleLogout"
            >
              退出登录
            </button>
          </div>
        </div>
        <!-- 第二行：信息类内容 + 修改/保存按钮 -->
        <div class="my-tests-profile-row my-tests-profile-row--bottom">
          <!-- 左侧：信息内容（只读或编辑表单） -->
          <div class="my-tests-profile-bottom-info">
            <!-- 只读：学校 / 地区 -->
            <p
                class="my-tests-profile-sub"
                v-if="!editingExtra"
            >
              <div>{{ profile?.school_name || '-' }}</div>
              <div>{{ profile?.city || profile?.province || '-' }}</div>
            </p>

            <!-- 只读资料字段 -->
            <div
                v-if="!editingExtra"
                class="my-tests-profile-extra-readonly"
            >
              <div>{{ profile?.mobile || '未填写' }}</div>
              <div>{{ profile?.study_id || '未填写' }}</div>
            </div>

            <!-- 编辑模式：表单 -->
            <div
                v-else
                class="my-tests-profile-extra-edit"
            >
              <!-- 家长手机号 -->
              <div class="my-tests-form-field">
                <label class="my-tests-form-label">家长手机号（非必填，建议填写）</label>
                <input
                    type="tel"
                    class="my-tests-form-input"
                    v-model="extraForm.mobile"
                    placeholder="填写家长常用手机号"
                />
              </div>

              <!-- 学号 -->
              <div class="my-tests-form-field">
                <label class="my-tests-form-label">学号（非必填，建议填写）</label>
                <input
                    type="text"
                    class="my-tests-form-input"
                    v-model="extraForm.study_id"
                    placeholder="填写学生学号"
                />
              </div>

              <!-- 学校名称 -->
              <div class="my-tests-form-field">
                <label class="my-tests-form-label">学校名称（非必填，建议填写）</label>
                <input
                    type="text"
                    class="my-tests-form-input"
                    v-model="extraForm.school_name"
                    placeholder="填写所在学校名称"
                />
              </div>

              <!-- 省份下拉 -->
              <div class="my-tests-form-field">
                <label class="my-tests-form-label">所在省份</label>
                <select
                    class="my-tests-form-input"
                    v-model="selectedProvince"
                >
                  <option value="">请选择省份</option>
                  <option
                      v-for="prov in provinces"
                      :key="prov.name"
                      :value="prov.name"
                  >
                    {{ prov.name }}
                  </option>
                </select>
              </div>

              <!-- 市级下拉 -->
              <div class="my-tests-form-field">
                <label class="my-tests-form-label">所在市级</label>
                <select
                    class="my-tests-form-input"
                    v-model="selectedCity"
                    :disabled="!selectedProvince"
                >
                  <option value="">请选择市级</option>
                  <option
                      v-for="city in currentCities"
                      :key="city"
                      :value="city"
                  >
                    {{ city }}
                  </option>
                </select>
              </div>
            </div>
          </div>

          <!-- 右侧：修改 / 取消 / 保存 按钮 -->
          <div class="my-tests-profile-actions-bottom">
            <template v-if="!editingExtra">
              <button
                  type="button"
                  class="btn btn-secondary btn-edit-extra "
                  @click="startEditExtra"
                  aria-label="修改基本信息"
              >
                <svg
                    class="btn-edit-extra__icon"
                    width="16"
                    height="16"
                    viewBox="0 0 16 16"
                    xmlns="http://www.w3.org/2000/svg"
                >
                  <g stroke="none" stroke-width="1" fill="none" fill-rule="evenodd" stroke-linecap="round">
                    <g transform="translate(-1093, -376)" stroke="currentColor">
                      <g transform="translate(1094, 376.9224)">
                        <polyline points="14 4.12441249 14 14.0775784 0 14.0775784 0 0.0775784122 10 0.0775784122"></polyline>
                        <line x1="8.24063621" y1="-1.54594155" x2="8.08572034" y2="12.5909334"
                              transform="translate(8.3871, 5.5) rotate(-136.3461) translate(-8.3871, -5.5)"></line>
                      </g>
                    </g>
                  </g>
                </svg>
              </button>
            </template>
            <template v-else>
              <button
                  type="button"
                  class="btn btn-ghost"
                  @click="cancelEditExtra"
              >
                取消
              </button>
              <button
                  type="button"
                  class="btn btn-primary"
                  @click="saveExtra"
              >
                保存
              </button>
            </template>
          </div>
        </div>
      </div>
    </header>

    <!-- 主体：两个 Tab -->
      <main class="my-tests-main container">
      <div class="my-tests-tabs">
        <button
            type="button"
            class="my-tests-tab"
            :class="{ 'my-tests-tab--active': activeTab === 'ongoing' }"
            @click="setActiveTab('ongoing')"
        >
          进行中的测试
          <span class="my-tests-tab__badge">{{ ongoingList.length }}</span>
        </button>
        <button
            type="button"
            class="my-tests-tab"
            :class="{ 'my-tests-tab--active': activeTab === 'completed' }"
            @click="setActiveTab('completed')"
        >
          已完成的测试
          <span class="my-tests-tab__badge">{{ completedList.length }}</span>
        </button>
      </div>

      <!-- Tab 1：进行中的测试 -->
      <section
          v-if="activeTab === 'ongoing'"
          class="my-tests-section"
      >
        <h2 class="my-tests-section-title">进行中的测试</h2>

        <template v-if="ongoingList.length">
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
                  创建时间：{{ formatDateTime(item.create_at) }}
                </p>

                <p class="my-tests-card-sub">
                  试卷编号：{{ item.public_id }}
                </p>
                <span
                    class="my-tests-status"
                    :class="{
                    'my-tests-status--ongoing': !item.report_status,
                    'my-tests-status--pending': item.report_status === 1
                  }"
                >
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
        </template>

        <template v-else>
          <div class="my-tests-empty">
            <h2>暂无进行中的测试</h2>
            <p>可以先回到首页，选择一套适合你的测评开始体验。</p>
            <button type="button" class="btn" @click="handleBackHome">
              去开始测试
            </button>
          </div>
        </template>
      </section>

      <!-- Tab 2：已完成的测试 -->
      <section
          v-else
          class="my-tests-section"
      >
        <h2 class="my-tests-section-title">已完成的测试与报告</h2>

        <template v-if="completedList.length">
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


                <span class="my-tests-status my-tests-status--done">
                  创建时间：{{ formatDateTime(item.create_at) }}
                 </span>

                <span class="my-tests-status my-tests-status--done">
                  试卷编号：{{ item.public_id }}
               </span>
                <span class="my-tests-status my-tests-status--done">
                  {{ renderStatusText(item) }}
                </span>
              </div>

              <button
                  type="button"
                  class="btn my-tests-card-btn"
                  @click="openReportPreview(item)"
              >
                查看报告
              </button>
            </article>
          </div>
        </template>

        <template v-else>
          <div class="my-tests-empty">
            <h2>暂无已完成的测试</h2>
            <p>完成测评后，你可以在这里查看报告。</p>
          </div>
        </template>
      </section>
    </main>

      <ReportPreviewModal
        v-if="reportPreviewVisible"
        :business-type="reportPreviewTarget?.business_type || ''"
        :public-id="reportPreviewTarget?.public_id || ''"
        @close="closeReportPreview"
    />
    </div>
  </div>
</template>

<script setup lang="ts">
import {useWechatProfile} from '@/controller/WechatProfile'
import ReportPreviewModal from '@/views/components/ReportPreviewModal.vue'

const {
  profile,
  ongoingList,
  completedList,
  renderTitle,
  renderStatusText,
  renderProfileTitle,
  openReportPreview,
  getAvatarInitial,
  formatDateTime,
  handleContinueTest,
  handleBackHome,
  editingExtra,
  extraForm,
  startEditExtra,
  cancelEditExtra,
  saveExtra,
  activeTab,
  setActiveTab,
  provinces,
  selectedProvince,
  selectedCity,
  currentCities,

  reportPreviewVisible,
  reportPreviewTarget,
  closeReportPreview,
  handleLogout,
} = useWechatProfile()

</script>

<style scoped src="@/styles/WechatProfile.css"></style>
