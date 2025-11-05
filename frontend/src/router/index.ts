import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import HobbiesView from '../views/HobbiesView.vue'
import QuestionsView from '../views/QuestionsView.vue'
import SummaryView from '../views/SummaryView.vue'
import ReportView from '../views/ReportView.vue'

const routes: RouteRecordRaw[] = [
    { path: '/', component: HomeView },
    { path: '/login', component: LoginView },
    { path: '/hobbies', component: HobbiesView },
    { path: '/questions', component: QuestionsView },
    { path: '/summary', component: SummaryView },
    { path: '/report', component: ReportView },
]

export const router = createRouter({
    history: createWebHistory(),
    routes,
})
