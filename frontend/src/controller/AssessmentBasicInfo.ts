import {computed, onMounted, reactive, ref} from 'vue'
import {useRouter} from 'vue-router'
import {useTestSession} from '@/controller/testSession'
import {apiRequest, getHobbies} from '@/api'
import {useAlert} from '@/controller/useAlert'
import {
    ModeOption,
    Mode33,
    Mode312,
    TestTypeBasic,
    CommonResponse,
    pushStageRoute,
} from "@/controller/common";

interface TestConfigForm {
    grade: string
    mode: ModeOption | ''
    hobby: string
}

export function useStartTestConfig() {
    const {showAlert} = useAlert()
    const router = useRouter()
    const {state, setTestConfig, getPublicID, setNextRouteItem} = useTestSession()

    function handleFlowError(msg?: string) {
        showAlert(msg ?? '测试流程异常，请返回首页重新开始', () => {
            router.replace('/').then()
        })
    }


    const form = reactive<TestConfigForm>({
        grade: state.grade ?? '',
        mode: state.mode ?? '',
        hobby: state.hobby ?? '',
    })

    const hobbies = ref<string[]>([])
    const errorMessage = ref('')
    const submitting = ref(false)
    const inviteCode = computed(() => state.inviteCode ?? '')
    const public_id: string = getPublicID() as string

    const selectedMode = computed<ModeOption | null>(() => {
        return form.mode === Mode33 || form.mode === Mode312 ? form.mode : null
    })

    const canSubmit = computed(() => {
        return Boolean(form.grade.trim() && selectedMode.value)
    })

    onMounted(async () => {
        if (!public_id) {
            handleFlowError("没有找到测试记录，请登录重试")
            return
        }

        try {
            const list = await getHobbies()
            hobbies.value = Array.isArray(list) ? list.map(String) : []
        } catch (error) {
            console.warn('[StartTestConfig] failed to load hobbies', error)
            hobbies.value = []
        }
    })


    interface BasicInfoReq {
        public_id: string
        grade: string
        mode: string
        hobby?: string | null
    }

    function updateTestBasicInfo(payload: BasicInfoReq) {
        return apiRequest<CommonResponse>('/api/tests/basic_info', {
            method: 'POST',
            body: payload,
        })
    }

    async function handleSubmit() {

        if (!selectedMode.value) {
            errorMessage.value = '请选择测试模式'
            return
        }
        if (!form.grade.trim()) {
            errorMessage.value = '请选择年级'
            return
        }

        errorMessage.value = ''
        submitting.value = true

        try {
            const grade = form.grade.trim()
            const hobby = form.hobby.trim()

            const res = await updateTestBasicInfo({
                public_id: public_id,
                grade,
                mode: selectedMode.value as ModeOption,
                hobby: hobby || null,
            })
            if (!res.ok) {
                showAlert('更新用户信息失败:' + res.msg)
                return
            }
            setTestConfig({
                grade,
                mode: selectedMode.value as ModeOption,
                hobby: hobby || undefined,
            })

            if (!res.next_route) {
                handleFlowError('测试流程异常，未找到下一步，请返回首页重新开始')
                return
            }

            setNextRouteItem(res.next_route, res.next_route_index)
            const businessType = state.businessType || TestTypeBasic
            await pushStageRoute(router, businessType, res.next_route)
        } catch (err) {
            console.error('[StartTestConfig] handleSubmit error', err)
            handleFlowError(
                (err as Error)?.message || '测试流程异常，请返回首页重新开始',
            )
        } finally {
            submitting.value = false
        }
    }

    return {
        inviteCode,
        hobbies,
        form,
        submitting,
        errorMessage,
        canSubmit,
        handleSubmit,
    }
}
