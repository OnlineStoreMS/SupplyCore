<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Edit } from '@element-plus/icons-vue'
import {
  fetchPurchaseOrder,
  submitPurchaseOrder,
  markPurchaseOrderPaid,
  completePurchaseOrder,
  cancelPurchaseOrder,
  deletePurchaseOrder,
  updatePurchaseOrderItemPrices,
  detachSalesOrder,
  PO_STATUS_MAP,
  PAY_STATUS_MAP,
  type PurchaseOrder,
  type PurchaseOrderItem,
} from '../../api/purchase'
import { resolveProductSkus, type ProductSkuSearchItem } from '../../api/productSku'
import { fetchShipments, type Shipment } from '../../api/poTracking'
import { decryptOrders, fetchOrder, type OrderBrief } from '../../api/order'
import { copyToClipboard } from '../../utils/clipboard'
import {
  buildMultiOrderCopyText,
  canDecryptOrder,
  isMaskedReceiver,
} from '../../utils/orderCopy'
import PoShipmentTab from './PoShipmentTab.vue'
import PoPaymentTab from './PoPaymentTab.vue'
import PoAttachmentTab from './PoAttachmentTab.vue'
import { orderCoreOrderUrl } from '../../utils/runtimeConfig'

const route = useRoute()
const router = useRouter()
const poId = computed(() => Number(route.params.id))
const activeTab = ref('overview')

const loading = ref(false)
const acting = ref(false)
const savingPrice = ref(false)
const decrypting = ref(false)
const copying = ref(false)
const po = ref<PurchaseOrder | null>(null)
const skuMap = ref<Map<number, ProductSkuSearchItem>>(new Map())
const shipments = ref<Shipment[]>([])

/** 订单完成前可改采购价（已完成/已取消不可改；与付款状态无关） */
const canEditPurchasePrice = computed(() => {
  if (!po.value) return false
  if (po.value.status === 'completed' || po.value.status === 'cancelled') return false
  return true
})

function roundMoney(n: number) {
  return Math.round((n + Number.EPSILON) * 100) / 100
}

async function saveItemUnitPrice(row: PurchaseOrderItem & { id?: number }, unitPrice: number) {
  if (!po.value || !row.id || row.cancelled || !canEditPurchasePrice.value) return
  if (isSplitChildItem(row)) {
    ElMessage.warning('拆分子行不参与采购计价')
    return
  }
  if (Number.isNaN(unitPrice) || unitPrice < 0) {
    ElMessage.warning('采购金额不能为负数')
    await loadData()
    return
  }
  savingPrice.value = true
  try {
    const price = roundMoney(unitPrice)
    row.unitPrice = price
    row.lineAmount = roundMoney(price * (row.qty || 0))
    po.value = await updatePurchaseOrderItemPrices(poId.value, [{ itemId: row.id, unitPrice: price }])
  } catch (e) {
    ElMessage.error((e as Error).message || '保存采购价失败')
    await loadData()
  } finally {
    savingPrice.value = false
  }
}

/** 改单价 → 按数量重算小计 */
async function onUnitPriceChange(row: PurchaseOrderItem & { id?: number }) {
  await saveItemUnitPrice(row, Number(row.unitPrice))
}

/** 改小计 → 反推单价（单价 = 小计 / 数量） */
async function onLineAmountChange(row: PurchaseOrderItem & { id?: number }) {
  const amount = Number(row.lineAmount)
  const qty = Number(row.qty) || 0
  if (qty <= 0) {
    ElMessage.warning('数量无效，无法反算单价')
    await loadData()
    return
  }
  await saveItemUnitPrice(row, amount / qty)
}

function isSplitChildItem(it?: PurchaseOrderItem | null) {
  if (!it) return false
  return !!(it.splitKind || (it.parentPoItemId && it.parentPoItemId > 0))
}

