import client, { unwrap } from './client'
import type { PurchaseOrderListItem } from './purchase'

export interface DashboardWorkbench {
  dropshipPO: number
  stockInPO: number
  draftPO: number
  orderedPO: number
  unpaidPO: number
  inTransitPO: number
  partialReceivedPO: number
  activeOffers: number
  todayDropshipSaleAmount: number
  todayDropshipPurchaseAmount: number
  todayDropshipProfit: number
}

export interface DashboardSupplierStats {
  total: number
  active: number
  offerCount: number
  orderedThisMonth: number
}

export interface DashboardPOStats {
  total: number
  draft: number
  inProgress: number
  completed: number
  cancelled: number
  todayCount: number
  weekCount: number
  monthCount: number
}

export interface DashboardCostStats {
  todayAmount: number
  weekAmount: number
  monthAmount: number
  unpaidAmount: number
  yearAmount: number
}

export interface DashboardSupplierRank {
  supplierId: number
  supplierName: string
  orderCount: number
  totalAmount: number
}

export interface DashboardStatusCount {
  status: string
  count: number
}

export interface DashboardStats {
  workbench: DashboardWorkbench
  supplier: DashboardSupplierStats
  purchaseOrder: DashboardPOStats
  cost: DashboardCostStats
  topSuppliers: DashboardSupplierRank[]
  recentOrders: PurchaseOrderListItem[]
  statusBreakdown: DashboardStatusCount[]
}

export async function fetchDashboardStats() {
  return unwrap<DashboardStats>(await client.get('/dashboard/stats'))
}

export interface DashboardTrendPoint {
  date: string
  orderCount: number
  saleAmount: number
  purchaseAmount: number
  profit: number
}

export interface DashboardTrend {
  startDate: string
  endDate: string
  orderCount: number
  saleAmount: number
  purchaseAmount: number
  profit: number
  points: DashboardTrendPoint[]
}

export async function fetchDashboardTrend(params: { startDate?: string; endDate?: string } = {}) {
  return unwrap<DashboardTrend>(await client.get('/dashboard/trend', { params }))
}
