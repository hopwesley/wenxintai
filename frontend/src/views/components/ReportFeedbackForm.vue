<template>
  <div v-if="visible" class="feedback-overlay">
    <div class="feedback-backdrop" @click="handleClose"></div>

    <section class="feedback-dialog" role="dialog" aria-modal="true">
      <!-- 顶部关闭按钮 -->
      <button
          type="button"
          class="feedback-close"
          @click="handleClose"
          aria-label="关闭"
      >
        ×
      </button>

      <!-- 标题 -->
      <header class="feedback-header">
        <h2 class="feedback-title">感谢您使用智择未来</h2>
        <p class="feedback-subtitle">您的反馈对我们非常重要</p>
      </header>

      <!-- 评分区域 -->
      <div class="feedback-body">
        <div class="rating-section">
          <label class="rating-label">您觉得这个报告是否对您的选课有帮助？</label>
          <div class="rating-scale">
            <button
                v-for="score in 11"
                :key="score - 1"
                type="button"
                class="rating-btn"
                :class="{ active: ratingScore === score - 1 }"
                @click="ratingScore = score - 1"
            >
              {{ score - 1 }}
            </button>
          </div>
          <div class="rating-hint">
            <span>完全没帮助</span>
            <span>非常有帮助</span>
          </div>
        </div>

        <!-- 文字反馈区域 -->
        <div class="feedback-section">
          <label class="feedback-label">请分享您的使用体验和建议（可选）</label>
          <textarea
              v-model="feedbackContent"
              class="feedback-textarea"
              placeholder="请分享您的使用体验和建议..."
              rows="5"
          ></textarea>
        </div>
      </div>

      <!-- 按钮区 -->
      <footer class="feedback-footer">
        <button
            type="button"
            class="feedback-btn feedback-btn--ghost"
            @click="handleSkip"
        >
          跳过
        </button>
        <button
            type="button"
            class="feedback-btn feedback-btn--primary"
            :disabled="ratingScore === null"
            @click="handleSubmit"
        >
          提交反馈
        </button>
      </footer>
    </section>
  </div>
</template>

<script setup lang="ts">
import {ref} from 'vue'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'confirm', data: { ratingScore: number; feedbackContent: string }): void
  (e: 'skip'): void
  (e: 'close'): void
}>()

const ratingScore = ref<number | null>(null)
const feedbackContent = ref('')

const handleSubmit = () => {
  if (ratingScore.value === null) return
  emit('confirm', {
    ratingScore: ratingScore.value,
    feedbackContent: feedbackContent.value,
  })
}

const handleSkip = () => {
  emit('skip')
}

const handleClose = () => {
  emit('close')
}
</script>

<style scoped>
.feedback-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.feedback-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(3px);
}

.feedback-dialog {
  position: relative;
  max-width: 600px;
  max-height: 90vh;
  width: 100%;
  background: #fdfbf7;
  border-radius: 16px;
  box-shadow: 0 18px 45px rgba(0, 0, 0, 0.18);
  padding: 24px 28px 20px;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.feedback-close {
  position: absolute;
  top: 12px;
  right: 14px;
  border: none;
  background: transparent;
  font-size: 22px;
  cursor: pointer;
  line-height: 1;
  color: #999;
}

.feedback-close:hover {
  color: #555;
}

.feedback-header {
  margin-bottom: 24px;
}

.feedback-title {
  margin: 0 0 6px;
  font-size: 20px;
  font-weight: 700;
  color: #333;
}

.feedback-subtitle {
  margin: 0;
  font-size: 14px;
  color: #666;
}

.feedback-body {
  flex: 1;
  overflow-y: auto;
  padding-right: 6px;
}

.rating-section {
  margin-bottom: 24px;
}

.rating-label {
  display: block;
  font-size: 15px;
  font-weight: 600;
  color: #333;
  margin-bottom: 12px;
}

.rating-scale {
  display: flex;
  gap: 8px;
  justify-content: space-between;
  margin-bottom: 8px;
}

.rating-btn {
  flex: 1;
  min-width: 40px;
  height: 40px;
  border: 2px solid #ddd;
  background: #fff;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #666;
  cursor: pointer;
  transition: all 0.2s ease;
}

.rating-btn:hover {
  border-color: #7b5cff;
  color: #7b5cff;
}

.rating-btn.active {
  border-color: #7b5cff;
  background: #7b5cff;
  color: #fff;
}

.rating-hint {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #999;
  padding: 0 4px;
}

.feedback-section {
  margin-bottom: 16px;
}

.feedback-label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
}

.feedback-textarea {
  width: 100%;
  padding: 12px;
  border: 2px solid #ddd;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  color: #333;
  resize: vertical;
  transition: border-color 0.2s ease;
  box-sizing: border-box;
}

.feedback-textarea:focus {
  outline: none;
  border-color: #7b5cff;
}

.feedback-textarea::placeholder {
  color: #aaa;
}

.feedback-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.feedback-btn {
  min-width: 100px;
  padding: 10px 20px;
  border-radius: 999px;
  font-size: 14px;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
  font-weight: 600;
}

.feedback-btn--ghost {
  background: transparent;
  color: #666;
  border: 1px solid #ddd;
}

.feedback-btn--ghost:hover {
  background: #f2f2f2;
}

.feedback-btn--primary {
  background: #7b5cff;
  color: #fff;
}

.feedback-btn--primary:hover:not(:disabled) {
  background: #6543ff;
}

.feedback-btn--primary:disabled {
  background: #ccc;
  cursor: not-allowed;
}

@media (max-width: 600px) {
  .feedback-dialog {
    padding: 20px 18px 16px;
    border-radius: 12px;
  }

  .feedback-title {
    font-size: 18px;
  }

  .rating-scale {
    gap: 4px;
  }

  .rating-btn {
    min-width: 32px;
    height: 36px;
    font-size: 13px;
  }

  .feedback-btn {
    min-width: 80px;
    padding: 8px 16px;
  }
}
</style>
