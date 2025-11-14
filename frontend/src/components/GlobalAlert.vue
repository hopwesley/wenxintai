<template>
  <AlertModal
      :model-value="visible"
      :title="title"
      :message="message"
      @update:modelValue="onUpdateModelValue"
      @confirm="handleConfirm"
  />
</template>

<script setup lang="ts">
import AlertModal from './AlertModal.vue'
import { useAlert } from '@/composables/useAlert'

const { visible, title, message, handleConfirm, closeAlert } = useAlert()

function onUpdateModelValue(v: boolean) {
  if (!v) {
    closeAlert()
  }
}
</script>
