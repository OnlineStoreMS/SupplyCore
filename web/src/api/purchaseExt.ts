import client, { unwrap, type PageData } from './client'

export interface PurchaseAccount {
  id: number
  channel: string
  accountAlias: string
  accountName: string
  isPrimary: boolean
  status: string
  authStatus: string
  lastSyncAt?: string
  operatorName: string
  remark: string
  createdAt: string
}

export async function fetchPurchaseAccounts(keyword?: string, page = 1, pageSize = 20) {
  const res = await client.get('/purchase-accounts', { params: { keyword, page, pageSize } })
  return unwrap<PageData<PurchaseAccount>>(res)
}

export async function createPurchaseAccount(data: Partial<PurchaseAccount>) {
  const res = await client.post('/purchase-accounts', data)
  return unwrap<PurchaseAccount>(res)
}

export async function updatePurchaseAccount(id: number, data: Partial<PurchaseAccount>) {
  const res = await client.put(`/purchase-accounts/${id}`, data)
  return unwrap<PurchaseAccount>(res)
}

export async function deletePurchaseAccount(id: number) {
  await client.delete(`/purchase-accounts/${id}`)
}

export interface PurchaseInboundListItem {
  id: number
  inboundNo: string
  status: string
  poId: number
  poNo: string
  supplierId: number
  supplierName: string
  warehouseId: number
  warehouseName: string
  totalQty: number
  totalAmount: number
  trackingNo: string
  creatorName: string
  createdAt: string
}

export interface PurchaseInboundDetail extends PurchaseInboundListItem {
  platformOrderNo: string
  buyerName: string
  remark: string
  whAuditorName: string
  whAuditedAt?: string
  finAuditorName: string
  finAuditedAt?: string
  items: {
    id: number
    poItemId: number
    skuId: number
    skuCode: string
    skuName: string
    purchaseQty: number
    qcQty: number
    rejectQty: number
    inboundQty: number
    unitPrice: number
    lineAmount: number
    locationCode: string
    remark: string
  }[]
}

export const INBOUND_STATUS_MAP: Record<string, { label: string; type: string }> = {
  draft: { label: '草稿', type: 'info' },
  pending_wh: { label: '待入库审核', type: 'warning' },
  pending_finance: { label: '待财务审核', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  void: { label: '已作废', type: 'danger' },
}

export async function fetchPurchaseInbounds(params?: {
  status?: string
  keyword?: string
  page?: number
  pageSize?: number
}) {
  const res = await client.get('/purchase-inbounds', { params })
  return unwrap<PageData<PurchaseInboundListItem>>(res)
}

export async function fetchPurchaseInbound(id: number) {
  const res = await client.get(`/purchase-inbounds/${id}`)
  return unwrap<PurchaseInboundDetail>(res)
}

export async function createPurchaseInbound(data: Record<string, unknown>) {
  const res = await client.post('/purchase-inbounds', data)
  return unwrap<PurchaseInboundDetail>(res)
}

export async function approveInboundWH(id: number) {
  const res = await client.post(`/purchase-inbounds/${id}/approve-wh`)
  return unwrap<PurchaseInboundDetail>(res)
}

export async function approveInboundFinance(id: number) {
  const res = await client.post(`/purchase-inbounds/${id}/approve-finance`)
  return unwrap<PurchaseInboundDetail>(res)
}

export async function voidInbound(id: number) {
  await client.post(`/purchase-inbounds/${id}/void`)
}

export interface PackageReceive {
  id: number
  warehouseId: number
  warehouseName: string
  carrier: string
  trackingNo: string
  packageType: string
  poId: number
  poNo: string
  inboundId: number
  inboundNo: string
  scannerName: string
  remark: string
  createdAt: string
}

export async function fetchPackageReceives(keyword?: string, page = 1, pageSize = 20) {
  const res = await client.get('/package-receives', { params: { keyword, page, pageSize } })
  return unwrap<PageData<PackageReceive>>(res)
}

export async function scanPackage(data: Partial<PackageReceive> & { trackingNo: string }) {
  const res = await client.post('/package-receives/scan', data)
  return unwrap<PackageReceive>(res)
}

export async function createInboundFromPackage(id: number) {
  const res = await client.post(`/package-receives/${id}/create-inbound`)
  return unwrap<PurchaseInboundDetail>(res)
}

export interface PurchaseReturnListItem {
  id: number
  returnNo: string
  status: string
  inboundId: number
  inboundNo: string
  supplierId: number
  supplierName: string
  warehouseName: string
  totalQty: number
  totalAmount: number
  actualAmount: number
  creatorName: string
  createdAt: string
}

export interface PurchaseReturnDetail extends PurchaseReturnListItem {
  trackingNo: string
  buyerName: string
  remark: string
  auditorName: string
  auditedAt?: string
  finAuditorName: string
  finAuditedAt?: string
  items: {
    id: number
    inboundItemId: number
    skuId: number
    skuCode: string
    skuName: string
    originalQty: number
    returnQty: number
    unitPrice: number
    returnAmount: number
    actualAmount: number
    remark: string
  }[]
}

export const RETURN_STATUS_MAP: Record<string, { label: string; type: string }> = {
  draft: { label: '草稿', type: 'info' },
  pending_return: { label: '待退回审核', type: 'warning' },
  pending_finance: { label: '待财务审核', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  void: { label: '已作废', type: 'danger' },
}

export async function fetchPurchaseReturns(params?: {
  status?: string
  keyword?: string
  page?: number
  pageSize?: number
}) {
  const res = await client.get('/purchase-returns', { params })
  return unwrap<PageData<PurchaseReturnListItem>>(res)
}

export async function fetchPurchaseReturn(id: number) {
  const res = await client.get(`/purchase-returns/${id}`)
  return unwrap<PurchaseReturnDetail>(res)
}

export async function createPurchaseReturn(data: Record<string, unknown>) {
  const res = await client.post('/purchase-returns', data)
  return unwrap<PurchaseReturnDetail>(res)
}

export async function approveReturn(id: number) {
  const res = await client.post(`/purchase-returns/${id}/approve`)
  return unwrap<PurchaseReturnDetail>(res)
}

export async function approveReturnFinance(id: number) {
  const res = await client.post(`/purchase-returns/${id}/approve-finance`)
  return unwrap<PurchaseReturnDetail>(res)
}

export async function voidReturn(id: number) {
  await client.post(`/purchase-returns/${id}/void`)
}

export interface SuggestionItem {
  skuId: number
  skuCode: string
  skuName: string
  supplierId: number
  supplierName: string
  stockoutQty: number
  salesQty: number
  suggestPurchase: number
  unitPrice: number
  offerId: number
  source: string
}

export async function fetchPurchaseSuggestions(source: 'stockout' | 'warning' | 'no_stock') {
  const res = await client.get('/purchase-suggestions', { params: { source } })
  return unwrap<SuggestionItem[]>(res)
}
