<template>
  <div v-if="visible" class="letter-overlay">
    <div class="letter-backdrop" @click="handleClose"></div>

    <section class="letter-dialog" role="dialog" aria-modal="true">
      <!-- 顶部关闭按钮 -->
      <button
          type="button"
          class="letter-close"
          @click="handleClose"
          aria-label="关闭"
      >
        ×
      </button>

      <!-- 信封抬头 -->
      <header class="letter-header">
        <div class="letter-badge">给你的一封信</div>
        <h2 class="letter-title">致未来的探索者与发光体</h2>
        <p class="letter-greeting">同学，你好！</p>
      </header>

      <!-- 正文 -->
      <div class="letter-body">
        <p>
          这份为你量身定制的选科建议，是你探索自我道路上的第一张“战略地图”。但请记住，这绝非终点，而是你扬帆起航的起点。无论你的指针偏向深邃的文科、精密的理科，还是充满无限可能的融合方向，你都拥有一个共同且最宝贵的资本——
          <strong>无限的未来潜力</strong>。
        </p>

        <p>
          我们生活在一个边界日益模糊的时代。顶尖的科学家需要哲学的思辨来指引方向，卓越的人文学者借助数据工具洞察社会规律，而伟大的创新往往诞生于不同领域交错的十字路口。因此，无论你选择哪一条学科赛道，
          <strong>深刻的理解力、丰富的创造力、独立的判断力与温暖的协作力</strong>，才是穿越未来一切不确定性的核心装备。
        </p>

        <p>
          我们鼓励你，在钻研学业的同时，请务必留一扇看向世界的窗。
          <strong>主动接触真实的社会</strong>：参与一次志愿活动，尝试一份暑期实习，与不同领域的师长深入交谈。这些经历将在书本之外，为你补上至关重要的一课，让你理解社会的运作，看见自己的责任，并找到为之奋斗的意义。
        </p>

        <p>
          请像耕耘一片沃土般对待你的成长时光。在你的“天赋区”深度挖掘，建立知识壁垒；同时，也要勇敢地在未知领域播撒种子，博观约取。
          <strong>你的未来，绝不会被“文科生”或“理科生”的标签所定义</strong>。它将由你的格局、你选择的挑战，以及你为这个世界创造的价值来共同书写。
        </p>

        <p>
          去吧，年轻的探索者与发光体。凭借逻辑与理性去构建你的世界，带着人文与温情去照亮他人的道路。我们坚信，你已手握罗盘，心灯长明，前方是海阔天空、万丈光芒的未来。
        </p>
      </div>

      <!-- 按钮区 -->
      <footer class="letter-footer">
        <button
            type="button"
            class="letter-btn letter-btn--ghost"
            @click="handleClose"
        >
          关闭
        </button>
        <button
            type="button"
            class="letter-btn letter-btn--primary"
            @click="handleConfirm"
        >
          确认报告已阅
        </button>
      </footer>
    </section>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  /** 是否显示这封信 */
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'confirm'): void
  (e: 'close'): void
}>()

const handleConfirm = () => {
  emit('confirm')
}

const handleClose = () => {
  emit('close')
}
</script>

<style scoped>
.letter-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.letter-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(3px);
}

.letter-dialog {
  position: relative;
  max-width: 720px;
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

/* 顶部关闭按钮 */
.letter-close {
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

.letter-close:hover {
  color: #555;
}

/* 头部 */
.letter-header {
  margin-bottom: 16px;
}

.letter-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #7b5cff;
  background: rgba(123, 92, 255, 0.08);
  margin-bottom: 8px;
}

.letter-title {
  margin: 0 0 6px;
  font-size: 20px;
  font-weight: 700;
  color: #333;
}

.letter-greeting {
  margin: 0;
  font-size: 14px;
  color: #666;
}

/* 正文 */
.letter-body {
  flex: 1;
  overflow-y: auto;
  padding-right: 6px;
  margin: 8px 0 18px;
  font-size: 14px;
  line-height: 1.8;
  color: #444;
}

.letter-body p {
  margin: 0 0 10px;
}

.letter-body strong {
  font-weight: 600;
  color: #5a3cff;
}

/* 按钮区域 */
.letter-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}

.letter-btn {
  min-width: 120px;
  padding: 8px 14px;
  border-radius: 999px;
  font-size: 14px;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
}

.letter-btn--ghost {
  background: transparent;
  color: #666;
  border: 1px solid #ddd;
}

.letter-btn--ghost:hover {
  background: #f2f2f2;
}

.letter-btn--primary {
  background: #7b5cff;
  color: #fff;
  font-weight: 600;
}

.letter-btn--primary:hover {
  background: #6543ff;
}

/* 小屏适配 */
@media (max-width: 600px) {
  .letter-dialog {
    padding: 20px 18px 16px;
    border-radius: 12px;
  }

  .letter-title {
    font-size: 18px;
  }

  .letter-body {
    font-size: 13px;
  }

  .letter-btn {
    min-width: 0;
    padding-inline: 12px;
  }
}
</style>
