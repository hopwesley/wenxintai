<template>
  <Teleport to="body">
    <Transition name="alert-fade">
      <div
          v-if="visible"
          class="alert-overlay"
          @click.self="closeAlert"
      >
        <div class="alert-dialog">
          <button class="alert-close" type="button" @click="closeAlert">
            ×
          </button>

          <h3 v-if="title" class="alert-title">
            {{ title }}
          </h3>

          <div class="alert-message">
            <slot>
              {{ message }}
            </slot>
          </div>

          <div class="alert-actions">
            <button
                type="button"
                class="btn btn-primary alert-confirm"
                @click="handleConfirm"
            >
              确定
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>


<script setup lang="ts">
import '@/styles/base.css'
import { useAlert } from '@/controller/useAlert'
const { visible, title, message, handleConfirm, closeAlert } = useAlert()
</script>

<style scoped>
.alert-overlay {
  position: fixed;
  inset: 0;
  z-index: 2200;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.45); /* 半透明深色遮罩 */
  backdrop-filter: blur(4px);
  --brand: #5A60EA;
  --brand-dark: #484DBB;
}

/* 弹窗本体 */
.alert-dialog {
  position: relative;
  max-width: 420px;
  width: 90%;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 18px 45px rgba(15, 23, 42, 0.25);
  padding: 20px 24px 18px;
  box-sizing: border-box;
}

/* 右上角关闭按钮 */
.alert-close {
  position: absolute;
  top: 10px;
  right: 12px;
  border: none;
  background: transparent;
  font-size: 20px;
  line-height: 1;
  cursor: pointer;
  color: #64748b;
}

.alert-close:hover {
  color: #0f172a;
}

.alert-title {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: #0f172a;
  text-align: center;
}

.alert-message {
  font-size: 14px;
  line-height: 1.6;
  color: #475569;
  margin-bottom: 16px;
  text-align: center;
}

/* 底部按钮区域 */
.alert-actions {
  display: flex;
  justify-content: center;
}

/* 如果你项目里已经有 .btn / .btn-primary，可以不需要这些 fallback 样式 */
.alert-confirm {
  min-width: 88px;
  background-color: var(--brand);
}

</style>