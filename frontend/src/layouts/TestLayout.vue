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
  min-height: 100vh;
  background: linear-gradient(180deg, #f1f5ff 0%, #f9f4ff 50%, #fff0f7 100%);
  padding: 24px 16px 32px;
  box-sizing: border-box;
  display: flex;
  justify-content: center;
  align-items: stretch;
}

.test-layout__inner {
  width: min(960px, 100%);
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.test-layout__header {
  min-height: 48px;
}

.test-layout__content {
  background-color: #ffffff;
  border-radius: 16px;
  box-shadow: 0 12px 36px rgba(44, 62, 80, 0.08);
  padding: 32px;
  box-sizing: border-box;
  flex: 1;
}

.test-layout__footer {
  font-size: 12px;
  color: rgba(30, 41, 59, 0.7);
  text-align: center;
}

@media (max-width: 768px) {
  .test-layout {
    padding: 16px 12px 24px;
  }

  .test-layout__content {
    padding: 24px 16px;
    border-radius: 12px;
  }
}
</style>
