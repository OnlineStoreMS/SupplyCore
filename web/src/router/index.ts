import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '../layouts/AdminLayout.vue'
import {redirectToPortal, ensureSession, clearToken} from '../utils/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/m/photo-upload',
      name: 'MobilePhotoUpload',
      component: () => import('../views/MobilePhotoUpload.vue'),
      meta: { public: true },
    },
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
        { path: 'suppliers', name: 'SupplierList', component: () => import('../views/supplier/SupplierList.vue'), meta: { title: '供应商信息' } },
        { path: 'suppliers/:id', name: 'SupplierDetail', component: () => import('../views/supplier/SupplierDetail.vue'), meta: { title: '供应商详情' } },
        { path: 'sku-offers', name: 'OfferList', component: () => import('../views/offer/OfferList.vue'), meta: { title: 'SKU 供货报价' } },
        {
          path: 'purchase-orders/dropship',
          name: 'DropshipOrderList',
          component: () => import('../views/purchase/PurchaseOrderList.vue'),
          meta: { title: '代发订单', fulfillmentType: 'dropship' },
        },
        {
          path: 'purchase-orders/stock-in',
          name: 'StockInOrderList',
          component: () => import('../views/purchase/PurchaseOrderList.vue'),
          meta: { title: '采购订单', fulfillmentType: 'stock_in' },
        },
        { path: 'purchase-orders', name: 'PurchaseOrderList', component: () => import('../views/purchase/PurchaseOrderList.vue'), meta: { title: '供应商订单' } },
        { path: 'purchase-orders/create', name: 'PurchaseOrderCreate', component: () => import('../views/purchase/PurchaseOrderForm.vue'), meta: { title: '新建供应商订单' } },
        { path: 'purchase-orders/:id(\\d+)/edit', name: 'PurchaseOrderEdit', component: () => import('../views/purchase/PurchaseOrderForm.vue'), meta: { title: '编辑供应商订单' } },
        { path: 'purchase-orders/:id(\\d+)', name: 'PurchaseOrderDetail', component: () => import('../views/purchase/PurchaseOrderDetail.vue'), meta: { title: '供应商订单详情' } },
        { path: 'purchase-accounts', name: 'PurchaseAccounts', component: () => import('../views/account/PurchaseAccountList.vue'), meta: { title: '采购账号' } },
        { path: 'purchase-inbounds', name: 'InboundList', component: () => import('../views/inbound/InboundList.vue'), meta: { title: '采购入库单' } },
        { path: 'purchase-inbounds/create', name: 'InboundCreate', component: () => import('../views/inbound/InboundForm.vue'), meta: { title: '新增采购入库单' } },
        { path: 'purchase-inbounds/:id', name: 'InboundDetail', component: () => import('../views/inbound/InboundDetail.vue'), meta: { title: '入库单详情' } },
        { path: 'inbound-sort', name: 'InboundSort', component: () => import('../views/inbound/InboundSort.vue'), meta: { title: '采购入库分拣' } },
        { path: 'package-receives', name: 'PackageReceives', component: () => import('../views/package/PackageReceiveList.vue'), meta: { title: '收货记录' } },
        { path: 'scan-inbound', name: 'ScanInbound', component: () => import('../views/package/ScanInbound.vue'), meta: { title: '包裹扫描入库' } },
        { path: 'purchase-returns', name: 'ReturnList', component: () => import('../views/returns/ReturnList.vue'), meta: { title: '采购退回单' } },
        { path: 'purchase-returns/:id', name: 'ReturnDetail', component: () => import('../views/returns/ReturnDetail.vue'), meta: { title: '退回单详情' } },
        { path: 'suggestions/stockout', name: 'SuggestionStockout', component: () => import('../views/suggestions/SuggestionList.vue'), meta: { title: '缺货采购' } },
        { path: 'suggestions/warning', name: 'SuggestionWarning', component: () => import('../views/suggestions/SuggestionList.vue'), meta: { title: '预警采购' } },
        { path: 'suggestions/no-stock', name: 'SuggestionNoStock', component: () => import('../views/suggestions/SuggestionList.vue'), meta: { title: '无库存采购' } },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.public) return true
  const ok = await ensureSession()
  if (!ok) {
    clearToken()
    redirectToPortal()
    return false
  }
  return true
})

export default router
