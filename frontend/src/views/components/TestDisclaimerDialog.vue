<template>
  <div
      v-if="visible"
      class="td-backdrop"
      @click.self="handleCancel"
  >
    <div class="td-dialog">
      <header class="td-header">
        <h2 class="td-title">测试前须知与免责声明</h2>
        <p class="td-subtitle">
          在开始正式测试之前，请先仔细阅读以下内容。
        </p>
      </header>

      <section class="td-body">
        <!-- 第一部分：作答方式与态度 -->
        <section class="td-section">
          <h3 class="td-section-title">一、请你全然为自己作答</h3>
          <p>
            请你全然为自己作答。用你的热爱、好奇、直觉与真诚，去选择每一个答案。
          </p>
          <p>
            你现在的每一次选择，都是在为自己擘画真正属于你的未来方向，而不是迎合任何外在标准。
          </p>
        </section>

        <!-- 第二部分：隐私与数据安全 -->
        <section class="td-section">
          <h3 class="td-section-title">二、关于你的隐私与数据安全</h3>
          <p>
            你的信息将受到我们最严密的守护。我们会谨慎处理与你有关的所有数据。
          </p>
          <p>
            你的隐私与安全，是「智择未来」不可动摇的承诺。
          </p>
        </section>

        <!-- 第三部分：测评定位与适用范围 -->
        <section class="td-section">
          <h3 class="td-section-title">三、本次测评的定位与边界</h3>
          <p>
            请你信任这场探索。「智择未来」将经典心理学智慧与前沿 AI 算法相结合，为你呈现独属于你的个性化选科策略。
          </p>
          <p>
            然而，本测评工具仅作为学习规划与自我探索的
            <span class="td-strong">辅助手段</span>，
            不构成升学、志愿填报或职业选择的唯一依据。
          </p>
          <p>
            测评结果基于你当前提供的信息与答题情况生成，平台不对因错误或不完整信息导致的结果偏差承担责任。
          </p>
        </section>

        <!-- 结尾强调 -->
        <section class="td-section td-section-highlight">
          <p class="td-highlight-main">
            继续开始测试，即表示你已阅读并同意本声明。
          </p>
          <p class="td-highlight-sub">
            「智择未来」团队愿与你勤谨同行。
          </p>
        </section>
      </section>

      <footer class="td-footer">
        <button
            type="button"
            class="td-btn td-btn-ghost"
            @click="handleCancel"
        >
          取消
        </button>
        <button
            type="button"
            class="td-btn td-btn-primary"
            @click="handleConfirm"
        >
          我已知晓并同意，开始测试
        </button>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

function handleConfirm() {
  emit('confirm')
}

function handleCancel() {
  emit('cancel')
}
</script>

<style scoped>
.td-backdrop {
  position: fixed;
  inset: 0;
  z-index: 999;
  background: rgba(15, 23, 42, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.td-dialog {
  width: 520px;
  max-width: 100%;
  max-height: 80vh;
  background: #ffffff;
  border-radius: 20px;
  box-shadow: 0 18px 45px rgba(15, 23, 42, 0.18);
  padding: 22px 26px 20px;
  display: flex;
  flex-direction: column;
}

.td-header {
  margin-bottom: 10px;
}

.td-title {
  margin: 0 0 6px;
  font-size: 20px;
  font-weight: 600;
  color: #0f172a;
}

.td-subtitle {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

.td-body {
  margin-top: 8px;
  margin-bottom: 20px;
  font-size: 14px;
  line-height: 1.7;
  color: #475569;
  overflow-y: auto;
}

/* 分段样式 */
.td-section {
  margin-bottom: 14px;
}

.td-section:last-of-type {
  margin-bottom: 0;
}

.td-section-title {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 600;
  color: #111827;
}

.td-section p {
  margin: 0 0 6px;
}

.td-section p:last-child {
  margin-bottom: 0;
}

.td-strong {
  font-weight: 600;
  color: #b91c1c;
}

/* 结尾强调区域 */
.td-section-highlight {
  margin-top: 18px;
  padding: 10px 12px;
  border-radius: 12px;
  background: #f3f4ff;
  border: 1px solid #e0e3ff;
}

.td-highlight-main {
  margin: 0 0 4px;
  font-weight: 600;
  color: #111827;
}

.td-highlight-sub {
  margin: 0;
  font-size: 13px;
  color: #4b5563;
}

/* 滚动条美化（可选） */
.td-body::-webkit-scrollbar {
  width: 6px;
}

.td-body::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.8);
}

.td-footer {
  margin-top: auto;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.td-btn {
  min-width: 100px;
  padding: 9px 18px;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 500;
  border: none;
  cursor: pointer;
  transition: background 0.15s ease, box-shadow 0.15s ease, transform 0.05s ease;
}

.td-btn-ghost {
  background: #e5e7eb;
  color: #374151;
}

.td-btn-ghost:hover {
  background: #d1d5db;
}

.td-btn-primary {
  background: #5a60ea;
  color: #ffffff;
}

.td-btn-primary:hover {
  background: #484dbb;
}

.td-btn:active {
  transform: translateY(1px);
}

/* 小屏适配 */
@media (max-width: 640px) {
  .td-dialog {
    padding: 18px 16px 16px;
  }

  .td-footer {
    flex-direction: column-reverse;
    align-items: stretch;
  }

  .td-btn {
    width: 100%;
  }
}
</style>
