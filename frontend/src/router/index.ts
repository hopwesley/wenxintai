import {createRouter, createWebHistory, RouteRecordRaw} from 'vue-router'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
    },
    {
        path: '/assessment/:typ/basic-info',
        name: 'test-basic-info',
        component: () => import('@/views/AssessmentBasicInfo.vue'),
    },
    {
        path: '/assessment/:typ/:scale',
        name: 'test-scale',
        component: () => import('@/views/AssessmentQuestions.vue'),
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
