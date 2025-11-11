import {createRouter, createWebHistory, RouteRecordRaw} from 'vue-router'
import {getAssessmentFlowState} from '@/store/assessmentFlow'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
    },
    // 2) 将 AssessmentStart 改到 /start（保持 name 不变以兼容守卫里的回退）
    {
        path: '/start',
        name: 'assessment-start',
        component: () => import('@/views/AssessmentStart.vue'),
    },
    // 3) 补上测试流程的动态路由，承接 HomeView 验证后的跳转
    {
        path: '/test/:variant/step/:step',
        name: 'test-step',
        component: () => import('@/views/TestPlaceholderView.vue'),
    },
    {
        path: '/questions/:questionSetId?',
        name: 'assessment-questions',
        component: () => import('@/views/QuestionPage.vue')
    },
    {
        path: '/report/:assessmentId',
        name: 'assessment-report',
        component: () => import('@/views/ReportPage.vue')
    },
    {
        path: '/:pathMatch(.*)*',
        redirect: '/'
    }
]

export const router = createRouter({
    history: createWebHistory(),
    routes
})

router.beforeEach((to, _from, next) => {
    const state = getAssessmentFlowState()
    if (to.name === 'assessment-questions') {
        const questionSetId = String(to.params.questionSetId ?? '')
        if (questionSetId) {
            next()
            return
        }
        if (state.activeQuestionSetId) {
            next({name: 'assessment-questions', params: {questionSetId: state.activeQuestionSetId}})
            return
        }
        next({name: 'assessment-start'})
        return
    }
    if (to.name === 'assessment-report') {
        if (!to.params.assessmentId && state.assessmentId) {
            next({name: 'assessment-report', params: {assessmentId: state.assessmentId}})
            return
        }
        if (!to.params.assessmentId) {
            next({name: 'assessment-start'})
            return
        }
    }
    next()
})
