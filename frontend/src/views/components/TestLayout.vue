<template>
  <div class="test-layout">
    <div class="test-layout__inner">
      <header class="test-layout__header">
        <slot name="header" />
      </header>
      <main class="test-layout__content">
        <slot />
      </main>
      <footer class="test-layout__footer">
        <slot name="footer" />
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts"></script>

<style scoped>
.test-layout {
  height: 278px;
  background: linear-gradient(90deg, #D0DAF8 0%, #5A60EA 100%);
  padding: 24px 16px 32px;
  box-sizing: border-box;
  display: flex;
  justify-content: center;
  align-items: stretch;
}

.test-layout__inner {
  /* 关键：宽度控制在 640–960 之间，理想值为 46vw */
  width: clamp(640px, 46vw, 960px);
  display: flex;
  flex-direction: column;
  gap: 24px;
  margin-top: 32px;
}

.test-layout__header {
  min-height: 48px;
}

.test-layout__content {
  border-radius: 32px;
  box-sizing: border-box;
  flex: 1;
}

.test-layout__footer {
  font-size: 12px;
  color: rgba(30, 41, 59, 0.7);
  text-align: center;
}

</style>
