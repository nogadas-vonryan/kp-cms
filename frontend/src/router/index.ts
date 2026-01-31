import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useAuthStore } from '@/modules/auth/store';
import { setApiRouter } from '@/core/api/client';
import LoginPage from '@/pages/LoginPage.vue';
import DocumentsPage from '@/modules/documents/pages/DocumentsPage.vue';
import DocumentDetailPage from '@/modules/documents/pages/DocumentDetailPage.vue';
import AdminDashboard from '@/pages/AdminDashboard.vue';
import NotFound from '@/pages/NotFound.vue';
import ReportExportPage from '@/modules/reports/pages/ReportExportPage.vue';
import ImportPage from '@/modules/import/pages/ImportPage.vue';
import BackupPage from '@/modules/backup/pages/BackupPage.vue';

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: LoginPage,
    meta: { public: true },
  },
  {
    path: '/',
    redirect: '/documents',
  },
  {
    path: '/documents',
    name: 'documents',
    component: DocumentsPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/documents/:documentId',
    name: 'document-detail',
    component: DocumentDetailPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/reports/export',
    name: 'report-export',
    component: ReportExportPage,
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  {
    path: '/import',
    name: 'import',
    component: ImportPage,
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  {
    path: '/backup',
    name: 'backup',
    component: BackupPage,
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  {
    path: '/admin',
    name: 'admin',
    component: AdminDashboard,
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: NotFound,
    meta: { public: true },
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

// Inject router into API client for auth error handling
setApiRouter(router);

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
