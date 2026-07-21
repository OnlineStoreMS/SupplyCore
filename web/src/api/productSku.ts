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

export function skuDisplayPic(item?: Pick<ProductSkuSearchItem, 'pic' | 'productPic'> | null): string {
  return item?.pic || item?.productPic || ''
}

const skuResolveCache = new Map<number, ProductSkuSearchItem>()

export async function searchProductSkus(params: {
  keyword: string
  page?: number
  pageSize?: number
}) {
  const res = await client.get('/product-skus/search', { params })
  const page = unwrap<PageData<ProductSkuSearchItem>>(res)
  for (const item of page.list) {
    skuResolveCache.set(item.skuId, item)
  }
  return page
}

/** 批量解析 SKU 详情（图片、规格等），带内存缓存 */
export async function resolveProductSkus(ids: number[]): Promise<Map<number, ProductSkuSearchItem>> {
  const unique = [...new Set(ids.filter((id) => id > 0))]
  const missing = unique.filter((id) => !skuResolveCache.has(id))
  await Promise.all(
    missing.map(async (id) => {
      try {
        const data = await searchProductSkus({ keyword: String(id), page: 1, pageSize: 10 })
        const hit = data.list.find((item) => item.skuId === id)
        if (hit) {
          skuResolveCache.set(id, hit)
        }
      } catch {
        // ignore single resolve failure
      }
    }),
  )
  const result = new Map<number, ProductSkuSearchItem>()
  for (const id of unique) {
    const hit = skuResolveCache.get(id)
    if (hit) result.set(id, hit)
  }
  return result
}
