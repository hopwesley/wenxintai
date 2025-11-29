import {createRouter, createWebHistory, RouteRecordRaw} from 'vue-router'
import {StageBasic, StageReport, TestTypeAdv, TestTypeBasic, TestTypePro, TestTypeSchool} from "@/controller/common";

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
    },
    {
        path: '/assessment/:typ/'+StageBasic,
        name: 'test-basic-info',
        component: () => import('@/views/AssessmentBasicInfo.vue'),
        meta: { stage: StageBasic },
    },
    {
        path: '/assessment/'+TestTypeBasic+'/'+StageReport,
        name: 'test-report-basic',
        component: () => import('@/views/report_basic.vue'),
        meta: { stage: StageReport },
    },
    {
        path: '/assessment/'+TestTypePro+'/'+StageReport,
        name: 'test-report-pro',
        component: () => import('@/views/AssessmentReport.vue'),
        meta: { stage: StageReport },
    },
    {
        path: '/assessment/'+TestTypeAdv+'/'+StageReport,
        name: 'test-report-adv',
        component: () => import('@/views/AssessmentReport.vue'),
        meta: { stage: StageReport },
    },
    {
        path: '/assessment/'+TestTypeSchool+'/'+StageReport,
        name: 'test-report-school',
        component: () => import('@/views/AssessmentReport.vue'),
        meta: { stage: StageReport },
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
