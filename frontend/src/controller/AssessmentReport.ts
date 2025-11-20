import {useRoute, useRouter} from "vue-router";
import {useTestSession} from "@/store/testSession";
import {useAlert} from "@/controller/useAlert";
import {computed, ref} from "vue";
import {useGlobalLoading} from "@/controller/useGlobalLoading";


export function useReportPage(){
    const route = useRoute()
    const router = useRouter()
    const {state, getPublicID, setNextRouteItem} = useTestSession()
    const {showAlert} = useAlert()
    const aiLoading = ref(true)
    const {showLoading, hideLoading} = useGlobalLoading()
    const errorMessage = ref('')
    const logLines = ref<string[]>([])
    const truncatedLatestMessage = computed(() => logLines.value)

    return {route,
        aiLoading,
        errorMessage,
        truncatedLatestMessage,
    }
}