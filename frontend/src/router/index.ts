import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { isVariant } from '@/config/testSteps'
import { useTestSession } from '@/store/testSession'

const routes: RouteRecordRaw[] = [
    { path: '/', component: () => import('@/views/HomeView.vue') },
    { path: '/login', component: () => import('@/views/LoginView.vue') },
    { path: '/hobbies', component: () => import('@/views/HobbiesView.vue') },
    { path: '/summary', component: () => import('@/views/SummaryView.vue') },
    { path: '/report', component: () => import('@/views/ReportView.vue') },
    {
      path: '/questions',
      redirect: '/test/basic/step/2',
    },
    {
      path: '/test/:variant/step/1',
      name: 'test-basic-info',
      component: () => import('@/views/TestBasicInfoView.vue'),
      beforeEnter: (to) => {
        const variant = String(to.params.variant)
        if (!isVariant(variant)) {
          return '/test/basic/step/1'
        }
      },
    },
    {
      path: '/test/:variant/step/2',
      name: 'test-stage-1',
      component: () => import('@/views/QuestionsStageView.vue'),
      props: { stage: 1 },
      beforeEnter: (to) => {
        const variant = String(to.params.variant)
        if (!isVariant(variant)) {
          return '/test/basic/step/1'
        }
      },
    },
    {
      path: '/test/:variant/step/3',
      name: 'test-stage-2',
      component: () => import('@/views/QuestionsStageView.vue'),
      props: { stage: 2 },
      beforeEnter: (to) => {
        const variant = String(to.params.variant)
        if (!isVariant(variant)) {
          return '/test/basic/step/1'
        }
      },
    },
    {
      path: '/test/:variant/step/4',
      name: 'test-report',
      component: () => import('@/views/ReportView.vue'),
      beforeEnter: (to) => {
        const variant = String(to.params.variant)
        if (!isVariant(variant)) {
          return '/test/basic/step/1'
        }
      },
    },
    {
      path: '/test/:variant/step/5',
      name: 'test-stage-placeholder-1',
      component: () => import('@/views/TestPlaceholderView.vue'),
      beforeEnter: (to) => {
        const variant = String(to.params.variant)
        if (!isVariant(variant)) {
          return '/test/basic/step/1'
        }
      },
    },
    {
      path: '/test/:variant/step/6',
      name: 'test-stage-placeholder-2',
      component: () => import('@/views/TestPlaceholderView.vue'),
      beforeEnter: (to) => {
        const variant = String(to.params.variant)
        if (!isVariant(variant)) {
          return '/test/basic/step/1'
        }
      },
    },
]

export const router = createRouter({
    history: createWebHistory(),
    routes,
})

const protectedPatterns = [/^\/test\//, /^\/questions$/, /^\/summary$/, /^\/report$/]

router.beforeEach((to, from, next) => {
    const requiresGuard = protectedPatterns.some(pattern => pattern.test(to.path))
    if (!requiresGuard) {
        next()
        return
    }

    const { getSessionId } = useTestSession()
    const sessionId = getSessionId()
    if (sessionId) {
        next()
        return
    }

    if (typeof window !== 'undefined') {
        window.alert('需要邀请码或登录后访问')
    }
    next('/')
})
