<template>
  <teleport to="body">
    <div v-if="open" class="invite-mask">
      <div class="invite-dialog" role="dialog" aria-modal="true">
        <button class="close-btn" type="button" @click="handleCancel" aria-label="close">
          ×
        </button>

        <!-- 顶部：产品信息 -->
        <div class="plan-section">
          <h3 class="title">确认测试方案</h3>
          <p class="plan-name">{{ product?.name }}</p>
          <p class="plan-price">￥{{ displayPrice }}</p>
          <p v-if="product?.desc" class="plan-desc">
            {{ product?.desc }}
          </p>
        </div>

        <div class="pay-section">
          <!-- 还没创建订单时，显示按钮 -->
          <button
              v-if="!payOrder"
              type="button"
              class="btn btn-primary pay-btn"
              :disabled="paying"
              @click="handleWeChatPayClick"
          >
            <span v-if="paying">创建订单中…</span>
            <span v-else>微信扫码支付并开始测试</span>
          </button>

          <!-- 已经有订单（有 code_url），显示二维码 -->
          <div v-else class="qrcode-wrapper">
            <!-- 这里用任何二维码库都行，比如 qrcode.vue；演示用 <img> 调第三方服务 -->
            <img
                class="qrcode-img"
                :src="`https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=${encodeURIComponent(payOrder.code_url)}`"
                alt="微信支付二维码"
            />
            <p class="qrcode-tip">请使用微信“扫一扫”完成支付</p>
          </div>

          <p class="pay-hint">支付成功后，将自动进入测试</p>
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
import {computed, nextTick, ref, watch} from 'vue'
import {verifyInviteWithMessage} from '@/controller/PaymentModal'
import {NativeCreateOrderResponse} from "@/controller/PaymentModal";
import {PlanInfo} from "@/controller/common";
import {useAlert} from "@/controller/useAlert";


const props = defineProps<{
  open: boolean
  product: PlanInfo | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

function handleWeChatPayClick() {
}

// 邀请码相关状态
const code = ref('')
const inviteLoading = ref(false)
const errorMessage = ref('')
const inputRef = ref<HTMLInputElement | null>(null)

const payOrder = ref<NativeCreateOrderResponse | null>(null)
const paying = ref(false)

const trimmedCode = computed(() => code.value.trim())
const payLoading = ref(false)

const displayPrice = computed(() => {
  return props.product?.price.toFixed(2)
})
const {showAlert} = useAlert()

watch(
    () => props.open,
    async (isOpen) => {
      if (isOpen) {
        await nextTick()
        reset()
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
  showAlert("您确定放弃本次测试报告吗？",()=>{
    emit('update:open', false)
  })
}

// 提交邀请码
async function handleInviteConfirm() {
  if (inviteLoading.value) return

  inviteLoading.value = true
  errorMessage.value = ''
  try {
    inviteLoading.value = false
    const trimmed = code.value.trim()

    await verifyInviteWithMessage(trimmed)
    emit('success')
  } catch (e) {
    errorMessage.value = JSON.stringify(e)
    inputRef.value?.focus()
    return
  } finally {

  }

}
</script>

<style scoped src="@/styles/payment_code.css"></style>
