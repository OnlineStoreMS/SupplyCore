/** 供应商订单列表筛选记忆（进详情再返回时恢复；工作台意图 / 手动改筛会覆盖） */

export type POListFilterSnapshot = {
  status: string
  statusesFilter: string
  payStatusFilter: string
  excludeStatusesFilter: string
  supplierId?: number
  keyword: string
  refSoId?: number
  createdRange: [string, string] | null
  orderedRange: [string, string] | null
  page: number
  pageSize: number
  sortBy: string
  sortOrder: 'asc' | 'desc'
}

function storageKey(fulfillmentType: string) {
  const ft = fulfillmentType || 'all'
  return `supplycore.poListFilters.${ft}`
}

export function savePOListFilters(fulfillmentType: string, snap: POListFilterSnapshot) {
  try {
    sessionStorage.setItem(storageKey(fulfillmentType), JSON.stringify(snap))
  } catch {
    // ignore
  }
}

export function loadPOListFilters(fulfillmentType: string): POListFilterSnapshot | null {
  try {
    const raw = sessionStorage.getItem(storageKey(fulfillmentType))
    if (!raw) return null
    const obj = JSON.parse(raw) as POListFilterSnapshot
    if (!obj || typeof obj !== 'object') return null
    return obj
  } catch {
    return null
  }
}

export function clearPOListFilters(fulfillmentType: string) {
  try {
    sessionStorage.removeItem(storageKey(fulfillmentType))
  } catch {
    // ignore
  }
}
