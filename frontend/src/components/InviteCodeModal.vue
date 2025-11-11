<template>
  <teleport to="body">
    <div v-if="open" class="invite-mask" @click.self="handleCancel">
      <div class="invite-dialog" role="dialog" aria-modal="true">
        <button class="close-btn" type="button" @click="handleCancel" aria-label="close">
          ×
        </button>
        <h3 class="title">请输入邀请码</h3>
        <p class="description">每个邀请码仅可使用一次，请确认后提交。</p>
        <form class="form" @submit.prevent="handleConfirm">
          <input
              ref="inputRef"
              v-model="code"
              class="code-input"
              type="text"
              placeholder='邀请码'
              :disabled="loading"
              autocomplete="one-time-code"
          />
          <p class="error-message" role="alert">{{ errorMessage || '\u00A0' }}</p>
          <div class="actions">
            <button
                type="button"
                class="btn btn-secondary"
                @click="handleCancel"
                :disabled="loading"
            >
              取消
            </button>
            <button
                type="submit"
                class="btn btn-primary"
                :disabled="loading || trimmedCode.length === 0"
            >
              <span v-if="loading">验证中…</span>
              <span v-else>确认开始</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { verifyInvite } from '@/api'
import { useTestSession } from '@/store/testSession'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ (e: 'update:open', value: boolean): void; (e: 'success'): void }>()

const code = ref('')
const loading = ref(false)
const errorMessage = ref('')
const inputRef = ref<HTMLInputElement | null>(null)

const trimmedCode = computed(() => code.value.trim())

const { getSessionId, setSessionId } = useTestSession()

watch(
  () => props.open,
  async (isOpen) => {
    if (isOpen) {
      await nextTick()
      errorMessage.value = ''
      loading.value = false
      inputRef.value?.focus()
    } else {
      reset()
    }
  }
)

function reset() {
  code.value = ''
  loading.value = false
  errorMessage.value = ''
}

function handleCancel() {
  emit('update:open', false)
}

async function handleConfirm() {
  if (loading.value) {
    return
  }
  if (trimmedCode.value.length === 0) {
    errorMessage.value = '请输入邀请码'
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    const sessionId = getSessionId() ?? undefined
    const response = await verifyInvite(trimmedCode.value, sessionId)
    if (response && typeof response.session_id === 'string') {
      setSessionId(response.session_id)
    }
    emit('update:open', false)
    emit('success')
  } catch (error) {
    console.error('[InviteCodeModal] verify failed', error)
    if (error instanceof Error) {
      errorMessage.value = error.message
    } else {
      errorMessage.value = '验证失败，请稍后再试'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.invite-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2100;
  backdrop-filter: blur(2px);
}

.invite-dialog {
  position: relative;
  width: 360px;
  background: #fff;
  border-radius: 18px;
  padding: 28px 26px 32px;
  box-shadow: 0 16px 40px rgba(15, 23, 42, 0.18);
  text-align: left;
}

.title {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 600;
  color: #0f172a;
}

.description {
  margin: 0 0 18px;
  font-size: 14px;
  color: #475569;
  line-height: 1.6;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.code-input {
  width: 100%;
  height: 44px;
  border-radius: 12px;
  border: 1px solid #cbd5f5;
  padding: 0 14px;
  font-size: 15px;
  outline: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.code-input:focus {
  border-color: #5b7cff;
  box-shadow: 0 0 0 3px rgba(91, 124, 255, 0.2);
}

.code-input:disabled {
  background: #f1f5f9;
}

.error-message {
  min-height: 18px;
  margin: -4px 0 0;
  font-size: 13px;
  color: #ef4444;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 6px;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 18px;
  height: 40px;
  border-radius: 12px;
  border: none;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.2s ease;
}

.btn:disabled {
  cursor: not-allowed;
  opacity: 0.7;
  transform: none;
  box-shadow: none;
}

.btn-primary {
  background: linear-gradient(90deg, #7aa2ff, #5b7cff);
  color: #fff;
  box-shadow: 0 12px 28px rgba(91, 124, 255, 0.28);
}

.btn-primary:not(:disabled):hover {
  transform: translateY(-1px);
}

.btn-secondary {
  background: #e2e8f0;
  color: #1e293b;
}

.close-btn {
  position: absolute;
  top: 12px;
  right: 14px;
  border: none;
  background: transparent;
  font-size: 22px;
  color: #94a3b8;
  cursor: pointer;
}

.close-btn:hover {
  color: #475569;
}
</style>
