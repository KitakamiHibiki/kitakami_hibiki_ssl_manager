import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../views/Dashboard.vue'),
    },
    {
      path: '/domains',
      name: 'domains',
      component: () => import('../views/Domains.vue'),
    },
    {
      path: '/certificates',
      name: 'certificates',
      component: () => import('../views/Certificates.vue'),
    },
    {
      path: '/deploy',
      name: 'deploy',
      component: () => import('../views/Deploy.vue'),
    },
  ],
})

export default router
