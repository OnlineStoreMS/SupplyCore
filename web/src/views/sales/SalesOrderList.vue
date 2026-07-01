<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Search, View } from '@element-plus/icons-vue'
import { fetchSalesOrders, SO_STATUS_MAP, type SalesOrderListItem } from '../../api/salesOrder'

const router = useRouter()
const tableData = ref<SalesOrderListItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const status = ref('')
const keyword = ref('')
const loading = ref(false)

async function loadData() {
  loading.value = true
  try {
    const data = await fetchSalesOrders({
      status: status.value || undefined,
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

onMounted(loadData)

function statusLabel(s: string) {
  return SO_STATUS_MAP[s]?.label || s
}

function statusType(s: string) {
  return SO_STATUS_MAP[s]?.type || 'info'
}

function openDetail(row: SalesOrderListItem) {
  router.push(`/sales-orders/${row.id}`)
}

function openCreate() {
  router.push('/sales-orders/create')
}
</script>

<template>
  <div class="so-page">
    <el-card v-loading="loading">
      <template #header>
        <span>销售订单</span>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建销售单</el-button>
      </template>

      <el-form inline class="filter-form" @submit.prevent="loadData">
        <el-form-item label="状态">
          <el-select v-model="status" clearable placeholder="全部" style="width: 120px" @change="loadData">
            <el-option v-for="(v, k) in SO_STATUS_MAP" :key="k" :label="v.label" :value="k" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="keyword" placeholder="单号 / traceId" clearable @keyup.enter="loadData" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="loadData">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" stripe>
        <el-table-column prop="soNo" label="销售单号" min-width="140" />
        <el-table-column prop="traceId" label="Trace ID" min-width="160" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sourceChannel" label="来源" width="100" />
        <el-table-column prop="itemCount" label="行数" width="70" align="center" />
        <el-table-column prop="createdAt" label="创建时间" width="170" />
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" :icon="View" @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadData"
          @size-change="loadData"
        />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.so-page :deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.filter-form {
  margin-bottom: 16px;
}
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
