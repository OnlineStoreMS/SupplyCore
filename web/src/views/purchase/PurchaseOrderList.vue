<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Search, View } from '@element-plus/icons-vue'
import {
  fetchPurchaseOrders,
  PO_STATUS_MAP,
  PAY_STATUS_MAP,
  type PurchaseOrderListItem,
} from '../../api/purchase'
import { fetchSuppliers, type Supplier } from '../../api/supplier'

const route = useRoute()
const router = useRouter()
const tableData = ref<PurchaseOrderListItem[]>([])
const suppliers = ref<Supplier[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const status = ref('')
const supplierId = ref<number | undefined>()
const keyword = ref('')
const refSoId = ref<number | undefined>()
const loading = ref(false)

function syncQueryFilters() {
  const q = route.query.refSoId
  if (q) {
    refSoId.value = Number(q)
  }
}

async function loadSuppliers() {
  try {
    const data = await fetchSuppliers(undefined, 1, 200)
    suppliers.value = data.list
  } catch {
    suppliers.value = []
  }
}

async function loadData() {
  loading.value = true
  try {
    const data = await fetchPurchaseOrders({
      status: status.value || undefined,
      supplierId: supplierId.value,
      refSoId: refSoId.value,
      keyword: keyword.value || undefined,
      page: page.value,
      pageSize: pageSize.value,
    })
    tableData.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  syncQueryFilters()
  await loadSuppliers()
  await loadData()
})

watch(() => route.query.refSoId, () => {
  syncQueryFilters()
  void loadData()
})

function statusLabel(s: string) {
  return PO_STATUS_MAP[s]?.label || s
}

function statusType(s: string) {
  return PO_STATUS_MAP[s]?.type || 'info'
}

function openDetail(row: PurchaseOrderListItem) {
  router.push(`/purchase-orders/${row.id}`)
}

function openCreate() {
  router.push('/purchase-orders/create')
}
</script>

<template>
  <div class="po-page">
    <el-card v-loading="loading">
      <template #header>
        <span>采购单</span>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建采购单</el-button>
      </template>

      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="采购单号"
          :prefix-icon="Search"
          clearable
          style="width: 180px"
          @change="() => { page = 1; loadData() }"
        />
        <el-select
          v-model="status"
          placeholder="状态"
          clearable
          style="width: 130px"
          @change="() => { page = 1; loadData() }"
        >
          <el-option v-for="(v, k) in PO_STATUS_MAP" :key="k" :label="v.label" :value="k" />
        </el-select>
        <el-select
          v-model="supplierId"
          placeholder="供应商"
          clearable
          filterable
          style="width: 200px"
          @change="() => { page = 1; loadData() }"
        >
          <el-option v-for="s in suppliers" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-button :icon="Search" @click="loadData">查询</el-button>
      </div>

      <el-table :data="tableData" stripe border>
        <el-table-column prop="poNo" label="采购单号" width="150">
          <template #default="{ row }">
            <el-link type="primary" @click="openDetail(row)">{{ row.poNo }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="supplierName" label="供应商" min-width="140" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="付款" width="90" align="center">
          <template #default="{ row }">
            {{ PAY_STATUS_MAP[row.payStatus] || row.payStatus }}
          </template>
        </el-table-column>
        <el-table-column label="金额" width="120" align="right">
          <template #default="{ row }">¥{{ row.totalAmount.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="itemCount" label="行数" width="70" align="center" />
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="View" @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="(p: number) => { page = p; loadData() }"
        />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.po-page :deep(.el-card__header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
