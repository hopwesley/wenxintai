<template>
  <teleport to="body">
    <div v-if="open" class="invite-mask" @click.self="handleCancel">
      <div class="invite-dialog" role="dialog" aria-modal="true">
        <button class="close-btn" type="button" @click="handleCancel" aria-label="close">
          ×
        </button>

        <!-- 顶部：产品信息 -->
        <div class="plan-section">
          <h3 class="title">确认测试方案</h3>
          <p class="plan-name">{{ productName }}</p>
          <p class="plan-price">￥{{ displayPrice }}</p>
          <p v-if="productDesc" class="plan-desc">
            {{ productDesc }}
          </p>
        </div>

        <!-- 微信支付主操作 -->
        <div class="pay-section">
          <button
              type="button"
              class="btn btn-primary pay-btn"
              :disabled="payLoading"
              @click="handleWeChatPay"
          >
            <span v-if="payLoading">唤起微信支付…</span>
            <span v-else>微信支付并开始测试</span>
          </button>
          <p class="pay-hint">使用微信完成支付后，将自动进入测试</p>
        </div>

        <!-- 分隔 -->
        <div class="divider">
          <span class="divider-line"></span>
          <span class="divider-text">或使用邀请码免费体验</span>
          <span class="divider-line"></span>
        </div>

        <!-- 邀请码区域 -->
        <form class="form" @submit.prevent="handleInviteConfirm">
          <div class="invite-row">
            <input
                ref="inputRef"
                v-model="code"
                class="code-input"
                type="text"
                placeholder="输入邀请码"
                :disabled="inviteLoading || payLoading"
                autocomplete="one-time-code"
            />
            <button
                type="submit"
                class="btn btn-secondary invite-btn"
                :disabled="inviteLoading || trimmedCode.length === 0"
            >
              <span v-if="inviteLoading">验证中…</span>
              <span v-else>使用邀请码</span>
            </button>
          </div>
          <p :class="['error-message', { visible: !!errorMessage }]" role="alert">
            {{ errorMessage }}
          </p>
        </form>
      </div>
    </div>
  </teleport>
</template>


<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { VerifyInviteResponse, verifyInviteWithMessage } from '@/controller/InviteCode'
import { useTestSession } from '@/controller/testSession'

const { setInviteCode } = useTestSession()

const props = defineProps<{
  open: boolean
  productName: string
  productPrice: number // 假设单位“元”，如果是分，这里注意换算
  productDesc?: string
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success', payload: VerifyInviteResponse): void
  (e: 'pay'): void
}>()

// 邀请码相关状态
const code = ref('')
const inviteLoading = ref(false)
const errorMessage = ref('')
const inputRef = ref<HTMLInputElement | null>(null)
const trimmedCode = computed(() => code.value.trim())

// 支付按钮的 loading 状态
const payLoading = ref(false)

const displayPrice = computed(() => {
  // 如果 productPrice 是“元”，可以直接展示；如果你用“分”，这里改成 (productPrice / 100).toFixed(2)
  return props.productPrice.toFixed(2)
})

watch(
    () => props.open,
    async (isOpen) => {
      if (isOpen) {
        await nextTick()
        reset()
        // 默认焦点先不给邀请码，主路径是支付
      } else {
        reset()
      }
    }
)

function reset() {
  code.value = ''
  inviteLoading.value = false
  payLoading.value = false
  errorMessage.value = ''
}

function handleCancel() {
  emit('update:open', false)
}

// 点击微信支付按钮
function handleWeChatPay() {
  if (payLoading.value) return
  payLoading.value = true
  errorMessage.value = ''

  // 把真正的支付逻辑交给父组件
  emit('pay')

  // 父组件可以在支付完成或失败后，通过 v-model:open 关闭 / 重开弹窗
  // 或者你也可以后面扩展一个事件，通知支付完成再重置 payLoading
}

// 提交邀请码
async function handleInviteConfirm() {
  if (inviteLoading.value) return

  inviteLoading.value = true
  errorMessage.value = ''

  const { ok, errorMessage: msg, response } = await verifyInviteWithMessage(code.value)

  inviteLoading.value = false

  if (!ok) {
    if (msg) {
      errorMessage.value = msg
    }
    // 自动聚焦输入框方便重试
    inputRef.value?.focus()
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
