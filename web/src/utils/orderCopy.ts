import type { OrderBrief, OrderItemBrief, OrderAddressBrief } from '../api/order'

function pad(n: number) {
  return String(n).padStart(2, '0')
}

/** 订单解密复制用：2026 07/11 12:42（与订单中心一致） */
export function formatOrderCopyDateTime(d: Date): string {
  return `${d.getFullYear()} ${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function formatAddress(addr?: OrderAddressBrief | null) {
  if (!addr) return '-'
  if (addr.fullText?.trim()) return addr.fullText.trim()
  const parts = [addr.name, addr.phone, addr.province, addr.city, addr.district, addr.address].filter(
    (s) => s?.trim(),
  )
  return parts.join(' ') || '-'
}

function looksMasked(s?: string | null) {
  return /[*＊]/.test(String(s || ''))
}

/**
 * 是否已是明文收件信息（与订单中心 orderHasPlainReceiver 一致）。
 * 优先用地址簿字段，避免单头仍残留脱敏时误判为未解密。
 */
export function hasPlainReceiver(order: Pick<OrderBrief, 'buyerName' | 'buyerPhone' | 'address'>) {
  let name = (order.buyerName || '').trim()
  let phone = (order.buyerPhone || '').trim()
  let full = ''
  let detail = ''
  if (order.address) {
    if (order.address.name?.trim()) name = order.address.name.trim()
    if (order.address.phone?.trim()) phone = order.address.phone.trim()
    full = (order.address.fullText || '').trim()
    detail = (order.address.address || '').trim()
  }
  const joined = [name, phone, full, detail].map((s) => s.trim()).filter(Boolean).join(' ')
  if (!joined) return false
  return !looksMasked(name) && !looksMasked(phone) && !looksMasked(full) && !looksMasked(detail)
}

/** 是否仍为脱敏地址（含 *） */
export function isMaskedReceiver(order: Pick<OrderBrief, 'buyerName' | 'buyerPhone' | 'address'>) {
  return !hasPlainReceiver(order)
}

const DECRYPTED_ORDERS_KEY = 'supplycore.decryptedOrders'

type DecryptedOrderCache = Record<string, OrderBrief>

function readDecryptedOrderCache(): DecryptedOrderCache {
  try {
    const raw = sessionStorage.getItem(DECRYPTED_ORDERS_KEY)
    if (!raw) return {}
    const obj = JSON.parse(raw) as unknown
    if (!obj || typeof obj !== 'object') return {}
    return obj as DecryptedOrderCache
  } catch {
    return {}
  }
}

function writeDecryptedOrderCache(cache: DecryptedOrderCache) {
  sessionStorage.setItem(DECRYPTED_ORDERS_KEY, JSON.stringify(cache))
}

/** 缓存本会话已解密的销售单明文；复制时优先用缓存，避免再次打解密接口 */
export function markOrdersDecrypted(orders: Array<number | OrderBrief>) {
  try {
    const cache = readDecryptedOrderCache()
    for (const item of orders) {
      if (typeof item === 'number') {
        if (item > 0 && !cache[String(item)]) {
          cache[String(item)] = { id: item } as OrderBrief
        }
        continue
      }
      if (item?.id > 0) {
        cache[String(item.id)] = item
      }
    }
    writeDecryptedOrderCache(cache)
  } catch {
    // ignore
  }
}

export function getCachedDecryptedOrder(orderId: number): OrderBrief | null {
  if (orderId <= 0) return null
  return readDecryptedOrderCache()[String(orderId)] || null
}

export function wasOrderDecrypted(orderId: number) {
  const cached = getCachedDecryptedOrder(orderId)
  return !!cached && hasPlainReceiver(cached)
}

/** 用会话内明文覆盖接口返回的脱敏单 */
export function applyDecryptedCache(orders: OrderBrief[]) {
  return orders.map((o) => {
    const cached = getCachedDecryptedOrder(o.id)
    if (!cached || !hasPlainReceiver(cached)) return o
    return { ...o, ...cached, address: cached.address || o.address, items: cached.items || o.items }
  })
}

export function formatOrderCopyGoodsLines(items?: OrderItemBrief[]) {
  return (items || [])
    .map((it) => {
      const spec = (it.skuSpecs || it.productName || it.skuCode || '').trim()
      if (!spec) return ''
      const num = it.quantity && it.quantity > 0 ? it.quantity : 1
      return `${spec} x${num}`
    })
    .filter(Boolean)
    .join('\n')
}

/** 与订单中心一致：时间 + 空行 + 收件信息 + --- + 规格行 */
export function buildOrderCopyText(order: OrderBrief, now = new Date()) {
  const body = buildOrderCopyBody(order)
  if (!body) return ''
  return [formatOrderCopyDateTime(now), '', body].join('\n')
}

/** 单笔内容（不含顶部时间）：收件信息 + --- + 规格行 */
export function buildOrderCopyBody(order: OrderBrief) {
  const address = formatAddress(order.address)
  const goodsBlock = formatOrderCopyGoodsLines(order.items)
  const lines = [address === '-' ? '' : address, '---']
  if (goodsBlock) lines.push(goodsBlock)
  return lines.join('\n')
}

/**
 * 多销售单复制：顶部一个时间，各单用【1】【2】标注。
 * 单条时不带序号，与订单中心单笔格式一致。
 */
export function buildMultiOrderCopyText(orders: OrderBrief[], now = new Date()) {
  const bodies = orders.map((o) => buildOrderCopyBody(o)).filter((t) => t.trim())
  if (!bodies.length) return ''
  if (bodies.length === 1) {
    return [formatOrderCopyDateTime(now), '', bodies[0]].join('\n')
  }
  const numbered = bodies.map((body, i) => `【${i + 1}】\n${body}`).join('\n\n')
  return [formatOrderCopyDateTime(now), '', numbered].join('\n')
}

export function canDecryptOrder(order: Pick<OrderBrief, 'sourceChannel' | 'platformSysTid'>) {
  return order.sourceChannel === 'kdzs' && !!order.platformSysTid?.trim()
}
