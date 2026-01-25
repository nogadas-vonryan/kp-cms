import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { defineComponent, h } from 'vue';
import { useAuthStore } from '@/modules/auth/store';

const makePlaceholder = (name: string) =>
  defineComponent({ name, setup: () => () => h('div', `${name} placeholder`) });

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: makePlaceholder('LoginPage'),
    meta: { public: true },
  },
  {
    path: '/',
    redirect: '/documents',
  },
  {
    path: '/documents',
    name: 'documents',
    component: makePlaceholder('DocumentsPage'),
    meta: { requiresAuth: true },
  },
  {
    path: '/admin',
    name: 'admin',
    component: makePlaceholder('AdminDashboard'),
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: makePlaceholder('NotFound'),
    meta: { public: true },
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore();
  const isAuthed = auth.isAuthenticated;
  const requiresAuth = Boolean(to.meta.requiresAuth);
  const requiresAdmin = Boolean(to.meta.requiresAdmin);
  const isPublic = Boolean(to.meta.public);

  if (requiresAuth && !isAuthed) {
    return next({ name: 'login', query: { redirect: to.fullPath } });
  }

  if (requiresAdmin && auth.role !== 'RoleAdmin') {
    return next({ name: 'documents' });
  }

  if (isPublic && isAuthed && to.name === 'login') {
    return next({ name: 'documents' });
  }

  return next();
});

export default router;
