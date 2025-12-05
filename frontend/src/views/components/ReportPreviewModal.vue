<template>
  <div class="report-preview-modal">
    <div class="report-preview-backdrop"></div>
    <div class="report-preview-panel" :key="panelKey">
      <header class="report-preview-header">
        <div class="report-preview-header__titles">
          <p class="report-preview-label">测评报告预览</p>
          <h3 class="report-preview-title">{{ headerTitle }}</h3>
          <p class="report-preview-meta">报告编号：{{ publicId }}</p>
        </div>
        <button
            type="button"
            class="btn btn-ghost"
            @click="handleModalExportPdf"
        >
          打印 pdf
        </button>
        <button
            type="button"
            class="btn btn-ghost"
            @click="emit('close')"
        >
          关闭
        </button>
      </header>
      <div class="report-preview-body">
        <div class="report-page report-page--pdf">
          <component :is="currentMainComponent"/>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {computed, watch} from 'vue'
import ReportBasic from '@/views/report_basic.vue'
import ReportPro from '@/views/report_pro.vue'
import {
  TestTypeAdv,
  TestTypeBasic,
  TestTypePro,
  TestTypeSchool,
} from '@/controller/common'
import {useReportController} from '@/controller/report_manager'

interface ReportPreviewModalProps {
  businessType: string
  publicId: string
}

const props = defineProps<ReportPreviewModalProps>()
const emit = defineEmits(['close'])

const panelKey = computed(() => `${props.businessType}-${props.publicId}`)

const {
  handleExportPdf,   // 里面就是你之前的 window.print + 改 title
  generateReport,
} = useReportController({
  publicId: computed(() => props.publicId),
  businessType: computed(() => props.businessType),
  autoQueryOnMounted: false,     // 不自动请求
})

watch(
    () => props.publicId,
    async (newId, oldId) => {
      if (!newId) return
      if (newId === oldId) return
      await generateReport()
    },
    {
      immediate: true,
    },
)

/**
 * 弹框里的打印：还是用同一个 handleExportPdf（window.print），
 * 但是通过本组件自己的 @media print 把布局拍平，只保留 report-page。
 */
const handleModalExportPdf = () => {
  handleExportPdf()
}

const headerTitle = computed(() => {
  switch (props.businessType) {
    case TestTypePro:
      return '进阶能力测评报告'
    case TestTypeAdv:
      return '深度选科规划报告'
    case TestTypeSchool:
      return '校园合作测评报告'
    case TestTypeBasic:
    default:
      return '基础能力测评报告'
  }
})

const currentMainComponent = computed(() => {
  switch (props.businessType) {
    case TestTypePro:
      return ReportPro
    case TestTypeAdv:
    case TestTypeSchool:
      return ReportBasic
    case TestTypeBasic:
    default:
      return ReportBasic
  }
})
</script>

<style scoped>
.report-preview-modal {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1200;
}

.report-preview-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
}

.report-preview-panel {
  position: relative;
  width: min(1100px, 92vw);
  max-height: 90vh;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 12px 50px rgba(15, 23, 42, 0.24);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.report-preview-header {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.report-preview-header__titles {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.report-preview-label {
  margin: 0;
  font-size: 12px;
  color: var(--text-third);
  letter-spacing: 0.2px;
}

.report-preview-title {
  margin: 0;
  font-size: 16px;
  color: var(--text-primary);
}

.report-preview-meta {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.report-preview-body {
  padding: 0 12px 16px;
  overflow-y: auto;
}

.report-preview-body :deep(.test-layout) {
  padding-top: 12px;
}

.report-preview-body :deep(.report-page__actions) {
  display: none;
}
</style>

<!-- 这里是不加 scoped 的全局打印规则，只在这个组件里定义，但不改 pdf.css 文件 -->
<style>
@media print {
  /* 1. 默认先都隐藏 */
  body * {
    visibility: hidden !important;
  }

  /* 2. 让预览弹框这条链 + 报告内容重新可见 */
  .report-preview-modal,
  .report-preview-modal *,
  .report-preview-panel,
  .report-preview-panel *,
  .report-preview-body,
  .report-preview-body *,
  .report-page,
  .report-page * {
    visibility: visible !important;
  }

  /* 3. 把弹框从 fixed 还原成普通文档流，避免 max-height / overflow 影响打印 */
  .report-preview-modal {
    position: static !important;
    inset: auto !important;
    display: block !important;
    align-items: flex-start !important;
    justify-content: flex-start !important;
    z-index: auto !important;
    background: #ffffff !important;
  }

  .report-preview-panel {
    position: static !important;
    width: 100% !important;
    max-height: none !important;
    box-shadow: none !important;
    border-radius: 0 !important;
    overflow: visible !important;
    margin: 0 !important;
  }

  .report-preview-body {
    overflow: visible !important;
    padding: 0 !important;
  }

  /* 4. 不需要的 UI 隐藏掉 */
  .report-preview-backdrop,
  .report-preview-header {
    display: none !important;
    visibility: hidden !important;
  }
}
</style>
