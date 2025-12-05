// src/controller/useTestLauncher.ts
import {useRouter} from 'vue-router'
import {API_PATHS, apiRequest} from '@/api'
import {useAlert} from '@/controller/useAlert'
import {useGlobalLoading} from '@/controller/useGlobalLoading'
import {useTestSession, type TestRecordDTO} from '@/controller/testSession'
import {
    pushStageRoute,
    StageBasic,
    type PlanKey,
    type TestFlowStep,
} from '@/controller/common'

export interface FetchTestFlowResponse {
    record: TestRecordDTO
    steps: TestFlowStep[]
    current_stage: string
    current_index: number
}

interface LaunchTestOptions {
    businessType: PlanKey
    /** 继续测试时传已有记录的 public_id；新建测试可不传 */
    publicId?: string
    /** 可选：自定义 loading 文案 */
    loadingText?: string
}

export function useTestLauncher() {
    const router = useRouter()
    const {showAlert} = useAlert()
    const {showLoading, hideLoading} = useGlobalLoading()
    const {setTestFlow, setNextRouteItem, setRecord} = useTestSession()

    async function launchTest(opts: LaunchTestOptions) {
        const {
            businessType,
            publicId,
            loadingText = publicId ? '正在为你恢复测试进度…' : '进入测试环节',
        } = opts

        showLoading(loadingText)

        try {
            const body: any = {business_type: businessType}
            if (publicId) {
                body.public_id = publicId
            }

            const resp = await apiRequest<FetchTestFlowResponse>(API_PATHS.TEST_FLOW, {
                method: 'POST',
                body,
            })

            const steps = resp.steps || []
            if (!steps.length) {
                console.error('[useTestLauncher] flow error: empty steps', resp)
                showAlert('测试流程异常，未配置任何测试阶段')
                return
            }

            const currentStage = resp.current_stage || StageBasic
            const currentIndex = resp.current_index ?? 0

            // 写入全局 session
            setRecord(resp.record)
            setTestFlow(steps)
            setNextRouteItem(currentStage, currentIndex)

            // 跳转到对应阶段
            await pushStageRoute(router, businessType, currentStage)
        } catch (e) {
            console.error('[useTestLauncher] launchTest failed', e)
            showAlert('创建/恢复测试失败，请稍后重试')
        } finally {
            hideLoading()
        }
    }

    return {
        launchTest,
    }
}
