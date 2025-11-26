<template>
  <teleport to="body">
    <div v-if="open" class="wechat-login-mask" @click.self="close">
      <img
          class="wechat-logo"
          src="/img/logo.png"
          alt="WeChat"
      />
      <div class="wechat-login-welcome">
        <div class="left-img"></div>

        <div class="wechat-login-dialog">
          <button class="close-btn" @click="close">×</button>
          <div class="wechat-login-content">
            <h3>微信扫码登录</h3>

            <div class="qrcode-box">
              <!-- 这里不再用 img，而是交给 WxLogin.js 渲染 iframe -->
              <div id="wx-login-qrcode" class="wx-qrcode-container"></div>
            </div>

            <p class="desc">
              注册登录即代表同意
              <span>用户服务协议、隐私政策、会员服务协议、授权许可协议</span>
            </p>

            <p
                v-if="authStore.loginStatus === 'pending'"
                class="status"
            >
              请使用微信扫一扫完成登录…
            </p>
            <p
                v-else-if="authStore.loginStatus === 'expired'"
                class="status status-error"
            >
              二维码已过期，请点击下方按钮重新获取。
            </p>
            <p
                v-else-if="authStore.loginStatus === 'error'"
                class="status status-error"
            >
              登录出现异常，请稍后重试。
            </p>

            <button
                v-if="authStore.loginStatus === 'expired'"
                type="button"
                class="refresh-btn"
                @click="authStore.startWeChatLogin()"
            >
              重新获取二维码
            </button>
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
  display: grid;
  grid-template-columns: 1fr 1fr; /* 两个横向等分的板块 */
  align-items: stretch;           /* 两块都铺满高度 */
  justify-items: stretch;
  gap: 0;                         /* 需要间距可改为 24px */
  z-index: 2000;
  animation: fadeInUp-d3366225 0.25s
  ease-out;
}

.left-img{
  width: 459px;
  height: 551px;
  background-image: url('/img/login-img.png');
  background-position: center;
  background-size: cover;
  background-repeat: no-repeat;
}

.wechat-login-dialog {
  position: relative;
  width: 450px;
  background: #fff;
  border-radius: 0 48px 48px 0;
  padding: 83px 110px;
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
.wechat-login-content span {
  font-size: 14px;
  font-weight: 600;
  color: #767678;
}

.qrcode-box {
  width: 240px;
  height: 240px;
  margin: 16px auto;
  border-radius: 12px;
  border: 1px solid #EAEAEA;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  padding: 16px;
}

/* 新增：给 WxLogin 渲染的 iframe 提供容器 */
.wx-qrcode-container {
  width: 100%;
  height: 100%;
}

/* 原来的 qrcode-box img 可以删除 */

.status {
  margin: 8px 0 0;
  font-size: 13px;
  color: #BAB2B2;
}

.status-error {
  color: #EF4444;
}

.refresh-btn {
  margin-top: 12px;
  padding: 6px 16px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  background: #16a34a;
  color: #fff;
  font-size: 14px;
}

.refresh-btn:hover {
  background: #15803d;
}

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

