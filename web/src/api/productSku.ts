import client, { unwrap, type PageData } from './client'

export interface ProductSkuSearchItem {
  productId: number
  productName: string
  materialCode?: string
  productSn?: string
  productPic?: string
  brandName?: string
  categoryName?: string
  skuId: number
  skuCode: string
  specs: Record<string, string>
  specLabel: string
  price: number
  stock: number
  pic?: string
}

export function formatSkuOptionLabel(item: ProductSkuSearchItem): string {
  const code = item.skuCode?.trim()
  const spec = item.specLabel?.trim() || '-'
  const name = item.productName?.trim() || ''
  const parts: string[] = []
  if (code) parts.push(code)
  parts.push(spec)
  if (name) parts.push(name)
  parts.push(`#${item.skuId}`)
  return parts.join(' · ')
}

export async function searchProductSkus(params: {
  keyword: string
  page?: number
  pageSize?: number
}) {
  const res = await client.get('/product-skus/search', { params })
  return unwrap<PageData<ProductSkuSearchItem>>(res)
}
