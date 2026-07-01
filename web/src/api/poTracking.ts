import client, { unwrap } from './client'

export interface ShipmentItem {
  id?: number
  poItemId: number
  skuId?: number
  qty: number
}

export interface Shipment {
  id: number
  poId: number
  shipmentNo: string
  status: string
  carrierCode?: string
  carrierName?: string
  trackingNo?: string
  shippedAt?: string
  expectedArrivalDate?: string
  deliveredAt?: string
  receiverName?: string
  receiverPhone?: string
  receiverAddress?: string
  remark?: string
  items: ShipmentItem[]
  createdAt: string
}

export interface Payment {
  id: number
  poId: number
  payAmount: number
  payMethod?: string
  payAccount?: string
  payeeAccount?: string
  payeeName?: string
  payStatus: string
  paidAt?: string
  remark?: string
  createdAt: string
}

export interface Attachment {
  id: number
  poId: number
  paymentId?: number
  fileType: string
  fileName: string
  fileUrl: string
  uploadedBy?: number
  remark?: string
  createdAt: string
}

export const SHIPMENT_STATUS_MAP: Record<string, string> = {
  pending: '待发货',
  shipped: '已发货',
  in_transit: '运输中',
  delivered: '已签收',
  exception: '异常',
}

export const ATTACHMENT_TYPE_MAP: Record<string, string> = {
  supplier_sales_order: '供货商销售单',
  payment_screenshot: '付款截图',
  contract: '合同',
  other: '其他',
}

export async function fetchShipments(poId: number) {
  return unwrap<Shipment[]>(await client.get(`/purchase-orders/${poId}/shipments`))
}

export async function createShipment(poId: number, data: Partial<Shipment> & { items?: ShipmentItem[] }) {
  return unwrap<Shipment>(await client.post(`/purchase-orders/${poId}/shipments`, data))
}

export async function updateShipmentStatus(poId: number, shipmentId: number, status: string) {
  return unwrap<Shipment>(await client.patch(`/purchase-orders/${poId}/shipments/${shipmentId}/status`, { status }))
}

export async function deleteShipment(poId: number, shipmentId: number) {
  return unwrap(await client.delete(`/purchase-orders/${poId}/shipments/${shipmentId}`))
}

export async function fetchPayments(poId: number) {
  return unwrap<Payment[]>(await client.get(`/purchase-orders/${poId}/payments`))
}

export async function createPayment(poId: number, data: Partial<Payment>) {
  return unwrap<Payment>(await client.post(`/purchase-orders/${poId}/payments`, data))
}

export async function deletePayment(poId: number, paymentId: number) {
  return unwrap(await client.delete(`/purchase-orders/${poId}/payments/${paymentId}`))
}

export async function fetchAttachments(poId: number) {
  return unwrap<Attachment[]>(await client.get(`/purchase-orders/${poId}/attachments`))
}

export async function createAttachment(poId: number, data: {
  fileType: string
  fileName: string
  fileUrl: string
  paymentId?: number
  remark?: string
}) {
  return unwrap<Attachment>(await client.post(`/purchase-orders/${poId}/attachments`, data))
}

export async function deleteAttachment(poId: number, attachmentId: number) {
  return unwrap(await client.delete(`/purchase-orders/${poId}/attachments/${attachmentId}`))
}

export async function uploadFile(file: File) {
  const form = new FormData()
  form.append('file', file)
  const res = await client.post('/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return unwrap<{ url: string; fileName: string }>(res)
}
