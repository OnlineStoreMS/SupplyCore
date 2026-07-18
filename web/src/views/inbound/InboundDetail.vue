<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  approveInboundFinance, approveInboundWH, fetchPurchaseInbound, voidInbound,
  INBOUND_STATUS_MAP, type PurchaseInboundDetail,
} from '../../api/purchaseExt'

const route = useRoute()
const router = useRouter()
const detail = ref<PurchaseInboundDetail | null>(null)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    detail.value = await fetchPurchaseInbound(Number(route.params.id))
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function run(action: () => Promise<unknown>, ok: string) {
  try {
    await action()
    ElMessage.success(ok)
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <div class="toolbar" v-if="detail">
      <div>
        <h2>入库单 {{ detail.inboundNo }}</h2>
        <el-tag :type="(INBOUND_STATUS_MAP[detail.status]?.type as any) || 'info'" size="small">
          {{ INBOUND_STATUS_MAP[detail.status]?.label || detail.status }}
        </el-tag>
      </div>
      <div class="actions">
        <el-button @click="router.push('/purchase-inbounds')">返回</el-button>
        <el-button
          v-if="detail.status === 'draft' || detail.status === 'pending_wh'"
          type="warning"
          @click="run(() => approveInboundWH(detail!.id), '入库审核完成')"
        >
          入库审核
        </el-button>
        <el-button
          v-if="detail.status === 'pending_finance'"
          type="success"
          @click="run(() => approveInboundFinance(detail!.id), '财务审核完成')"
        >
          财务审核
        </el-button>
        <el-button
          v-if="detail.status !== 'completed' && detail.status !== 'void'"
          type="danger"
          plain
          @click="run(() => voidInbound(detail!.id), '已作废')"
        >
          作废
        </el-button>
      </div>
    </div>
    <el-descriptions v-if="detail" :column="3" border class="meta">
      <el-descriptions-item label="采购单">{{ detail.poNo || '-' }}</el-descriptions-item>
      <el-descriptions-item label="供应商">{{ detail.supplierName || '-' }}</el-descriptions-item>
      <el-descriptions-item label="仓库">{{ detail.warehouseName || '-' }}</el-descriptions-item>
      <el-descriptions-item label="物流单号">{{ detail.trackingNo || '-' }}</el-descriptions-item>
      <el-descriptions-item label="总数量">{{ detail.totalQty }}</el-descriptions-item>
      <el-descriptions-item label="总金额">{{ detail.totalAmount }}</el-descriptions-item>
      <el-descriptions-item label="制单人">{{ detail.creatorName }}</el-descriptions-item>
      <el-descriptions-item label="入库审核">{{ detail.whAuditorName || '-' }} {{ detail.whAuditedAt || '' }}</el-descriptions-item>
      <el-descriptions-item label="财务审核">{{ detail.finAuditorName || '-' }} {{ detail.finAuditedAt || '' }}</el-descriptions-item>
      <el-descriptions-item label="备注" :span="3">{{ detail.remark || '-' }}</el-descriptions-item>
    </el-descriptions>
    <el-table v-if="detail" :data="detail.items" stripe class="items">
      <el-table-column prop="skuCode" label="SKU" min-width="120" />
      <el-table-column prop="skuName" label="名称" min-width="140" />
      <el-table-column prop="purchaseQty" label="采购数" width="90" />
      <el-table-column prop="qcQty" label="质检数" width="90" />
      <el-table-column prop="rejectQty" label="不合格" width="90" />
      <el-table-column prop="inboundQty" label="入库数" width="90" />
      <el-table-column prop="unitPrice" label="单价" width="100" />
      <el-table-column prop="lineAmount" label="金额" width="100" />
      <el-table-column prop="locationCode" label="库位" width="100" />
    </el-table>
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; gap: 12px; }
.toolbar h2 { margin: 0 0 8px; font-size: 18px; }
.actions { display: flex; gap: 8px; }
.meta, .items { background: #fff; }
.items { margin-top: 16px; }
</style>
