import client, { unwrap, type PageData } from './client'

export interface PurchaseOrderItem {
  id?: number
  skuId: number
  offerId?: number
  productName?: string
  skuCode?: string
  skuSpecs?: string
  picUrl?: string
  supplierSkuCode?: string
  qty: number
  saleUnitPrice?: number
  saleAmount?: number
  unitPrice: number
  lineAmount?: number
  receivedQty?: number
  refSoId?: number
  refOrderNo?: string
  cancelled?: boolean
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
  saleAmount?: number
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
  fulfillmentType: string
  totalAmount: number
  currency: string
  itemCount: number
  skuSpecs?: string
  refSoId?: number
  refTraceId?: string
  orderedAt?: string
  createdAt: string
}

export interface PurchaseOrderInput {
  supplierId?: number
  fulfillmentType?: string
  currency?: string
  expectedArrivalDate?: string
  warehouseId?: number
  refSoId?: number
  refTraceId?: string
  orderedAt?: string
  saleAmount?: number
  remark?: string
  items: Array<{
    skuId?: number
    offerId?: number
    productName?: string
    skuCode?: string
    skuSpecs?: string
    picUrl?: string
    supplierSkuCode?: string
    qty: number
    saleUnitPrice?: number
    saleAmount?: number
    unitPrice?: number
    remark?: string
  }>
}

export const PO_STATUS_MAP: Record<string, { label: string; type: '' | 'success' | 'warning' | 'info' | 'danger' }> = {
  draft: { label: '草稿', type: 'info' },
  ordered: { label: '已下单', type: '' },
  paid: { label: '已付款', type: 'warning' },
  partial_shipped: { label: '部分发货', type: '' },
  shipped: { label: '已发货', type: 'warning' },
  partial_received: { label: '部分到货', type: '' },
  completed: { label: '已完成', type: 'success' },
  cancelled: { label: '已取消', type: 'danger' },
}

export const PAY_STATUS_MAP: Record<string, string> = {
  unpaid: '未付款',
  partial: '部分付款',
  paid: '已付款',
}

export const FULFILLMENT_TYPE_MAP: Record<string, string> = {
  stock_in: '采购入仓',
  dropship: '代发直邮',
}

export async function fetchPurchaseOrders(params: {
  status?: string
  /** 多状态，逗号分隔，优先于 status */
  statuses?: string
  /** 付款状态，逗号分隔 unpaid|partial|paid */
  payStatus?: string
  /** 排除状态，逗号分隔 */
  excludeStatuses?: string
  fulfillmentType?: string
  supplierId?: number
  refSoId?: number
  refTraceId?: string
  keyword?: string
  createdAtStart?: string
  createdAtEnd?: string
  orderedAtStart?: string
  orderedAtEnd?: string
  sortBy?: string
  sortOrder?: string
  page?: number
  pageSize?: number
} = {}) {
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

export async function updatePurchaseOrderItemPrices(
  id: number,
  items: { itemId: number; unitPrice: number }[],
) {
  return unwrap<PurchaseOrder>(await client.put(`/purchase-orders/${id}/item-prices`, { items }))
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

export async function detachSalesOrder(data: {
  poNo: string
  orderNo?: string
  soId?: number
  reason?: string
}) {
  return unwrap<{ purchaseOrder: PurchaseOrder; unlinkWarning?: string }>(
    await client.post('/purchase-orders/detach-sales-order', data),
  )
}

export async function mergePurchaseOrders(data: { sourcePoIds: number[]; targetPoId?: number }) {
  return unwrap<PurchaseOrder & { mergedFromPoNos?: string[]; relinked?: number }>(
    await client.post('/purchase-orders/merge', data),
  )
}
