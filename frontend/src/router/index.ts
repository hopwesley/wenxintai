import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
    { path: '/', component: () => import('@/views/HomeView.vue') },
    { path: '/login', component: () => import('@/views/LoginView.vue') },
    { path: '/hobbies', component: () => import('@/views/HobbiesView.vue') },
    { path: '/questions', component: () => import('@/views/QuestionsView.vue') },
    { path: '/summary', component: () => import('@/views/SummaryView.vue') },
    { path: '/report', component: () => import('@/views/ReportView.vue') },
]

export const router = createRouter({
    history: createWebHistory(),
    routes,
})
