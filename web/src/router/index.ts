import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '../layouts/AdminLayout.vue'
import { getToken, redirectToPortal, ensureSession, clearToken } from '../utils/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/auth/callback',
      name: 'AuthCallback',
      component: () => import('../views/AuthCallback.vue'),
      meta: { public: true },
    },
    {
      path: '/auth/logout',
      name: 'AuthLogout',
      component: () => import('../views/AuthLogout.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: AdminLayout,
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', name: 'Dashboard', component: () => import('../views/Dashboard.vue'), meta: { title: '工作台' } },
        { path: 'suppliers', name: 'SupplierList', component: () => import('../views/supplier/SupplierList.vue'), meta: { title: '供应商' } },
        { path: 'suppliers/:id', name: 'SupplierDetail', component: () => import('../views/supplier/SupplierDetail.vue'), meta: { title: '供应商详情' } },
        { path: 'sku-offers', name: 'OfferList', component: () => import('../views/offer/OfferList.vue'), meta: { title: 'SKU 供货报价' } },
        { path: 'purchase-orders', name: 'PurchaseOrderList', component: () => import('../views/purchase/PurchaseOrderList.vue'), meta: { title: '采购单' } },
        { path: 'purchase-orders/create', name: 'PurchaseOrderCreate', component: () => import('../views/purchase/PurchaseOrderForm.vue'), meta: { title: '新建采购单' } },
        { path: 'purchase-orders/:id/edit', name: 'PurchaseOrderEdit', component: () => import('../views/purchase/PurchaseOrderForm.vue'), meta: { title: '编辑采购单' } },
        { path: 'purchase-orders/:id', name: 'PurchaseOrderDetail', component: () => import('../views/purchase/PurchaseOrderDetail.vue'), meta: { title: '采购单详情' } },
        { path: 'sales-orders', name: 'SalesOrderList', component: () => import('../views/sales/SalesOrderList.vue'), meta: { title: '销售订单' } },
        { path: 'sales-orders/create', name: 'SalesOrderCreate', component: () => import('../views/sales/SalesOrderForm.vue'), meta: { title: '新建销售单' } },
        { path: 'sales-orders/:id/edit', name: 'SalesOrderEdit', component: () => import('../views/sales/SalesOrderForm.vue'), meta: { title: '编辑销售单' } },
        { path: 'sales-orders/:id', name: 'SalesOrderDetail', component: () => import('../views/sales/SalesOrderDetail.vue'), meta: { title: '销售单详情' } },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.public) return true
  if (!getToken()) {
    redirectToPortal()
    return false
  }
  const ok = await ensureSession()
  if (!ok) {
    clearToken()
    redirectToPortal()
    return false
  }
  return true
})

export default router
