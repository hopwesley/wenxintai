<template>
  <div class="hobbies">
    <h1>选择爱好</h1>
    <ul>
      <li v-for="h in hobbies" :key="h">
        <label>
          <input type="radio" v-model="selected" :value="h" />
          {{ h }}
        </label>
      </li>
    </ul>
    <button :disabled="!selected" @click="next">下一步</button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getHobbies } from '@/api'

const router = useRouter()
const hobbies = ref<string[]>([])
const selected = ref<string>('')

onMounted(async () => {
  try {
    hobbies.value = await getHobbies()
  } catch (e) {
    alert((e as Error).message)
  }
})

function next() {
  // Store selected hobby in localStorage for later use when requesting questions
  localStorage.setItem('hobby', selected.value)
  router.push('/questions')
}
</script>