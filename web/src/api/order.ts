import client, { unwrap, type PageData } from './client'

export interface OrderItemBrief {
  id: number
  skuId?: number
  skuCode?: string
  productName?: string
  skuSpecs?: string
  picUrl?: string
  quantity: number
  price?: number
  totalAmount?: number
}

export interface OrderAddressBrief {
  name?: string
  phone?: string
  province?: string
  city?: string
  district?: string
  address?: string
  fullText?: string
}

export interface OrderShipmentBrief {
  id: number
  shipmentNo?: string
  expressCompany?: string
  expressNo?: string
  callbackStatus?: string
  callbackMessage?: string
  shippedAt?: string
  remark?: string
}

export interface OrderBrief {
  id: number
  orderNo: string
  sourceChannel?: string
  platform?: string
  platformOrderId?: string
  platformSysTid?: string
  shopName?: string
  buyerName?: string
  buyerNick?: string
  buyerPhone?: string
  status?: string
  shipStatus?: string
  totalAmount?: number
  payAmount?: number
  remark?: string
  sellerRemark?: string
  orderedAt?: string
  createdAt?: string
  address?: OrderAddressBrief
  items?: OrderItemBrief[]
  shipments?: OrderShipmentBrief[]
}

/** 从发货回传结果取出提示文案（失败时尽量保留快递助手/平台原始报错） */
export function shipCallbackTip(order: OrderBrief | null | undefined, expressNo?: string) {
  const want = (expressNo || '').trim()
  const list = order?.shipments || []
  let sh = want ? list.find((s) => (s.expressNo || '').trim() === want) : undefined
  if (!sh && list.length) {
    sh = [...list].sort((a, b) => (b.id || 0) - (a.id || 0))[0]
  }
  const status = (sh?.callbackStatus || '').toLowerCase()
  const raw = (sh?.callbackMessage || '').trim()
  if (status === 'failed') {
    return { ok: false as const, message: raw || '平台回传失败（无详细报错）' }
  }
  if (status === 'succeeded') {
    return { ok: true as const, message: raw || '已回传电商平台' }
  }
  if (status === 'skipped') {
    return { ok: true as const, message: raw || '已记录发货（未回传平台）' }
  }
  return { ok: true as const, message: '已回传订单中心' }
}

export function formatOrderReceiverAddress(addr?: OrderAddressBrief | null) {
  if (!addr) return ''
  if (addr.fullText?.trim()) return addr.fullText.trim()
  return [addr.province, addr.city, addr.district, addr.address].filter((s) => s?.trim()).join('')
}

export async function searchOrders(params: {
  keyword: string
  page?: number
  pageSize?: number
}) {
  const res = await client.get('/orders/search', { params })
  return unwrap<PageData<OrderBrief>>(res)
}

export async function fetchOrder(id: number) {
  return unwrap<OrderBrief>(await client.get(`/orders/${id}`))
}

export async function decryptOrders(orderIds: number[]) {
  return unwrap<{ items: OrderBrief[]; success: number }>(
    await client.post('/orders/decrypt', { orderIds }),
  )
}

export async function shipOrder(
  id: number,
  body: { expressCompany: string; expressNo: string; remark?: string; callback?: boolean },
) {
  // 回传快递助手可能较慢（查单+发货+核对），避免默认 30s 超时中断
  return unwrap<OrderBrief>(await client.post(`/orders/${id}/ship`, body, { timeout: 180000 }))
}
