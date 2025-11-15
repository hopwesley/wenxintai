import {createRouter, createWebHistory, RouteRecordRaw} from 'vue-router'
import {getAssessmentFlowState} from '@/store/assessmentFlow'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
    },
    {
        path: '/assessment/:typ/basic-info',
        name: 'test-basic-info',
        component: () => import('@/views/StartTestConfigView.vue'),
    },
    {
        path: '/assessment/:typ/:scale',
        name: 'test-scale',
        component: () => import('@/views/QuestionsStageView.vue'),
    },
    {
        path: '/assessment/:typ/report',
        name: 'test-scale',
        component: () => import('@/views/ReportView.vue'),
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
