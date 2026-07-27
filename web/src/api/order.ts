import client, { unwrap, type PageData } from './client'

export interface OrderBrief {
  id: number
  orderNo: string
  platformOrderId?: string
  shopName?: string
  buyerName?: string
  buyerNick?: string
  status?: string
  shipStatus?: string
  totalAmount?: number
  payAmount?: number
  orderedAt?: string
  createdAt?: string
}

export async function searchOrders(params: {
  keyword: string
  page?: number
  pageSize?: number
}) {
  const res = await client.get('/orders/search', { params })
  return unwrap<PageData<OrderBrief>>(res)
}
