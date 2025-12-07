<template>
  <teleport to="body">
    <div v-if="open" class="wechat-login-mask" @click.self="close">

      <div class="wechat-login-welcome">
        <div class="wechat-login-left">
          <img src="/img/login-img.png" alt="登录插画" />
        </div>
        <div class="wechat-login-dialog">
          <div class="wechat-login-dialog-inner">
          <button class="close-btn" @click="close">×</button>
          <div class="wechat-login-content">
            <img
                class="wechat-logo"
                src="/img/logo.png"
                alt="WeChat"
            />
            <div id="wx-login-qrcode" style="margin-top: 24px"></div>
            <div class="desc">
              注册/登录即代表你同意
              <a href="/agreements/user" target="_blank">《用户服务协议》</a>、
              <a href="/agreements/privacy" target="_blank">《隐私政策》</a>、
              <a href="/agreements/license" target="_blank">《授权许可协议》</a>
            </div>
          </div>
          </div>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/controller/wx_auth'
defineProps<{
  open: boolean
}>();
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

// 直接拿全局 authStore 来展示状态 + 重新获取二维码
const authStore = useAuthStore()

function close() {
  authStore.cancelWeChatLogin()
  emit('update:open', false)
}
</script>

<style scoped>
.wechat-login-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(2px);
}

.wechat-login-welcome {
  height: 552px;
  width: 60%;
  border-radius: 48px;
  /* 1. 响应式宽度 + 上下限 */
  max-width: 960px;     /* 最大就这么宽，超过不再变宽 */
  min-width: 640px;     /* 最小就这么宽，再小就不缩了 */

  display: grid;
  grid-template-columns: 1fr 1fr;
  align-items: stretch;           /* 两块都铺满高度 */
  justify-items: stretch;
  gap: 0;                         /* 需要间距可改为 24px */
  z-index: 2000;
  animation: fadeInUp-d3366225 0.25s
  ease-out;
}

/* 左侧图片区域 */
.wechat-login-left {
  width: 100%;
  height: 100%;
  overflow: hidden;
  border-radius: 48px 0 0 48px;
}

.wechat-login-left img {
  width: 100%;
  height: 100%;
  object-fit: cover;      /* 自适应裁剪填满，效果类似你截图 */
  display: block;
}

.wechat-login-dialog {
  position: relative;
  width: 100%;
  background: #fff;
  border-radius: 0 48px 48px 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 右侧内层，用来做 padding 和布局 */
.wechat-login-dialog-inner {
  padding: 72px;    /* 原来那一圈大 padding 挪到这里 */
  text-align: center;
  color: #333;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.wechat-login-content h3 {
  margin: 12px 0 6px;
  font-size: 20px;
  font-weight: 600;
  color: #1D1D20;
  line-height: 25px;
}

.wechat-login-content .desc {
  margin: 0 0 16px;
  font-size: 14px;
  font-weight: 400;
  color: #BAB2B2;
}

.wechat-login-content .desc a {
  text-decoration: none;
  color: #333; /* 示例：设置为深灰色 */
  outline: none;
}
.wechat-login-content .desc a:hover {
  color: var(--brand); /* 示例：悬停时变为红色 */
  text-decoration: underline; /* 悬停时重新添加下划线 */
  opacity: 0.8; /* 稍微透明化 */
}
.wechat-login-content span {
  font-size: 14px;
  font-weight: 600;
  color: #767678;
}

/* 原来的 qrcode-box img 可以删除 */

.close-btn {
  position: absolute;
  right: 24px;
  top: 24px;
  border: none;
  background: transparent;
  font-size: 22px;
  cursor: pointer;
  color: #888;
}

.close-btn:hover {
  color: #000;
}

.wechat-logo {
  width: 50px;
  height: 50px;
  position: absolute;
  top: 32px;
  left: calc(50% - 25px);
}

@keyframes fadeInUp {
  from {
    transform: translateY(10px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
</style>

