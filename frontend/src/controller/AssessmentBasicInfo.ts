import {computed, onMounted, reactive, ref} from 'vue'
import {useRouter} from 'vue-router'
import {TestRecordDTO, useTestSession} from '@/controller/testSession'
import {apiRequest} from '@/api'
import {useAlert} from '@/controller/useAlert'
import {
    ModeOption,
    Mode33,
    Mode312,
    CommonResponse,
    pushStageRoute, DEFAULT_HOBBIES,
} from "@/controller/common";
import {useGlobalLoading} from "@/controller/useGlobalLoading";

interface TestConfigForm {
    grade: string
    mode: ModeOption|''
    hobby: string
}

export function useStartTestConfig() {
    const {showAlert} = useAlert()
    const router = useRouter()
    const {state, setNextRouteItem} = useTestSession()
    const {showLoading, hideLoading} = useGlobalLoading()
    const record = computed<TestRecordDTO | undefined>(() => state.record)
    const form = reactive<TestConfigForm>({
        grade: record.value?.grade ?? '',
        mode: record.value?.mode ?? '',
        hobby: record.value?.hobby ?? '',
    })
    const hobbies = ref<string[]>([])
    const errorMessage = ref('')
    const publicId = computed(() => record.value?.public_id ?? '')
    const selectedMode = computed<ModeOption | null>(() => {
        return form.mode === Mode33 || form.mode === Mode312 ? form.mode : null
    })
    const canSubmit = computed(() => {
        return Boolean(form.grade.trim() && selectedMode.value)
    })

    function handleFlowError(msg?: string) {
        showAlert(msg ?? '测试流程异常，请返回首页重新开始', () => {
            router.replace({name: 'home'}).then()
        })
    }

    onMounted(async () => {
        if (!publicId) {
            handleFlowError("没有找到测试记录，请登录重试")
            return
        }
        hobbies.value = Array.isArray(DEFAULT_HOBBIES) ? DEFAULT_HOBBIES.map(String) : []
    })

    async function handleSubmit() {
        if (!record.value) {
            showAlert('测试会话异常：记录不存在')
            return
        }

        if (!selectedMode.value) {
            errorMessage.value = '请选择测试模式'
            return
        }
        if (!form.grade.trim()) {
            errorMessage.value = '请选择年级'
            return
        }

        record.value.grade = form.grade
        record.value.mode = form.mode
        record.value.hobby = form.hobby

        errorMessage.value = ''
        showLoading("正式开始测试")
        try {
            const res = await apiRequest<CommonResponse>('/api/tests/basic_info', {
                method: 'POST',
                body: {
                    public_id: record.value.public_id,
                    grade: form.grade.trim(),
                    mode: form.mode,
                    hobby: form.hobby,
                },
            })

            if (!res.ok) {
                showAlert('更新用户信息失败:' + res.msg)
                return
            }

            if (!res.next_route) {
                handleFlowError('测试流程异常，未找到下一步，请返回首页重新开始')
                return
            }

            setNextRouteItem(res.next_route, res.next_route_index)
            await pushStageRoute(router, record.value.business_type, res.next_route)
        } catch (err) {
            console.error('[StartTestConfig] handleSubmit error', err)
            handleFlowError(
                (err as Error)?.message || '测试流程异常，请返回首页重新开始',
            )
        } finally {
            hideLoading()
        }
    }

    return {
        hobbies,
        form,
        errorMessage,
        canSubmit,
        handleSubmit,
    }
}
