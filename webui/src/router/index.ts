import { createRouter, createWebHistory } from 'vue-router'

import HomeView from '@/views/HomeView.vue'
import LoginView from '@/views/LoginView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
    },
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
  ],
})

router.beforeEach((to) => {
  const accessToken = localStorage.getItem('access_token')

  if (to.path !== '/login' && !accessToken) {
    return '/login'
  }

  if (to.path === '/login' && accessToken) {
    return '/'
  }

  return true
})

export default router
