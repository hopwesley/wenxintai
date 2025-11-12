import {createRouter, createWebHistory, RouteRecordRaw} from 'vue-router'
import {getAssessmentFlowState} from '@/store/assessmentFlow'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
    },
    {
        path: '/test/basic/step/1',
        name: 'test-config',
        component: () => import('@/views/StartTestConfigView.vue'),
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
        path: '/test/:variant/step/:step(\\d+)',
        name: 'test-stage',
        component: () => import('@/views/QuestionsStageView.vue'),
    },
    {
        path: '/:pathMatch(.*)*',
        redirect: '/'
    },
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
        next({name: 'home'})
        return
    }
    if (to.name === 'assessment-report') {
        if (!to.params.assessmentId && state.assessmentId) {
            next({name: 'assessment-report', params: {assessmentId: state.assessmentId}})
            return
        }
        if (!to.params.assessmentId) {
            next({name: 'home'})
            return
        }
    }
    next()
})
