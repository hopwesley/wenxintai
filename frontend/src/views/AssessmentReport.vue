<template>
  <TestLayout :key="route.fullPath">
    <template #header>
      <StepIndicator/>
    </template>
    <component :is="currentMainComponent"/>
    <div class="report-page__actions">
      <button
          class="btn btn-secondary report-page__action"
          @click="showFinishLetter = true">
        完成报告
      </button>

      <button class="btn btn-primary report-page__action" @click="handleExportPdf">
        导出 PDF
      </button>
    </div>
    <AiGeneratingOverlay
        v-if="aiLoading"
        title="AI 正在为你生成专属报告…"
        subtitle="正在分析你的测试各项参数，为您全面展示智能分析结果"
        :log-lines="truncatedLatestMessage"
        stage="AI报告"
    />
    <ReportFinishLetter
        :visible="showFinishLetter"
        @close="showFinishLetter = false"
        @confirm="handleLetterConfirm"
    />
    <PaymentModal
        v-model:open="paymentDialogShow"
        :product="currentPlan"
        :publicId="publicId"
        @success="generateReport"
    />
  </TestLayout>
</template>

<script setup lang="ts">

import StepIndicator from '@/views/components/StepIndicator.vue'
import TestLayout from '@/views/components/TestLayout.vue'
import AiGeneratingOverlay from '@/views/components/AiGeneratingOverlay.vue'
import {useReportController} from '@/controller/report_manager'
import {computed} from "vue";

import ReportBasic from '@/views/report_basic.vue'
import ReportPro from '@/views/report_pro.vue'
import {TestTypeAdv, TestTypeBasic, TestTypePro, TestTypeSchool} from "@/controller/common";
import ReportFinishLetter from "@/views/components/ReportFinishLetter.vue";
import PaymentModal from "@/views/components/PaymentModal.vue";

const {
  route,
  aiLoading,
  truncatedLatestMessage,
  handleExportPdf,
  showFinishLetter,
  handleLetterConfirm,
  paymentDialogShow,
  currentPlan,
  publicId,
  generateReport,
} = useReportController()

const businessType = computed(() =>
    String(route.params.typ ?? "")
)

const currentMainComponent = computed(() => {
  switch (businessType.value) {
    case TestTypePro:
      return ReportPro
    case TestTypeAdv:
      return ReportBasic
    case TestTypeSchool:
      return ReportBasic
    case TestTypeBasic:
    default:
      return ReportBasic
  }
})
</script>

<style scoped src="@/styles/assessment-report.css"></style>
<style scoped src="@/styles/pdf.css"></style>
