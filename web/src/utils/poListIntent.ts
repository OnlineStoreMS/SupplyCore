import type { Router } from 'vue-router'

/** 供应商订单列表进入意图（不写进 URL，避免地址栏堆参数） */
export type POListIntent = {
  fulfillmentType?: string
  status?: string
  /** 多状态，如部分发货+运输中 */
  statuses?: string[]
  /** 付款状态，如 unpaid,partial */
  payStatuses?: string[]
  /** 排除状态，如 draft,cancelled */
  excludeStatuses?: string[]
  refSoId?: number
  /** 工作台「今日」：按业务日 COALESCE(ordered_at, created_at) 筛今天 */
  today?: boolean
}

const KEY = 'supplycore.poListIntent'

type IntentListener = () => void
let intentListener: IntentListener | null = null

/** 列表页注册：同路径再次带意图进入时回调 */
export function onPOListIntent(listener: IntentListener) {
  intentListener = listener
  return () => {
    if (intentListener === listener) intentListener = null
  }
}

export function setPOListIntent(intent: POListIntent) {
  try {
    sessionStorage.setItem(KEY, JSON.stringify(intent))
  } catch {
    // ignore
  }
}

/** 读取并清除意图 */
export function takePOListIntent(): POListIntent | null {
  try {
    const raw = sessionStorage.getItem(KEY)
    if (!raw) return null
    sessionStorage.removeItem(KEY)
    return JSON.parse(raw) as POListIntent
  } catch {
    return null
  }
}

export function goPurchaseOrders(
  router: Router,
  intent: POListIntent = {},
) {
  const ft = intent.fulfillmentType
  setPOListIntent(intent)
  let path = '/purchase-orders'
  if (ft === 'dropship') path = '/purchase-orders/dropship'
  else if (ft === 'stock_in') path = '/purchase-orders/stock-in'

  const samePath = router.currentRoute.value.path === path
  void router.push(path).then(() => {
    // 同路径 push 不会触发路由 watch，需显式通知列表重新应用意图
    if (samePath) intentListener?.()
  })
}
