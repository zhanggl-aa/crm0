import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import AppLayout from '../components/AppLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: AppLayout,
    meta: { requiresAuth: true },
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue')
      },
      {
        path: 'customers',
        name: 'Customers',
        component: () => import('../views/Customers.vue')
      },
      {
        path: 'customers/:id',
        name: 'CustomerDetail',
        component: () => import('../views/CustomerDetail.vue')
      },
      {
        path: 'subscriptions',
        name: 'Subscriptions',
        component: () => import('../views/Subscriptions.vue')
      },
      {
        path: 'analytics/churn',
        name: 'ChurnAnalysis',
        component: () => import('../views/analytics/ChurnAnalysis.vue')
      },
      {
        path: 'analytics/segments',
        name: 'SegmentAnalysis',
        component: () => import('../views/analytics/SegmentAnalysis.vue')
      },
      {
        path: 'analytics/ltv',
        name: 'LTVAnalysis',
        component: () => import('../views/analytics/LTVAnalysis.vue')
      },
      {
        path: 'analytics/nba',
        name: 'NBARecommendations',
        component: () => import('../views/analytics/NBARecommendations.vue')
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('../views/Settings.vue')
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth !== false && !authStore.isAuthenticated) {
    next('/login')
  } else if ((to.path === '/login' || to.path === '/register') && authStore.isAuthenticated) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
