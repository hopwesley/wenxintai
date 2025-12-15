<template>
  <div class="report-preview-modal">
    <div class="report-preview-backdrop"></div>
    <div class="report-preview-panel" :key="panelKey">
      <header class="report-preview-header">
        <div class="report-preview-header-close">
          <button
              type="button"
              class="btn btn-ghost close-btn-header"
              @click="emit('close')"
          >
            关闭
          </button>

        </div>
        <div class="report-preview-header-title-pdf">
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

        </div>

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
  max-height: 90vh;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 12px 50px rgba(15, 23, 42, 0.24);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.report-preview-header {
  padding: 16px 32px;
  border-bottom: 1px solid #e5e7eb;
}
.report-preview-header-close{
  width: 100%;
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.report-preview-header-title-pdf {
  display: flex;
  align-items: center;
  justify-content: space-between;
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
  padding: 0 32px;
  overflow-y: auto;
}
</style>

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
.btn-ghost {
  background: #5a60ea;
  color: #fff;
  border: none;
}

.btn-ghost:hover {
  background: var(--brand-dark)
}

.close-btn-header {
  background: transparent;
  font-size: 14px;
  cursor: pointer;
  color: #888;
}
.close-btn-header:hover {
  color: #000;
  background-color: transparent;
}

</style>
