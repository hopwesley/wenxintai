import {createRouter, createWebHistory, RouteRecordRaw} from 'vue-router'
import {StageBasic, StageReport} from "@/controller/common";

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
    },
    {
        path: '/my-tests',
        name: 'my-tests',
        component: () => import('@/views/WechatProfile.vue'),
    },
    {
        path: '/assessment/:typ/'+StageBasic,
        name: 'test-basic-info',
        component: () => import('@/views/AssessmentBasicInfo.vue'),
        meta: { stage: StageBasic },
    },
    {
        path: '/assessment/:typ/'+StageReport,
        name: 'test-report',
        component: () => import('@/views/AssessmentReport.vue'),
        meta: {  stage: StageReport },
    },
    {
        path: '/assessment/:businessType/:testStage',
        name: 'test-stage',
        component: () => import('@/views/AssessmentQuestions.vue'),
    },
    {
        path: '/agreements/user',
        name: 'agreement-user',
        component: () => import('@/views/agreements/UserAgreementView.vue'),
    },
    {
        path: '/agreements/privacy',
        name: 'agreement-privacy',
        component: () => import('@/views/agreements/PrivacyPolicyView.vue'),
    },
    {
        path: '/agreements/license',
        name: 'agreement-license',
        component: () => import('@/views/agreements/LicenseAgreementView.vue'),
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
