<template>
  <TestLayout :key="route.fullPath">
    <!-- 顶部步骤条 -->
    <template #header>
      <StepIndicator/>
    </template>


    <AiGeneratingOverlay
        v-if="aiLoading"
        title="AI 正在为你生成专属报告…"
        subtitle="正在分析你的测试各项参数，为您全面展示智能分析结果"
        :log-lines="truncatedLatestMessage"
    />

  </TestLayout>
</template>

<script setup lang="ts">
import StepIndicator from "@/views/components/StepIndicator.vue";
import TestLayout from "@/layouts/TestLayout.vue";
import {useReportPage} from "@/controller/AssessmentReport";
import AiGeneratingOverlay from "@/views/components/AiGeneratingOverlay.vue";

const {
  route,
  aiLoading,
  truncatedLatestMessage,
} = useReportPage()
truncatedLatestMessage.value.push("this is a test page")
truncatedLatestMessage.value.push("this is a test page")
truncatedLatestMessage.value.push("this is a test page")
truncatedLatestMessage.value.push("this is a test page")
</script>