/** 基本信息采购明细：父行 + └ 拆分子行（对齐物流/订单中心） */
const itemTreeRows = computed(() => {
  const items = po.value?.items || []
  type Row = {
    key: string
    item: PurchaseOrderItem
    isSplitChild: boolean
    isSplitParent: boolean
  }
  const childrenByParent = new Map<number, PurchaseOrderItem[]>()
  const fullChildren: PurchaseOrderItem[] = []
  const roots: PurchaseOrderItem[] = []
  for (const it of items) {
    if (it.splitKind === 'full') {
      fullChildren.push(it)
      continue
    }
    if (it.splitKind === 'partial' && it.parentPoItemId) {
      const list = childrenByParent.get(it.parentPoItemId) || []
      list.push(it)
      childrenByParent.set(it.parentPoItemId, list)
      continue
    }
    roots.push(it)
  }
  const out: Row[] = []
  for (const root of roots) {
    const kids = childrenByParent.get(root.id || 0) || []
    out.push({
      key: `root-${root.id}`,
      item: root,
      isSplitChild: false,
      isSplitParent: kids.length > 0,
    })
    for (const ch of kids) {
      out.push({
        key: `child-${ch.id}`,
        item: ch,
        isSplitChild: true,
        isSplitParent: false,
      })
    }
  }
  if (fullChildren.length) {
    for (const ch of fullChildren) {
      out.push({
        key: `full-${ch.id}`,
        item: ch,
        isSplitChild: true,
        isSplitParent: false,
      })
    }
  }
  return out
})

function openOrderCore(soId?: number) {
  const id = soId && soId > 0 ? soId : po.value?.refSoId
  if (!id) return
  window.open(orderCoreOrderUrl(id), '_blank', 'noopener,noreferrer')
}

function splitRefOrders(trace?: string) {
  if (!trace?.trim()) return [] as string[]
  return trace
    .split(/[,，;\s]+/)
    .map((s) => s.trim())
    .filter(Boolean)
}

/** 关联订单：优先从明细汇总（含已撤回），便于整单撤销后仍能看到历史关联 */
const refOrders = computed(() => {
  const items = po.value?.items || []
  const map = new Map<string, { cancelled: boolean; soId: number }>()
  for (const it of items) {
    const no = (it.refOrderNo || '').trim()
    if (!no) continue
    const prev = map.get(no)
    if (!prev) {
      map.set(no, { cancelled: !!it.cancelled, soId: Number(it.refSoId || 0) })
    } else {
      // 任一明细未撤回，则该销售单视为仍关联
      prev.cancelled = prev.cancelled && !!it.cancelled
      if (!prev.soId && it.refSoId) prev.soId = Number(it.refSoId)
    }
  }
  if (map.size > 0) {
    return [...map.entries()].map(([no, v]) => ({ no, cancelled: v.cancelled, soId: v.soId }))
  }
  const allCancelled = po.value?.status === 'cancelled'
  return splitRefOrders(po.value?.refTraceId).map((no) => ({
    no,
    cancelled: allCancelled,
    soId: 0,
  }))
})

const activeRefOrders = computed(() => refOrders.value.filter((r) => !r.cancelled))
const allRefWithdrawn = computed(
  () => refOrders.value.length > 0 && activeRefOrders.value.length === 0,
)

const canDetachSales = computed(
  () => po.value?.fulfillmentType === 'dropship' && po.value.status !== 'cancelled',
)

async function handleDetachSales(r: { no: string; soId: number }) {
  if (!po.value || !canDetachSales.value || !r.no) return
  const paidLike =
    po.value.payStatus === 'paid' ||
    po.value.payStatus === 'partial' ||
    ['awaiting_ship', 'paid', 'partial_shipped', 'shipped', 'partial_received', 'completed'].includes(po.value.status)
  const tip = paidLike
    ? `确定从本代发单解绑销售单 ${r.no}？\n单据已付款/履约，仅划线解绑并回写订单中心为待分配，不冲销付款记录。`
    : `确定从本代发单解绑销售单 ${r.no}？将同步清空订单中心分配。`
  try {
    await ElMessageBox.confirm(tip, '解绑销售单', { type: 'warning' })
  } catch {
    return
  }
  acting.value = true
  try {
    const res = await detachSalesOrder({
      poNo: po.value.poNo,
      orderNo: r.no,
      soId: r.soId || undefined,
      reason: '供应链手工解绑',
    })
    po.value = res.purchaseOrder
    if (po.value?.items) {
      skuMap.value = await resolveProductSkus(po.value.items.map((it) => it.skuId))
    }
    if (res.unlinkWarning) {
      ElMessage.warning(`已解绑代发明细，但回写订单中心失败：${res.unlinkWarning}`)
    } else {
      ElMessage.success('已解绑，订单中心已恢复待分配')
    }
  } catch (e) {
    ElMessage.error((e as Error).message || '解绑失败')
  } finally {
    acting.value = false
  }
}

