<template>
  <div class="questions-redirect">
    <p>{{ message }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const message = computed(() =>  '加载中…')

onMounted(() => {
  router.replace('/test/basic/step/2')
})
</script>

<style scoped>
.questions-redirect {
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(30, 41, 59, 0.7);
}
</style>
