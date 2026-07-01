import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '../stores/auth'
import Layout from '../Layout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/Login.vue'),
      meta: { guest: true },
    },
    {
      path: '/',
      component: Layout,
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('../views/Dashboard.vue'),
        },
        {
          path: 'domains',
          name: 'domains',
          component: () => import('../views/Domains.vue'),
        },
        {
          path: 'certificates',
          name: 'certificates',
          component: () => import('../views/Certificates.vue'),
        },
        {
          path: 'deploy',
          name: 'deploy',
          component: () => import('../views/Deploy.vue'),
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('../views/Users.vue'),
        },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const { isLoggedIn } = useAuth()
  if (to.meta.guest) {
    if (isLoggedIn()) return next('/')
    return next()
  }
  if (!isLoggedIn()) return next('/login')
  next()
})

export default router
