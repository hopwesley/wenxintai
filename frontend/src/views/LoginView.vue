<template>
  <div class="login">
    <h1>登录</h1>
    <form @submit.prevent="submit">
      <div>
        <label>
          微信 ID：
          <input v-model="wechatId" required />
        </label>
      </div>
      <div>
        <label>
          昵称：
          <input v-model="nickname" />
        </label>
      </div>
      <div>
        <label>
          头像地址：
          <input v-model="avatarUrl" />
        </label>
      </div>
      <button type="submit">提交</button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '@/api'

const router = useRouter()
const wechatId = ref('')
const nickname = ref('')
const avatarUrl = ref('')

async function submit() {
  try {
    const result = await login({
      wechat_id: wechatId.value,
      nickname: nickname.value,
      avatar_url: avatarUrl.value
    })
    // Save session ID somewhere (e.g., localStorage) for subsequent API calls
    localStorage.setItem('session_id', result.session_id)
    router.push('/hobbies')
  } catch (e) {
    alert((e as Error).message)
  }
}
</script>