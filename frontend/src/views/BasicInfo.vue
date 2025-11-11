<template>
  <div class="basic-info-page">
    <h1 class="title">填写基本信息</h1>

    <form class="form" @submit.prevent="submit">
      <label class="field">
        <span>姓名（可选）</span>
        <input v-model="name" type="text" />
      </label>

      <label class="field">
        <span>学校（可选）</span>
        <input v-model="school" type="text" />
      </label>

      <label class="field">
        <span>年级（可选）</span>
        <input v-model="grade" type="text" />
      </label>

      <label class="field">
        <span>年龄（可选）</span>
        <input v-model="age" type="number" />
      </label>

      <label class="field">
        <span>性别（可选）</span>
        <select v-model="gender">
          <option value="">未选择</option>
          <option value="male">男</option>
          <option value="female">女</option>
        </select>
      </label>

      <button class="btn btn-primary" type="submit">
        下一步
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const assessmentId = route.params.assessmentId as string

const name = ref('')
const school = ref('')
const grade = ref('')
const age = ref('')
const gender = ref('')

async function submit() {
  // 这里暂时不调用任何后端接口，因为你说 BasicInfo 提交功能还没开发
  // 下一步直接进入测试占位即可
  router.push('/test/basic/step/1')
}
</script>

<style scoped>
.basic-info-page {
  max-width: 500px;
  margin: 40px auto;
  padding: 20px;
}

.title {
  font-size: 24px;
  margin-bottom: 20px;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.field {
  display: flex;
  flex-direction: column;
}

input,
select {
  height: 36px;
  padding: 8px;
  border-radius: 6px;
  border: 1px solid #ccc;
}

.btn-primary {
  background: #5b7cff;
  color: #fff;
  padding: 10px;
  border-radius: 6px;
  border: none;
}
</style>
