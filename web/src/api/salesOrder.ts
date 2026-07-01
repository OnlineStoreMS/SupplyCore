import client, { unwrap, type PageData } from './client'

export interface SalesOrderItem {
  id?: number
  skuId: number
  qty: number
  fulfillmentMode?: string
  selectedOfferId?: number
  linkedPoId?: number
  linkedPoNo?: string
  remark?: string
}

export interface SalesOrder {
  id: number
  soNo: string
  traceId: string
  status: string
  sourceChannel?: string
  receiverName?: string
  receiverPhone?: string
  province?: string
  city?: string
  district?: string
  receiverAddress?: string
  remark?: string
  createdAt?: string
  items: SalesOrderItem[]
}

export interface SalesOrderListItem {
  id: number
  soNo: string
  traceId: string
  status: string
  sourceChannel?: string
  itemCount: number
  createdAt: string
}

export interface SalesOrderInput {
  sourceChannel?: string
  receiverName?: string
  receiverPhone?: string
  province?: string
  city?: string
  district?: string
  receiverAddress?: string
  remark?: string
  items: Array<{
    skuId: number
    qty: number
    fulfillmentMode?: string
    remark?: string
  }>
}

export interface SupplyOptionOffer {
  offerId: number
  supplierId: number
  supplierName?: string
  supplierCode?: string
  supplierSkuCode?: string
  supplyPrice: number
  currency?: string
  supportsDropship?: boolean
  isPrimary?: boolean
  priority?: number
}

export interface SourcingLinePlan {
  soItemId: number
  skuId: number
  qty: number
  fulfillmentMode: string
  needsPo: boolean
  message?: string
  offers?: SupplyOptionOffer[]
  recommendedOfferId?: number
}

export interface SourcingEvaluateResp {
  soId: number
  traceId: string
  lines: SourcingLinePlan[]
}

export interface SourcingCreatePOResp {
  refSoId: number
  refTraceId: string
  poIds: number[]
  poNos: string[]
}

export const SO_STATUS_MAP: Record<string, { label: string; type: '' | 'success' | 'warning' | 'info' | 'danger' }> = {
  draft: { label: '草稿', type: 'info' },
  confirmed: { label: '已确认', type: '' },
  sourced: { label: '已寻源', type: 'success' },
  cancelled: { label: '已取消', type: 'danger' },
}

export const FULFILLMENT_MODE_MAP: Record<string, string> = {
  dropship: '代发',
  self: '自发货',
}

export async function fetchSalesOrders(params: {
  status?: string
  keyword?: string
  page?: number
  pageSize?: number
}) {
  const res = await client.get('/sales-orders', { params })
  return unwrap<PageData<SalesOrderListItem>>(res)
}

export async function fetchSalesOrder(id: number) {
  return unwrap<SalesOrder>(await client.get(`/sales-orders/${id}`))
}

export async function createSalesOrder(data: SalesOrderInput) {
  return unwrap<SalesOrder>(await client.post('/sales-orders', data))
}

export async function updateSalesOrder(id: number, data: SalesOrderInput) {
  return unwrap<SalesOrder>(await client.put(`/sales-orders/${id}`, data))
}

export async function deleteSalesOrder(id: number) {
  return unwrap(await client.delete(`/sales-orders/${id}`))
}

export async function confirmSalesOrder(id: number) {
  return unwrap<SalesOrder>(await client.post(`/sales-orders/${id}/confirm`))
}

export async function evaluateSourcing(id: number) {
  return unwrap<SourcingEvaluateResp>(await client.post(`/sales-orders/${id}/evaluate-sourcing`))
}

export async function createPurchaseOrdersFromSO(
  id: number,
  data: {
    autoSubmit?: boolean
    selections: Array<{ soItemId: number; offerId: number; qty?: number }>
  },
) {
  return unwrap<SourcingCreatePOResp>(await client.post(`/sales-orders/${id}/create-purchase-orders`, data))
}