const trackable = computed(() => po.value && po.value.status !== 'draft' && po.value.status !== 'cancelled')

const logisticsByItem = computed(() => {
  const map = new Map<number, string[]>()
  for (const sh of shipments.value) {
    const tracking = [sh.carrierName, sh.trackingNo].filter(Boolean).join(' ')
    if (!tracking) continue
    for (const it of sh.items || []) {
      const arr = map.get(it.poItemId) || []
      if (!arr.includes(tracking)) arr.push(tracking)
      map.set(it.poItemId, arr)
    }
  }
  return map
})

async function loadData() {
  loading.value = true
  try {
    po.value = await fetchPurchaseOrder(poId.value)
    skuMap.value = await resolveProductSkus((po.value.items || []).map((it) => it.skuId))
    if (po.value.status !== 'draft' && po.value.status !== 'cancelled') {
      shipments.value = await fetchShipments(poId.value)
    } else {
      shipments.value = []
    }
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function statusLabel(s: string) {
  return PO_STATUS_MAP[s]?.label || s
}

function statusType(s: string) {
  return PO_STATUS_MAP[s]?.type || 'info'
}

function lineSkuCode(row: { skuId: number; skuCode?: string }) {
  const code = row.skuCode?.trim()
  if (code) return code
  return skuMap.value.get(row.skuId)?.skuCode?.trim() || '—'
}

function itemRowClass({ row }: { row: { item?: PurchaseOrderItem; cancelled?: boolean } }) {
  const cancelled = row.item?.cancelled ?? row.cancelled
  return cancelled ? 'po-item-cancelled' : ''
}

async function doAction(label: string, fn: () => Promise<unknown>) {
  try {
    await ElMessageBox.confirm(`确定${label}？`, '确认')
  } catch {
    return
  }
  acting.value = true
  try {
    po.value = (await fn()) as PurchaseOrder
    if (po.value?.items) {
      skuMap.value = await resolveProductSkus(po.value.items.map((it) => it.skuId))
    }
    if (po.value && po.value.status !== 'draft' && po.value.status !== 'cancelled') {
      shipments.value = await fetchShipments(poId.value)
    } else {
      shipments.value = []
    }
    ElMessage.success('操作成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '操作失败')
  } finally {
    acting.value = false
  }
}

async function handleDelete() {
  await doAction('删除此供应商订单（物流/付款等一并删除）', async () => {
    await deletePurchaseOrder(poId.value)
    router.push('/purchase-orders')
    return null
  })
}

function collectRefSoIds(order: PurchaseOrder): number[] {
  const ids = new Set<number>()
  if (order.refSoId && order.refSoId > 0) ids.add(order.refSoId)
  for (const it of order.items || []) {
    if (it.refSoId && it.refSoId > 0) ids.add(it.refSoId)
  }
  return [...ids]
}

async function loadLinkedOrders(): Promise<OrderBrief[]> {
  if (!po.value) throw new Error('采购单未加载')
  const ids = collectRefSoIds(po.value)
  if (!ids.length) throw new Error('该代发单未关联销售订单')
  const orders: OrderBrief[] = []
  for (const id of ids) {
    orders.push(await fetchOrder(id))
  }
  return orders
}

async function handleDecrypt() {
  if (po.value?.fulfillmentType !== 'dropship') return
  decrypting.value = true
  try {
    const orders = await loadLinkedOrders()
    const ecommerce = orders.filter((o) => canDecryptOrder(o))
    if (!ecommerce.length) {
      ElMessage.warning('关联销售单中无可解密的电商订单')
      return
    }
    const data = await decryptOrders(ecommerce.map((o) => o.id))
    ElMessage.success(
      data.success > 1 ? `已依次解密 ${data.success} 笔电商订单` : '解密成功',
    )
  } catch (e) {
    ElMessage.error((e as Error).message || '解密失败')
  } finally {
    decrypting.value = false
  }
}

async function handleCopy() {
  if (po.value?.fulfillmentType !== 'dropship') return
  copying.value = true
  try {
    let orders = await loadLinkedOrders()
    const needDecrypt = orders.filter((o) => canDecryptOrder(o) && isMaskedReceiver(o))
    if (needDecrypt.length) {
      const data = await decryptOrders(needDecrypt.map((o) => o.id))
      const byId = new Map((data.items || []).map((o) => [o.id, o]))
      orders = orders.map((o) => byId.get(o.id) || o)
    }
    const text = buildMultiOrderCopyText(orders)
    if (!text.trim()) {
      ElMessage.warning('暂无收件信息可复制')
      return
    }
    const ok = await copyToClipboard(text)
    if (ok) {
      ElMessage.success(orders.length > 1 ? `已复制 ${orders.length} 笔（已标序号）` : '已复制')
    } else {
      ElMessage.error('复制失败')
    }
  } catch (e) {
    ElMessage.error((e as Error).message || '复制失败')
  } finally {
    copying.value = false
  }
}
</script>

<template>
  <div v-loading="loading" class="po-detail">
    <div class="top-bar">
      <el-button :icon="ArrowLeft" text @click="router.push('/purchase-orders')">返回列表</el-button>
      <div v-if="po" class="actions">
        <template v-if="po.fulfillmentType === 'dropship'">
          <el-button type="warning" plain :loading="decrypting" @click="handleDecrypt">解密</el-button>
          <el-button :loading="copying" @click="handleCopy">复制</el-button>
        </template>
        <el-button
          v-if="po.status === 'draft'"
          :icon="Edit"
          @click="router.push(`/purchase-orders/${po.id}/edit`)"
        >
          编辑
        </el-button>
        <el-button
          v-if="po.status !== 'completed'"
          type="danger"
          plain
          :loading="acting"
          @click="handleDelete"
        >
          删除
        </el-button>
        <el-button v-if="po.status === 'draft'" type="primary" :loading="acting" @click="doAction('提交下单', () => submitPurchaseOrder(poId))">
          提交下单
        </el-button>
        <el-button v-if="po.status === 'ordered'" type="warning" :loading="acting" @click="doAction('标记已付款', () => markPurchaseOrderPaid(poId))">
          快捷标记已付款
        </el-button>
        <el-button
          v-if="['awaiting_ship', 'paid', 'partial_shipped', 'shipped', 'partial_received'].includes(po.status)"
          type="success"
          :loading="acting"
          @click="doAction('完成采购', () => completePurchaseOrder(poId))"
        >
          完成
        </el-button>
        <el-button
          v-if="po.status === 'draft' || po.status === 'ordered'"
          :loading="acting"
          @click="doAction('取消采购单', () => cancelPurchaseOrder(poId))"
        >
          取消
        </el-button>
      </div>
    </div>

    <el-card v-if="po">
      <template #header>
        <div class="header-row">
          <span>{{ po.poNo }}</span>
          <el-tag :type="statusType(po.status)">{{ statusLabel(po.status) }}</el-tag>
        </div>
      </template>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="基本信息" name="overview">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="供应商">{{ po.supplierName }}（{{ po.supplierCode }}）</el-descriptions-item>
            <el-descriptions-item label="付款状态">{{ PAY_STATUS_MAP[po.payStatus] || po.payStatus }}</el-descriptions-item>
            <el-descriptions-item label="订单类型">{{ po.fulfillmentType === 'dropship' ? '代发直邮' : '采购入仓' }}</el-descriptions-item>
            <el-descriptions-item label="关联订单" :span="2">
              <template v-if="refOrders.length">
                <div class="ref-orders">
                  <div
                    v-for="(r, idx) in refOrders"
                    :key="idx"
                    class="ref-order-line"
                    :class="{ 'line-cancelled': r.cancelled }"
                  >
                    <span
                      class="ref-order-no"
                      :class="{ linkable: !r.cancelled && (!!r.soId || !!po.refSoId) }"
                      @click="!r.cancelled && (r.soId || po.refSoId) ? openOrderCore(r.soId || po.refSoId) : undefined"
                    >
                      {{ r.no }}
                    </span>
                    <el-tag v-if="r.cancelled" type="info" size="small" class="cancel-tag">已撤回</el-tag>
                    <el-button
                      v-else-if="canDetachSales"
                      type="danger"
                      link
                      size="small"
                      :loading="acting"
                      class="detach-btn"
                      @click.stop="handleDetachSales(r)"
                    >
                      解绑
                    </el-button>
                  </div>
                </div>
                <div v-if="refOrders.length > 1" class="ref-count">
                  <template v-if="allRefWithdrawn">共 {{ refOrders.length }} 单（均已撤回）</template>
                  <template v-else-if="activeRefOrders.length < refOrders.length">
                    有效 {{ activeRefOrders.length }} / 共 {{ refOrders.length }} 单
                  </template>
                  <template v-else>共 {{ refOrders.length }} 单</template>
                </div>
              </template>
              <el-link v-else-if="po.refSoId" type="primary" @click="openOrderCore(po.refSoId)">
                #{{ po.refSoId }}
              </el-link>
              <span v-else>—</span>
            </el-descriptions-item>
            <el-descriptions-item label="订单总金额">
              ¥{{ Number(po.saleAmount || 0).toFixed(2) }} {{ po.currency }}
            </el-descriptions-item>
            <el-descriptions-item label="采购总额">¥{{ po.totalAmount.toFixed(2) }} {{ po.currency }}</el-descriptions-item>
            <el-descriptions-item label="预计到货">{{ po.expectedArrivalDate || '—' }}</el-descriptions-item>
            <el-descriptions-item label="采购员">{{ po.buyerName || '—' }}</el-descriptions-item>
            <el-descriptions-item label="采购时间">{{ po.orderedAt || '—' }}</el-descriptions-item>
            <el-descriptions-item label="完成时间">{{ po.completedAt || '—' }}</el-descriptions-item>
            <el-descriptions-item label="备注" :span="2">{{ po.remark || '—' }}</el-descriptions-item>
          </el-descriptions>

          <h4 class="section-title">采购明细</h4>
          <el-table :data="itemTreeRows" border stripe :row-class-name="itemRowClass">
            <el-table-column label="图片" width="72" align="center">
              <template #default="{ row }">
                <el-image
                  v-if="row.item.picUrl"
                  :src="row.item.picUrl"
                  :preview-src-list="[row.item.picUrl]"
                  fit="cover"
                  style="width: 40px; height: 40px; border-radius: 4px"
                  preview-teleported
                />
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="销售单" width="140" show-overflow-tooltip>
              <template #default="{ row }">
                <span v-if="row.isSplitChild" class="muted">—</span>
                <template v-else>
                  <span :class="{ 'line-cancelled': row.item.cancelled }">{{ row.item.refOrderNo || '—' }}</span>
                  <el-tag v-if="row.item.cancelled" type="info" size="small" class="cancel-tag">已撤回</el-tag>
                </template>
              </template>
            </el-table-column>
            <el-table-column label="规格" min-width="260" show-overflow-tooltip>
              <template #default="{ row }">
                <div class="spec-cell" :class="{ child: row.isSplitChild, 'line-cancelled': row.item.cancelled }">
                  <span v-if="row.isSplitChild" class="tree-prefix">└ </span>
                  <span>{{
                    row.isSplitChild
                      ? (row.item.skuSpecs || row.item.productName || '规格')
                      : (row.item.skuSpecs || skuMap.get(row.item.skuId)?.specLabel || row.item.productName || '—')
                  }}</span>
                  <el-tag v-if="row.isSplitChild" size="small" type="warning" class="split-tag">拆分</el-tag>
                  <el-tag v-else-if="row.isSplitParent" size="small" type="info" class="split-tag">已拆分</el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="商家编码" width="140">
              <template #default="{ row }">
                <span v-if="row.isSplitChild" class="muted">—</span>
                <span v-else :class="{ 'line-cancelled': row.item.cancelled }">{{ lineSkuCode(row.item) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="supplierSkuCode" label="对方货号" width="120">
              <template #default="{ row }">
                <span v-if="row.isSplitChild" class="muted">—</span>
                <span v-else :class="{ 'line-cancelled': row.item.cancelled }">{{ row.item.supplierSkuCode || '—' }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="qty" label="数量" width="80" align="center">
              <template #default="{ row }">
                <span :class="{ 'line-cancelled': row.item.cancelled }">{{ row.item.qty }}</span>
              </template>
            </el-table-column>
            <el-table-column label="实付金额" width="110" align="right">
              <template #default="{ row }">
                <span v-if="row.isSplitChild" class="muted">—</span>
                <span v-else-if="(row.item.saleAmount || 0) > 0" :class="{ 'line-cancelled': row.item.cancelled }">
                  ¥{{ Number(row.item.saleAmount).toFixed(2) }}
                </span>
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="采购单价" width="120" align="right" class-name="unit-price-col">
              <template #default="{ row }">
                <span v-if="row.isSplitChild" class="muted">—</span>
                <el-input
                  v-else-if="canEditPurchasePrice && !row.item.cancelled"
                  v-model="row.item.unitPrice"
                  size="small"
                  class="unit-price-input"
                  :disabled="savingPrice"
                  @change="onUnitPriceChange(row.item)"
                />
                <span v-else :class="{ 'line-cancelled': row.item.cancelled }">¥{{ Number(row.item.unitPrice || 0).toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="采购小计" width="120" align="right" class-name="unit-price-col">
              <template #default="{ row }">
                <span v-if="row.isSplitChild" class="muted">—</span>
                <el-input
                  v-else-if="canEditPurchasePrice && !row.item.cancelled"
                  v-model="row.item.lineAmount"
                  size="small"
                  class="unit-price-input"
                  :disabled="savingPrice"
                  @change="onLineAmountChange(row.item)"
                />
                <span v-else :class="{ 'line-cancelled': row.item.cancelled }">
                  ¥{{ Number(row.item.lineAmount || 0).toFixed(2) }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="物流" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">
                <template v-if="row.item.id && logisticsByItem.get(row.item.id)?.length">
                  <div v-for="t in logisticsByItem.get(row.item.id)" :key="t">{{ t }}</div>
                </template>
                <span v-else class="muted">未关联</span>
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="商品备注" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">{{ row.item.remark || '—' }}</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="物流" name="shipments" :disabled="!trackable">
          <PoShipmentTab v-if="trackable && po" :po-id="poId" :po="po" :readonly="false" @refresh="loadData" />
          <el-empty v-else description="提交下单后可维护物流" />
        </el-tab-pane>

        <el-tab-pane label="付款" name="payments" :disabled="!trackable">
          <PoPaymentTab v-if="trackable && po" :po-id="poId" :po="po" :readonly="false" @refresh="loadData" />
          <el-empty v-else description="提交下单后可记录付款" />
        </el-tab-pane>

        <el-tab-pane label="附件" name="attachments" :disabled="!trackable">
          <PoAttachmentTab v-if="trackable" :po-id="poId" :readonly="false" />
          <el-empty v-else description="提交下单后可上传附件" />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<style scoped>
.po-detail {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
}
.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.header-row {
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 600;
}
.section-title {
  margin: 20px 0 12px;
  font-size: 15px;
}
.muted {
  color: #c0c4cc;
}
.line-cancelled {
  color: #a8abb2;
  text-decoration: line-through;
}
.cancel-tag {
  margin-left: 6px;
}
:deep(.unit-price-col .cell) {
  overflow: visible;
}
.unit-price-input {
  width: 100%;
}
.unit-price-input :deep(.el-input__inner) {
  text-align: right;
}
:deep(.po-item-cancelled) {
  background-color: #f5f7fa !important;
  color: #a8abb2;
}
:deep(.po-item-cancelled td) {
  background-color: #f5f7fa !important;
}
.ref-orders {
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: 16em;
  overflow-y: auto;
  line-height: 1.4;
}
.ref-order-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px 8px;
  font-size: 13px;
  color: #606266;
  word-break: break-all;
}
.ref-order-no.linkable {
  cursor: pointer;
  color: var(--el-color-primary);
}
.detach-btn {
  flex-shrink: 0;
}
.ref-count {
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}
.spec-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}
.spec-cell.child {
  padding-left: 8px;
  color: #475569;
}
.tree-prefix {
  color: #8f959e;
  flex-shrink: 0;
}
.split-tag {
  flex-shrink: 0;
}
</style>
