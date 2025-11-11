import {createRouter, createWebHistory, RouteRecordRaw} from 'vue-router'
import {getAssessmentFlowState} from '@/store/assessmentFlow'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
    },
    {
        path: '/start',
        name: 'assessment-start',
        component: () => import('@/views/AssessmentStart.vue'),
    },
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
    },
    {
        path: '/basic-info/:assessmentId',
        name: 'basic-info',
        component: () => import('@/views/BasicInfo.vue'),
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
