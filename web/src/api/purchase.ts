import client, { unwrap, type PageData } from './client'

export interface PurchaseOrderItem {
  id?: number
  skuId: number
  offerId?: number
  supplierSkuCode?: string
  qty: number
  unitPrice: number
  lineAmount?: number
  receivedQty?: number
  remark?: string
}

export interface PurchaseOrder {
  id: number
  poNo: string
  supplierId: number
  supplierName?: string
  supplierCode?: string
  status: string
  totalAmount: number
  currency: string
  expectedArrivalDate?: string
  warehouseId?: number
  fulfillmentType: string
  refSoId?: number
  refTraceId?: string
  buyerId?: number
  buyerName?: string
  payStatus: string
  remark?: string
  orderedAt?: string
  completedAt?: string
  createdAt?: string
  items: PurchaseOrderItem[]
}

export interface PurchaseOrderListItem {
  id: number
  poNo: string
  supplierId: number
  supplierName?: string
  status: string
  payStatus: string
  totalAmount: number
  currency: string
  itemCount: number
  refSoId?: number
  refTraceId?: string
  orderedAt?: string
  createdAt: string
}

export interface PurchaseOrderInput {
  supplierId: number
  fulfillmentType?: string
  currency?: string
  expectedArrivalDate?: string
  warehouseId?: number
  refSoId?: number
  remark?: string
  items: Array<{
    skuId: number
    offerId?: number
    supplierSkuCode?: string
    qty: number
    unitPrice?: number
    remark?: string
  }>
}

export const PO_STATUS_MAP: Record<string, { label: string; type: '' | 'success' | 'warning' | 'info' | 'danger' }> = {
  draft: { label: '草稿', type: 'info' },
  ordered: { label: '已下单', type: '' },
  paid: { label: '已付款', type: 'warning' },
  partial_shipped: { label: '部分发货', type: '' },
  in_transit: { label: '运输中', type: 'warning' },
  partial_received: { label: '部分到货', type: '' },
  completed: { label: '已完成', type: 'success' },
  cancelled: { label: '已取消', type: 'danger' },
}

export const PAY_STATUS_MAP: Record<string, string> = {
  unpaid: '未付款',
  partial: '部分付款',
  paid: '已付款',
}

export async function fetchPurchaseOrders(params: {
  status?: string
  supplierId?: number
  refSoId?: number
  refTraceId?: string
  keyword?: string
  page?: number
  pageSize?: number
}) {
  const res = await client.get('/purchase-orders', { params })
  return unwrap<PageData<PurchaseOrderListItem>>(res)
}

export async function fetchPurchaseOrder(id: number) {
  return unwrap<PurchaseOrder>(await client.get(`/purchase-orders/${id}`))
}

export async function createPurchaseOrder(data: PurchaseOrderInput) {
  return unwrap<PurchaseOrder>(await client.post('/purchase-orders', data))
}

export async function updatePurchaseOrder(id: number, data: PurchaseOrderInput) {
  return unwrap<PurchaseOrder>(await client.put(`/purchase-orders/${id}`, data))
}

export async function deletePurchaseOrder(id: number) {
  return unwrap(await client.delete(`/purchase-orders/${id}`))
}

export async function submitPurchaseOrder(id: number) {
  return unwrap<PurchaseOrder>(await client.post(`/purchase-orders/${id}/submit`))
}

export async function markPurchaseOrderPaid(id: number) {
  return unwrap<PurchaseOrder>(await client.post(`/purchase-orders/${id}/mark-paid`))
}

export async function completePurchaseOrder(id: number) {
  return unwrap<PurchaseOrder>(await client.post(`/purchase-orders/${id}/complete`))
}

export async function cancelPurchaseOrder(id: number) {
  return unwrap<PurchaseOrder>(await client.post(`/purchase-orders/${id}/cancel`))
}
