<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  approveInboundFinance, approveInboundWH, fetchPurchaseInbounds, voidInbound,
  INBOUND_STATUS_MAP, type PurchaseInboundListItem,
} from '../../api/purchaseExt'

const router = useRouter()
const loading = ref(false)
const list = ref<PurchaseInboundListItem[]>([])
const total = ref(0)
const page = ref(1)
const status = ref('')
const keyword = ref('')

async function load() {
  loading.value = true
  try {
    const data = await fetchPurchaseInbounds({
      status: status.value || undefined,
      keyword: keyword.value || undefined,
      page: page.value,
      pageSize: 20,
    })
    list.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function onApproveWH(row: PurchaseInboundListItem) {
  await ElMessageBox.confirm(`对入库单 ${row.inboundNo} 执行入库审核？`, '确认')
  try {
    await approveInboundWH(row.id)
    ElMessage.success('入库审核完成')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function onApproveFin(row: PurchaseInboundListItem) {
  await ElMessageBox.confirm(`对入库单 ${row.inboundNo} 执行财务审核？`, '确认')
  try {
    await approveInboundFinance(row.id)
    ElMessage.success('财务审核完成')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function onVoid(row: PurchaseInboundListItem) {
  await ElMessageBox.confirm(`作废入库单 ${row.inboundNo}？`, '确认')
  try {
    await voidInbound(row.id)
    ElMessage.success('已作废')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>采购入库单</h2>
      <div class="actions">
        <el-select v-model="status" clearable placeholder="状态" style="width: 140px" @change="load">
          <el-option v-for="(v, k) in INBOUND_STATUS_MAP" :key="k" :label="v.label" :value="k" />
        </el-select>
        <el-input v-model="keyword" clearable placeholder="单号/物流单号" style="width: 180px" @keyup.enter="load" />
        <el-button @click="load">查询</el-button>
        <el-button type="primary" @click="router.push('/purchase-inbounds/create')">新增采购入库单</el-button>
      </div>
    </div>
    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="inboundNo" label="入库单号" width="160" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="(INBOUND_STATUS_MAP[row.status]?.type as any) || 'info'" size="small">
            {{ INBOUND_STATUS_MAP[row.status]?.label || row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="poNo" label="关联采购单" width="140" />
      <el-table-column prop="supplierName" label="供应商" min-width="140" />
      <el-table-column prop="warehouseName" label="入库仓库" width="120" />
      <el-table-column prop="totalQty" label="总数量" width="90" />
      <el-table-column prop="totalAmount" label="总金额" width="100" />
      <el-table-column prop="trackingNo" label="物流单号" width="140" />
      <el-table-column prop="creatorName" label="制单人" width="100" />
      <el-table-column prop="createdAt" label="制单时间" width="170" />
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push(`/purchase-inbounds/${row.id}`)">详情</el-button>
          <el-button
            v-if="row.status === 'draft' || row.status === 'pending_wh'"
            link
            type="warning"
            @click="onApproveWH(row)"
          >
            入库审核
          </el-button>
          <el-button
            v-if="row.status === 'pending_finance'"
            link
            type="success"
            @click="onApproveFin(row)"
          >
            财务审核
          </el-button>
          <el-button
            v-if="row.status !== 'completed' && row.status !== 'void'"
            link
            type="danger"
            @click="onVoid(row)"
          >
            作废
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="page"
      class="pager"
      layout="total, prev, pager, next"
      :total="total"
      @current-change="load"
    />
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.toolbar h2 { margin: 0; font-size: 18px; }
.actions { display: flex; gap: 8px; }
.pager { margin-top: 16px; justify-content: flex-end; }
</style>
