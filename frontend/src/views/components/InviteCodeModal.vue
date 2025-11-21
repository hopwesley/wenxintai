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
          <p :class="['error-message', { visible: !!errorMessage }]" role="alert">{{ errorMessage }}</p>
          <div class="actions">
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
import {computed, nextTick, ref, watch} from 'vue'
import {VerifyInviteResponse, verifyInviteWithMessage} from '@/controller/InviteCode'
import {useTestSession} from '@/controller/testSession'

const {setInviteCode} = useTestSession()

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success', payload: VerifyInviteResponse): void
}>()

const code = ref('')
const loading = ref(false)
const errorMessage = ref('')
const inputRef = ref<HTMLInputElement | null>(null)
const trimmedCode = computed(() => code.value.trim())

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

  loading.value = true
  errorMessage.value = ''

  const {ok, errorMessage: msg, response} = await verifyInviteWithMessage(code.value)

  loading.value = false

  if (!ok) {
    if (msg) {
      errorMessage.value = msg
    }
    return
  }

  const trimmed = code.value.trim()
  setInviteCode(trimmed)
  emit('update:open', false)

  if (response) {
    emit('success', response)
  }
}

</script>

<style scoped src="@/styles/invite_code.css"></style>
