<template>
  <div class="questions-redirect">
    <p>{{ message }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '@/i18n'

const router = useRouter()
const { t } = useI18n()

const message = computed(() => t('loading.default'))

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
