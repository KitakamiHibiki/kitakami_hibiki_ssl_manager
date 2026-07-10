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
          path: 'domains/add',
          name: 'add-domain',
          component: () => import('../views/AddDomain.vue'),
        },
        {
          path: 'domains/detail',
          name: 'domain-detail',
          component: () => import('../views/DomainDetail.vue'),
        },
        {
          path: 'domains/detail/cert-apply',
          name: 'cert-apply',
          component: () => import('../views/CertApply.vue'),
        },
        {
          path: 'domains/detail/cert-download',
          name: 'cert-download',
          component: () => import('../views/CertDownload.vue'),
        },
        {
          path: 'deploy',
          name: 'deploy',
          component: () => import('../views/Deploy.vue'),
        },
        {
          path: 'certs',
          name: 'certs',
          component: () => import('../views/Certs.vue'),
        },
        {
          path: 'certs/detail',
          name: 'cert-detail',
          component: () => import('../views/CertDetail.vue'),
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
