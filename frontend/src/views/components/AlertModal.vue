<template>
  <Teleport to="body">
    <Transition name="alert-fade">
      <div
          v-if="modelValue"
          class="alert-overlay"
          @click.self="close"
      >
        <div class="alert-dialog">
          <button class="alert-close" type="button" @click="close">
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
                @click="onConfirm"
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
const props = defineProps<{
  modelValue: boolean
  title?: string
  message?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'confirm'): void
}>()

function close() {
  emit('update:modelValue', false)
}

function onConfirm() {
  emit('confirm')
  close()
}
</script>

<style scoped>
.alert-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.45); /* 半透明深色遮罩 */
  backdrop-filter: blur(4px);
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
}

.alert-message {
  font-size: 14px;
  line-height: 1.6;
  color: #475569;
  margin-bottom: 16px;
}

/* 底部按钮区域 */
.alert-actions {
  display: flex;
  justify-content: flex-end;
}

/* 如果你项目里已经有 .btn / .btn-primary，可以不需要这些 fallback 样式 */
.alert-confirm {
  min-width: 88px;
}

/* 过渡动画 */
.alert-fade-enter-active,
.alert-fade-leave-active {
  transition: opacity 0.18s ease-out;
}

.alert-fade-enter-from,
.alert-fade-leave-to {
  opacity: 0;
}
</style>
