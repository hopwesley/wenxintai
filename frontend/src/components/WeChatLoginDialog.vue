<template>
  <teleport to="body">
    <!-- 蒙层，点击空白区域可关闭 -->
    <div v-if="open" class="wechat-login-mask" @click.self="close">
      <div class="wechat-login-dialog">
        <!-- 右上角关闭按钮 -->
        <button class="close-btn" @click="close">×</button>

        <!-- 主体内容 -->
        <div class="wechat-login-content">
          <img
              class="wechat-logo"
              src="/img/logo.png"
              alt="WeChat"
          />
          <h3>微信扫码登录</h3>
          <p class="desc">请使用微信扫描二维码登录系统</p>

          <!-- 模拟二维码区域 -->
          <div class="qrcode-box">
            <img
                src="/img/logo.png"
                alt="WeChat QR Code"
            />
          </div>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
/**
 * 外部用法：
 * <WeChatLoginDialog v-model:open="showLogin" />
 */
const props = defineProps({
  open: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['update:open'])

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

.wechat-login-dialog {
  position: relative;
  width: 340px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.25);
  padding: 30px 24px 36px;
  text-align: center;
  color: #333;
  animation: fadeInUp 0.25s ease-out;
}

.wechat-login-content h3 {
  margin: 12px 0 6px;
  font-size: 20px;
  font-weight: 600;
}

.wechat-login-content .desc {
  margin: 0 0 16px;
  color: #666;
  font-size: 14px;
}

.qrcode-box {
  width: 200px;
  height: 200px;
  margin: 0 auto;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
}

.qrcode-box img {
  width: 160px;
  height: 160px;
}

.close-btn {
  position: absolute;
  right: 10px;
  top: 10px;
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

