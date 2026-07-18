<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  approveReturn, approveReturnFinance, fetchPurchaseReturn, voidReturn,
  RETURN_STATUS_MAP, type PurchaseReturnDetail,
} from '../../api/purchaseExt'

const route = useRoute()
const router = useRouter()
const detail = ref<PurchaseReturnDetail | null>(null)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    detail.value = await fetchPurchaseReturn(Number(route.params.id))
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
        <h2>退回单 {{ detail.returnNo }}</h2>
        <el-tag :type="(RETURN_STATUS_MAP[detail.status]?.type as any) || 'info'" size="small">
          {{ RETURN_STATUS_MAP[detail.status]?.label || detail.status }}
        </el-tag>
      </div>
      <div class="actions">
        <el-button @click="router.push('/purchase-returns')">返回</el-button>
        <el-button
          v-if="detail.status === 'draft' || detail.status === 'pending_return'"
          type="warning"
          @click="run(() => approveReturn(detail!.id), '退回审核完成')"
        >
          退回审核
        </el-button>
        <el-button
          v-if="detail.status === 'pending_finance'"
          type="success"
          @click="run(() => approveReturnFinance(detail!.id), '财务审核完成')"
        >
          财务审核
        </el-button>
        <el-button
          v-if="detail.status !== 'completed' && detail.status !== 'void'"
          type="danger"
          plain
          @click="run(() => voidReturn(detail!.id), '已作废')"
        >
          作废
        </el-button>
      </div>
    </div>
    <el-descriptions v-if="detail" :column="3" border>
      <el-descriptions-item label="原入库单">{{ detail.inboundNo || '-' }}</el-descriptions-item>
      <el-descriptions-item label="供应商">{{ detail.supplierName || '-' }}</el-descriptions-item>
      <el-descriptions-item label="仓库">{{ detail.warehouseName || '-' }}</el-descriptions-item>
      <el-descriptions-item label="退回数量">{{ detail.totalQty }}</el-descriptions-item>
      <el-descriptions-item label="退回金额">{{ detail.totalAmount }}</el-descriptions-item>
      <el-descriptions-item label="实际金额">{{ detail.actualAmount }}</el-descriptions-item>
      <el-descriptions-item label="备注" :span="3">{{ detail.remark || '-' }}</el-descriptions-item>
    </el-descriptions>
    <el-table v-if="detail" :data="detail.items" stripe class="items">
      <el-table-column prop="skuCode" label="SKU" min-width="120" />
      <el-table-column prop="skuName" label="名称" min-width="140" />
      <el-table-column prop="originalQty" label="原入库" width="90" />
      <el-table-column prop="returnQty" label="退回数" width="90" />
      <el-table-column prop="unitPrice" label="单价" width="100" />
      <el-table-column prop="returnAmount" label="退回金额" width="100" />
      <el-table-column prop="actualAmount" label="实际金额" width="100" />
    </el-table>
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; }
.toolbar h2 { margin: 0 0 8px; font-size: 18px; }
.actions { display: flex; gap: 8px; }
.items { margin-top: 16px; background: #fff; }
</style>
