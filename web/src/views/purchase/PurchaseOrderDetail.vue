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
  PO_STATUS_MAP,
  PAY_STATUS_MAP,
  type PurchaseOrder,
} from '../../api/purchase'
import { resolveProductSkus, type ProductSkuSearchItem } from '../../api/productSku'
import { fetchShipments, type Shipment } from '../../api/poTracking'
import PoShipmentTab from './PoShipmentTab.vue'
import PoPaymentTab from './PoPaymentTab.vue'
import PoAttachmentTab from './PoAttachmentTab.vue'

const route = useRoute()
const router = useRouter()
const poId = computed(() => Number(route.params.id))
const activeTab = ref('overview')

const loading = ref(false)
const acting = ref(false)
const po = ref<PurchaseOrder | null>(null)
const skuMap = ref<Map<number, ProductSkuSearchItem>>(new Map())
const shipments = ref<Shipment[]>([])

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
</script>

<template>
  <div v-loading="loading" class="po-detail">
    <div class="top-bar">
      <el-button :icon="ArrowLeft" text @click="router.push('/purchase-orders')">返回列表</el-button>
      <div v-if="po" class="actions">
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
          v-if="['paid', 'partial_shipped', 'in_transit', 'partial_received'].includes(po.status)"
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
            <el-descriptions-item label="关联订单">
              <template v-if="po.refTraceId || po.refSoId">
                <span v-if="po.refTraceId">{{ po.refTraceId }}</span>
                <span v-else-if="po.refSoId">#{{ po.refSoId }}</span>
              </template>
              <span v-else>—</span>
            </el-descriptions-item>
            <el-descriptions-item label="订单总金额">
              ¥{{ Number(po.saleAmount || 0).toFixed(2) }} {{ po.currency }}
            </el-descriptions-item>
            <el-descriptions-item label="采购总额">¥{{ po.totalAmount.toFixed(2) }} {{ po.currency }}</el-descriptions-item>
            <el-descriptions-item label="预计到货">{{ po.expectedArrivalDate || '—' }}</el-descriptions-item>
            <el-descriptions-item label="采购员">{{ po.buyerName || '—' }}</el-descriptions-item>
            <el-descriptions-item label="下单时间">{{ po.orderedAt || '—' }}</el-descriptions-item>
            <el-descriptions-item label="完成时间">{{ po.completedAt || '—' }}</el-descriptions-item>
            <el-descriptions-item label="备注" :span="2">{{ po.remark || '—' }}</el-descriptions-item>
          </el-descriptions>

          <h4 class="section-title">采购明细</h4>
          <el-table :data="po.items" border stripe>
            <el-table-column label="图片" width="72" align="center">
              <template #default="{ row }">
                <el-image
                  v-if="row.picUrl"
                  :src="row.picUrl"
                  :preview-src-list="[row.picUrl]"
                  fit="cover"
                  style="width: 40px; height: 40px; border-radius: 4px"
                  preview-teleported
                />
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column prop="productName" label="商品" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.productName || skuMap.get(row.skuId)?.productName || '—' }}
              </template>
            </el-table-column>
            <el-table-column label="规格" width="130" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.skuSpecs || skuMap.get(row.skuId)?.specLabel || '—' }}
              </template>
            </el-table-column>
            <el-table-column label="商家编码" width="140">
              <template #default="{ row }">{{ lineSkuCode(row) }}</template>
            </el-table-column>
            <el-table-column prop="supplierSkuCode" label="对方货号" width="120" />
            <el-table-column prop="qty" label="数量" width="80" align="center" />
            <el-table-column label="实付金额" width="110" align="right">
              <template #default="{ row }">
                <span v-if="row.saleAmount > 0">¥{{ Number(row.saleAmount).toFixed(2) }}</span>
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="采购单价" width="100" align="right">
              <template #default="{ row }">¥{{ row.unitPrice.toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="采购小计" width="110" align="right">
              <template #default="{ row }">¥{{ row.lineAmount.toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="物流" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">
                <template v-if="row.id && logisticsByItem.get(row.id)?.length">
                  <div v-for="t in logisticsByItem.get(row.id)" :key="t">{{ t }}</div>
                </template>
                <span v-else class="muted">未关联</span>
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="商品备注" min-width="140" show-overflow-tooltip />
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
</style>
