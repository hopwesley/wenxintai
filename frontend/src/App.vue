<template>
  <router-view />
  <GlobalAlert />
</template>

<script setup lang="ts">
import GlobalAlert from '@/components/GlobalAlert.vue'
</script>

<style>
body {
  margin: 0;
  font-family: sans-serif;
}
</style>