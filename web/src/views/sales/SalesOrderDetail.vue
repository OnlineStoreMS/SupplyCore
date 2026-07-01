<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Edit, Refresh } from '@element-plus/icons-vue'
import {
  confirmSalesOrder,
  createPurchaseOrdersFromSO,
  evaluateSourcing,
  fetchSalesOrder,
  FULFILLMENT_MODE_MAP,
  SO_STATUS_MAP,
  type SalesOrder,
  type SourcingEvaluateResp,
  type SourcingLinePlan,
} from '../../api/salesOrder'

const route = useRoute()
const router = useRouter()
const soId = computed(() => Number(route.params.id))

const loading = ref(false)
const sourcingLoading = ref(false)
const creatingPO = ref(false)
const so = ref<SalesOrder | null>(null)
const sourcing = ref<SourcingEvaluateResp | null>(null)
const selections = ref<Record<number, number>>({})
const autoSubmit = ref(false)

async function loadSO() {
  loading.value = true
  try {
    so.value = await fetchSalesOrder(soId.value)
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadSO)

function statusLabel(s: string) {
  return SO_STATUS_MAP[s]?.label || s
}

function statusType(s: string) {
  return SO_STATUS_MAP[s]?.type || 'info'
}

function fulfillmentLabel(m: string) {
  return FULFILLMENT_MODE_MAP[m] || m
}

async function doConfirm() {
  try {
    await ElMessageBox.confirm('确认该销售单？', '提示')
    so.value = await confirmSalesOrder(soId.value)
    ElMessage.success('已确认')
  } catch (e) {
    if ((e as string) !== 'cancel') {
      ElMessage.error((e as Error).message || '操作失败')
    }
  }
}

async function loadSourcing() {
  sourcingLoading.value = true
  try {
    sourcing.value = await evaluateSourcing(soId.value)
    const next: Record<number, number> = {}
    for (const line of sourcing.value.lines) {
      if (line.needsPo && line.recommendedOfferId) {
        next[line.soItemId] = line.recommendedOfferId
      }
    }
    selections.value = next
  } catch (e) {
    ElMessage.error((e as Error).message || '寻源评估失败')
  } finally {
    sourcingLoading.value = false
  }
}

function lineNeedsSelection(line: SourcingLinePlan) {
  return line.needsPo && (line.offers?.length ?? 0) > 0
}

async function createPOs() {
  if (!sourcing.value) {
    await loadSourcing()
  }
  const picks = Object.entries(selections.value)
    .filter(([, offerId]) => offerId > 0)
    .map(([soItemId, offerId]) => ({
      soItemId: Number(soItemId),
      offerId,
    }))
  if (!picks.length) {
    ElMessage.warning('请选择供货报价')
    return
  }
  creatingPO.value = true
  try {
    const resp = await createPurchaseOrdersFromSO(soId.value, {
      autoSubmit: autoSubmit.value,
      selections: picks,
    })
    ElMessage.success(`已生成采购单：${resp.poNos.join(', ')}`)
    await loadSO()
    sourcing.value = null
  } catch (e) {
    ElMessage.error((e as Error).message || '生成采购单失败')
  } finally {
    creatingPO.value = false
  }
}

function openPO(poId: number) {
  router.push(`/purchase-orders/${poId}`)
}

function openEdit() {
  router.push(`/sales-orders/${soId.value}/edit`)
}

function openPOListByRef() {
  router.push({ path: '/purchase-orders', query: { refSoId: String(soId.value) } })
}
</script>

<template>
  <div v-loading="loading" class="so-detail-page">
    <el-page-header :icon="ArrowLeft" @back="router.back()">
      <template #content>
        <span v-if="so">{{ so.soNo }}</span>
        <el-tag v-if="so" :type="statusType(so.status)" size="small" class="status-tag">
          {{ statusLabel(so.status) }}
        </el-tag>
      </template>
      <template #extra>
        <el-button v-if="so?.status === 'draft'" :icon="Edit" @click="openEdit">编辑</el-button>
        <el-button v-if="so?.status === 'draft'" type="primary" @click="doConfirm">确认</el-button>
      </template>
    </el-page-header>

    <el-card v-if="so" class="info-card">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="Trace ID">{{ so.traceId }}</el-descriptions-item>
        <el-descriptions-item label="来源">{{ so.sourceChannel || '-' }}</el-descriptions-item>
        <el-descriptions-item label="收货人">{{ so.receiverName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="电话">{{ so.receiverPhone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="地址" :span="2">
          {{ [so.province, so.city, so.district, so.receiverAddress].filter(Boolean).join(' ') || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ so.remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ so.createdAt }}</el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">明细</el-divider>
      <el-table :data="so.items" border size="small">
        <el-table-column prop="skuId" label="SKU ID" width="100" />
        <el-table-column prop="qty" label="数量" width="80" />
        <el-table-column label="履约" width="90">
          <template #default="{ row }">{{ fulfillmentLabel(row.fulfillmentMode) }}</template>
        </el-table-column>
        <el-table-column label="关联采购单" min-width="140">
          <template #default="{ row }">
            <el-button v-if="row.linkedPoId" link type="primary" @click="openPO(row.linkedPoId)">
              {{ row.linkedPoNo || `#${row.linkedPoId}` }}
            </el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" />
      </el-table>

      <div class="sourcing-actions">
        <el-button :icon="Refresh" :loading="sourcingLoading" @click="loadSourcing">评估寻源</el-button>
        <el-button link type="primary" @click="openPOListByRef">查看关联采购单</el-button>
      </div>

      <el-card v-if="sourcing" v-loading="sourcingLoading" class="sourcing-card" shadow="never">
        <template #header>寻源方案</template>
        <el-table :data="sourcing.lines" border size="small">
          <el-table-column prop="skuId" label="SKU" width="90" />
          <el-table-column prop="qty" label="数量" width="70" />
          <el-table-column label="履约" width="80">
            <template #default="{ row }">{{ fulfillmentLabel(row.fulfillmentMode) }}</template>
          </el-table-column>
          <el-table-column label="需采购" width="80">
            <template #default="{ row }">
              <el-tag :type="row.needsPo ? 'warning' : 'info'" size="small">
                {{ row.needsPo ? '是' : '否' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="供货报价" min-width="220">
            <template #default="{ row }">
              <template v-if="lineNeedsSelection(row)">
                <el-select v-model="selections[row.soItemId]" placeholder="选择报价" style="width: 100%">
                  <el-option
                    v-for="o in row.offers"
                    :key="o.offerId"
                    :label="`${o.supplierName || o.supplierId} · ¥${o.supplyPrice}`"
                    :value="o.offerId"
                  />
                </el-select>
              </template>
              <span v-else class="hint">{{ row.message || '-' }}</span>
            </template>
          </el-table-column>
        </el-table>
        <div class="create-po-bar">
          <el-checkbox v-model="autoSubmit">生成后自动提交采购单</el-checkbox>
          <el-button type="primary" :loading="creatingPO" @click="createPOs">生成采购单</el-button>
        </div>
      </el-card>
    </el-card>
  </div>
</template>

<style scoped>
.so-detail-page {
  max-width: 1000px;
}
.status-tag {
  margin-left: 8px;
}
.info-card {
  margin-top: 16px;
}
.sourcing-actions {
  margin-top: 16px;
  display: flex;
  gap: 12px;
  align-items: center;
}
.sourcing-card {
  margin-top: 16px;
}
.create-po-bar {
  margin-top: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
</style>
