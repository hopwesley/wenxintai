<template>
  <div class="questions">
    <h1>答题</h1>
    <div v-if="!questionsLoaded">加载中…</div>
    <form v-else @submit.prevent="submit">
      <div v-for="(q, idx) in questions" :key="idx" class="question">
        <p>{{ idx + 1 }}. {{ q.text }}</p>
        <div class="options" role="radiogroup">
          <label v-for="v in [1,2,3,4,5]" :key="v">
            <input type="radio" :name="'q' + idx" :value="v" v-model.number="answers[idx]" />
            <span>{{ v }}</span>
          </label>
        </div>
      </div>
      <button type="submit">提交答案</button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getQuestions, sendAnswers } from '../api'

const router = useRouter()
const questions = ref<any[]>([])
const answers = reactive<number[]>([])
const questionsLoaded = ref(false)

onMounted(async () => {
  const sessionId = localStorage.getItem('session_id') || ''
  const hobby = localStorage.getItem('hobby') || ''
  try {
    const resp = await getQuestions({
      session_id: sessionId,
      mode: 'A',
      gender: '',
      grade: '',
      hobby
    })
    questions.value = resp.questions.questions || []
    answers.splice(0, answers.length, ...Array(questions.value.length).fill(3))
    questionsLoaded.value = true
  } catch (e) {
    alert((e as Error).message)
  }
})

async function submit() {
  const sessionId = localStorage.getItem('session_id') || ''
  try {
    const riasecAnswers: any[] = []
    const ascAnswers: any[] = []
    // API expects separate arrays; map into appropriate structures here.
    // For demonstration we leave these empty; implement as needed.
    await sendAnswers({
      session_id: sessionId,
      mode: 'A',
      riasec_answers: riasecAnswers,
      asc_answers: ascAnswers
    })
    router.push('/summary')
  } catch (e) {
    alert((e as Error).message)
  }
}
</script>