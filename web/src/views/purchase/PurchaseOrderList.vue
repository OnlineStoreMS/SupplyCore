<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Search, View } from '@element-plus/icons-vue'
import {
  fetchPurchaseOrders,
  PO_STATUS_MAP,
  PAY_STATUS_MAP,
  FULFILLMENT_TYPE_MAP,
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
const fulfillmentType = ref('')
const loading = ref(false)

const activeTab = computed({
  get: () => fulfillmentType.value || 'all',
  set: (v: string) => {
    fulfillmentType.value = v === 'all' ? '' : v
    page.value = 1
    syncRouteQuery()
    void loadData()
  },
})

const pageTitle = computed(() => {
  if (fulfillmentType.value === 'dropship') return '代发订单'
  if (fulfillmentType.value === 'stock_in') return '采购订单'
  return '供应商订单'
})

const createLabel = computed(() => {
  if (fulfillmentType.value === 'dropship') return '新建代发单'
  if (fulfillmentType.value === 'stock_in') return '新建采购单'
  return '新建供应商订单'
})

function syncQueryFilters() {
  const q = route.query
  if (q.refSoId) {
    refSoId.value = Number(q.refSoId)
  } else {
    refSoId.value = undefined
  }
  if (typeof q.status === 'string') {
    status.value = q.status
  }
  if (typeof q.fulfillmentType === 'string') {
    fulfillmentType.value = q.fulfillmentType
  } else {
    fulfillmentType.value = ''
  }
}

function syncRouteQuery() {
  const query: Record<string, string> = {}
  if (fulfillmentType.value) query.fulfillmentType = fulfillmentType.value
  if (status.value) query.status = status.value
  if (refSoId.value) query.refSoId = String(refSoId.value)
  router.replace({ path: '/purchase-orders', query })
}

async function loadSuppliers() {
  try {
    const data = await fetchSuppliers({ page: 1, pageSize: 200 })
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
      fulfillmentType: fulfillmentType.value || undefined,
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

watch(() => [route.query.refSoId, route.query.status, route.query.fulfillmentType], () => {
  syncQueryFilters()
  void loadData()
})

function statusLabel(s: string) {
  return PO_STATUS_MAP[s]?.label || s
}

function statusType(s: string) {
  return PO_STATUS_MAP[s]?.type || 'info'
}

function fulfillmentLabel(t: string) {
  return FULFILLMENT_TYPE_MAP[t] || t || '—'
}

function openDetail(row: PurchaseOrderListItem) {
  router.push(`/purchase-orders/${row.id}`)
}

function openCreate() {
  const query: Record<string, string> = {}
  if (fulfillmentType.value) query.fulfillmentType = fulfillmentType.value
  router.push({ path: '/purchase-orders/create', query })
}

function onFilterChange() {
  page.value = 1
  syncRouteQuery()
  void loadData()
}
</script>

<template>
  <div class="po-page">
    <el-card v-loading="loading">
      <template #header>
        <span>{{ pageTitle }}</span>
        <el-button type="primary" :icon="Plus" @click="openCreate">{{ createLabel }}</el-button>
      </template>

      <el-tabs v-model="activeTab" class="type-tabs">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane label="代发订单" name="dropship" />
        <el-tab-pane label="采购订单" name="stock_in" />
      </el-tabs>

      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="订单号"
          :prefix-icon="Search"
          clearable
          style="width: 180px"
          @change="onFilterChange"
        />
        <el-select
          v-model="status"
          placeholder="状态"
          clearable
          style="width: 130px"
          @change="onFilterChange"
        >
          <el-option v-for="(v, k) in PO_STATUS_MAP" :key="k" :label="v.label" :value="k" />
        </el-select>
        <el-select
          v-model="supplierId"
          placeholder="供应商"
          clearable
          filterable
          style="width: 200px"
          @change="onFilterChange"
        >
          <el-option v-for="s in suppliers" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-button :icon="Search" @click="loadData">查询</el-button>
      </div>

      <el-table :data="tableData" stripe border>
        <el-table-column prop="poNo" label="订单号" width="150">
          <template #default="{ row }">
            <el-link type="primary" @click="openDetail(row)">{{ row.poNo }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="supplierName" label="供应商" min-width="140" />
        <el-table-column label="订单类型" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="row.fulfillmentType === 'dropship' ? 'warning' : ''" size="small">
              {{ fulfillmentLabel(row.fulfillmentType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="关联销售单" width="110" align="center">
          <template #default="{ row }">
            <span v-if="row.refSoId">{{ row.refSoId }}</span>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
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
.type-tabs {
  margin-bottom: 4px;
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
.muted {
  color: #c0c4cc;
}
</style>
