import type { OrderBrief, OrderItemBrief, OrderAddressBrief } from '../api/order'

function pad(n: number) {
  return String(n).padStart(2, '0')
}

/** 订单解密复制用：2026 07/11 12:42 */
export function formatOrderCopyDateTime(d: Date): string {
  return `${d.getFullYear()} ${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** 供应商订单复制地址：姓名,电话,省 市 区 详细地址 */
export function formatSupplierCopyAddress(addr?: OrderAddressBrief | null) {
  if (!addr) return ''
  if (addr.fullText?.trim()) return addr.fullText.trim()
  const name = (addr.name || '').trim()
  const phone = (addr.phone || '').trim()
  const region = [addr.province, addr.city, addr.district].filter((s) => s?.trim()).join(' ')
  const detail = (addr.address || '').trim()
  const loc = [region, detail].filter(Boolean).join(region && detail ? ' ' : '')
  return [name, phone, loc].filter(Boolean).join(',')
}

/** @deprecated 兼容旧调用；供应商复制请用 formatSupplierCopyAddress */
export function formatAddress(addr?: OrderAddressBrief | null) {
  return formatSupplierCopyAddress(addr) || '-'
}

/** 是否仍为脱敏地址（含 *） */
export function isMaskedReceiver(order: Pick<OrderBrief, 'buyerName' | 'buyerPhone' | 'address'>) {
  const text = [order.buyerName, order.buyerPhone, formatSupplierCopyAddress(order.address)].join(' ')
  return /[*＊]/.test(text)
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
}

function resolveTrackingNo(order: OrderBrief, trackingBySoId?: Record<number, string>) {
  const mapped = (trackingBySoId?.[order.id] || '').trim()
  if (mapped) return mapped
  for (const s of order.shipments || []) {
    const no = (s.expressNo || '').trim()
    if (no) return no
  }
  return ''
}

/**
 * 单笔块：规格 x数量[       ---已打快递单 单号]
 * 收件人,电话,地址
 */
export function buildSupplierCopyBlock(
  order: OrderBrief,
  opts?: { index?: number; trackingBySoId?: Record<number, string> },
) {
  const goods = formatOrderCopyGoodsLines(order.items)
  const address = formatSupplierCopyAddress(order.address)
  const tracking = resolveTrackingNo(order, opts?.trackingBySoId)
  if (!goods.length && !address) return ''

  const prefix = opts?.index != null ? `【${opts.index}】` : ''
  if (goods.length) {
    const lines = [...goods]
    if (tracking) {
      lines[lines.length - 1] = `${lines[lines.length - 1]}       ---已打快递单 ${tracking}`
    }
    lines[0] = `${prefix}${lines[0]}`
    if (address) lines.push(address)
    return lines.join('\n')
  }
  const head = `${prefix}${address}`
  return tracking ? `${head}       ---已打快递单 ${tracking}` : head
}

/** 供应商订单复制：顶部时间；多单带【1】【2】；已发货有单号则追加 ---已打快递单 */
export function buildMultiOrderCopyText(
  orders: OrderBrief[],
  now = new Date(),
  trackingBySoId?: Record<number, string>,
) {
  const multi = orders.length > 1
  const blocks = orders
    .map((o, i) =>
      buildSupplierCopyBlock(o, {
        index: multi ? i + 1 : undefined,
        trackingBySoId,
      }),
    )
    .filter((t) => t.trim())
  if (!blocks.length) return ''
  return [formatOrderCopyDateTime(now), '', blocks.join('\n\n')].join('\n')
}

/** 兼容旧单笔调用 */
export function buildOrderCopyText(order: OrderBrief, now = new Date()) {
  return buildMultiOrderCopyText([order], now)
}

/** 从代发单物流明细汇总：销售单 id → 快递单号 */
export function buildTrackingBySoId(
  items: Array<{ id?: number; refSoId?: number }>,
  shipments: Array<{ trackingNo?: string; items?: Array<{ poItemId: number }> }>,
): Record<number, string> {
  const itemToSo = new Map<number, number>()
  for (const it of items) {
    if (it.id && it.refSoId && it.refSoId > 0) itemToSo.set(it.id, it.refSoId)
  }
  const out: Record<number, string> = {}
  for (const sh of shipments) {
    const no = (sh.trackingNo || '').trim()
    if (!no) continue
    for (const si of sh.items || []) {
      const soId = itemToSo.get(si.poItemId)
      if (soId && !out[soId]) out[soId] = no
    }
  }
  return out
}

export function canDecryptOrder(order: Pick<OrderBrief, 'sourceChannel' | 'platformSysTid'>) {
  return order.sourceChannel === 'kdzs' && !!order.platformSysTid?.trim()
}

/** 复制用：只收未撤回明细关联的销售单（整单已撤回的不算） */
export function collectActiveRefSoIds(po: {
  refSoId?: number
  items?: Array<{ refSoId?: number; cancelled?: boolean }>
}): number[] {
  const items = po.items || []
  const ids: number[] = []
  const seen = new Set<number>()
  if (items.length) {
    for (const it of items) {
      if (it.cancelled) continue
      const id = Number(it.refSoId || 0)
      if (id <= 0 || seen.has(id)) continue
      seen.add(id)
      ids.push(id)
    }
    return ids
  }
  if (po.refSoId && po.refSoId > 0) return [po.refSoId]
  return []
}
