<template>
  <div v-if="open" class="new-user-dialog-backdrop">
    <div class="new-user-dialog">
      <h2 class="new-user-dialog__title">完善基础信息</h2>
      <p class="new-user-dialog__desc">
        这里是新用户首次登录后补充基础信息的界面（当前为占位，你之后可以加表单）。
      </p>

      <label class="new-user-dialog__checkbox">
        <input type="checkbox" v-model="dontRemind" />
        <span>下次不再提醒</span>
      </label>

      <div class="new-user-dialog__footer">
        <button
            type="button"
            class="new-user-dialog__btn"
            @click="handleConfirm"
        >
          完成
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref} from 'vue'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'never-remind'): void
}>()

const dontRemind = ref(false)

function handleConfirm() {
  if (dontRemind.value) {
    emit('never-remind')   // ✅ 外面可以据此写 localStorage
  }
  emit('update:open', false)
}
</script>

<style scoped>
.new-user-dialog-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2100;
}

.new-user-dialog {
  width: min(520px, 90vw);
  border-radius: 16px;
  background: #fff;
  padding: 24px 24px 20px;
  box-shadow: 0 18px 45px rgba(15, 23, 42, 0.18);
}

.new-user-dialog__title {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 600;
  color: #111827;
}

.new-user-dialog__desc {
  margin: 0 0 24px;
  font-size: 14px;
  line-height: 1.6;
  color: #4b5563;
}

.new-user-dialog__close-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  font-size: 14px;
  color: #374151;
  cursor: pointer;
}

.new-user-dialog__close-btn:hover {
  background: #f3f4f6;
}
</style>
