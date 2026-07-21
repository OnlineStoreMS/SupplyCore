<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Document, Money, Van, Box, OfficeBuilding, PriceTag,
  ShoppingCart, TrendCharts, Coin, Calendar,
} from '@element-plus/icons-vue'
import { fetchDashboardStats, type DashboardStats } from '../api/dashboard'
import { PO_STATUS_MAP, PAY_STATUS_MAP } from '../api/purchase'

const router = useRouter()
const loading = ref(false)
const stats = ref<DashboardStats | null>(null)

const emptyWorkbench = {
  draftPO: 0, orderedPO: 0, unpaidPO: 0,
  inTransitPO: 0, partialReceivedPO: 0, activeOffers: 0,
}
const emptySupplier = { total: 0, active: 0, offerCount: 0, orderedThisMonth: 0 }
const emptyPO = {
  total: 0, draft: 0, inProgress: 0, completed: 0, cancelled: 0,
  todayCount: 0, weekCount: 0, monthCount: 0,
}
const emptyCost = { todayAmount: 0, weekAmount: 0, monthAmount: 0, unpaidAmount: 0, yearAmount: 0 }

const wb = computed(() => stats.value?.workbench ?? emptyWorkbench)
const supplier = computed(() => stats.value?.supplier ?? emptySupplier)
const po = computed(() => stats.value?.purchaseOrder ?? emptyPO)
const cost = computed(() => stats.value?.cost ?? emptyCost)

const workCards = computed(() => [
  {
    key: 'draft',
    label: '草稿待提交',
    tip: '供应商订单草稿',
    value: wb.value.draftPO,
    color: '#909399',
    icon: Document,
    go: () => router.push({ path: '/purchase-orders', query: { status: 'draft' } }),
  },
  {
    key: 'ordered',
    label: '已下单',
    tip: '待推进付款/发货',
    value: wb.value.orderedPO,
    color: '#409eff',
    icon: ShoppingCart,
    go: () => router.push({ path: '/purchase-orders', query: { status: 'ordered' } }),
  },
  {
    key: 'unpaid',
    label: '待付款',
    tip: '未付/部分付款',
    value: wb.value.unpaidPO,
    color: '#e6a23c',
    icon: Money,
    go: () => router.push({ path: '/purchase-orders' }),
  },
  {
    key: 'transit',
    label: '在途采购',
    tip: '部分发货/运输中',
    value: wb.value.inTransitPO,
    color: '#0f766e',
    icon: Van,
    go: () => router.push({ path: '/purchase-orders', query: { status: 'in_transit' } }),
  },
  {
    key: 'partial',
    label: '部分到货',
    tip: '待继续收货',
    value: wb.value.partialReceivedPO,
    color: '#722ed1',
    icon: Box,
    go: () => router.push({ path: '/purchase-orders', query: { status: 'partial_received' } }),
  },
  {
    key: 'offers',
    label: '有效报价',
    tip: 'SKU 供货报价',
    value: wb.value.activeOffers,
    color: '#67c23a',
    icon: PriceTag,
    go: () => router.push('/sku-offers'),
  },
])

const supplierCards = computed(() => [
  { label: '供应商总数', value: supplier.value.total, color: '#409eff', icon: OfficeBuilding },
  { label: '启用中', value: supplier.value.active, color: '#67c23a', icon: OfficeBuilding },
  { label: '本月有采购', value: supplier.value.orderedThisMonth, color: '#e6a23c', icon: Calendar },
  { label: '供货报价数', value: supplier.value.offerCount, color: '#909399', icon: PriceTag },
])

const poCards = computed(() => [
  { label: '订单总数', value: po.value.total, sub: `进行中 ${po.value.inProgress}` },
  { label: '今日新单', value: po.value.todayCount, sub: `近7日 ${po.value.weekCount}` },
  { label: '本月订单', value: po.value.monthCount, sub: `已完成 ${po.value.completed}` },
  { label: '草稿 / 取消', value: po.value.draft, sub: `已取消 ${po.value.cancelled}` },
])

const costCards = computed(() => [
  { label: '今日采购额', value: cost.value.todayAmount, color: '#1677ff' },
  { label: '近7日采购额', value: cost.value.weekAmount, color: '#13c2c2' },
  { label: '本月采购额', value: cost.value.monthAmount, color: '#0f766e' },
  { label: '待付金额', value: cost.value.unpaidAmount, color: '#e6a23c' },
])

async function load() {
  loading.value = true
  try {
    stats.value = await fetchDashboardStats()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)

function fmtMoney(v?: number) {
  const n = Number(v || 0)
  return n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function statusLabel(s: string) {
  return PO_STATUS_MAP[s]?.label || s
}

function statusType(s: string) {
  return PO_STATUS_MAP[s]?.type || 'info'
}

function goPO(id: number) {
  router.push(`/purchase-orders/${id}`)
}

function goSupplier(id: number) {
  router.push(`/suppliers/${id}`)
}
</script>

<template>
  <div v-loading="loading" class="dashboard">
    <section>
      <div class="section-title">工作场景</div>
      <el-row :gutter="16">
        <el-col v-for="card in workCards" :key="card.key" :xs="12" :sm="8" :lg="4">
          <el-card shadow="hover" class="work-card" @click="card.go()">
            <div class="work-inner">
              <div>
                <div class="work-label">{{ card.label }}</div>
                <div class="work-value" :style="{ color: card.color }">{{ card.value }}</div>
                <div class="work-tip">{{ card.tip }}</div>
              </div>
              <div class="work-icon" :style="{ background: card.color + '18', color: card.color }">
                <el-icon :size="22"><component :is="card.icon" /></el-icon>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </section>

    <section>
      <div class="section-title">采购成本统计</div>
      <el-row :gutter="16">
        <el-col v-for="card in costCards" :key="card.label" :xs="12" :sm="12" :lg="6">
          <el-card shadow="hover" class="metric-card">
            <div class="metric-label">
              <el-icon :style="{ color: card.color }"><Coin /></el-icon>
              {{ card.label }}
            </div>
            <div class="metric-value">¥{{ fmtMoney(card.value) }}</div>
            <div v-if="card.label === '本月采购额'" class="metric-sub">
              本年累计 ¥{{ fmtMoney(cost.yearAmount) }}
            </div>
          </el-card>
        </el-col>
      </el-row>
    </section>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <span>供应商订单统计</span>
            <el-button type="primary" link @click="router.push('/purchase-orders')">查看全部</el-button>
          </template>
          <el-row :gutter="12" class="po-metrics">
            <el-col v-for="card in poCards" :key="card.label" :span="12">
              <div class="po-metric">
                <div class="po-metric-label">{{ card.label }}</div>
                <div class="po-metric-value">{{ card.value }}</div>
                <div class="po-metric-sub">{{ card.sub }}</div>
              </div>
            </el-col>
          </el-row>
          <el-table
            v-if="stats?.statusBreakdown?.length"
            :data="stats.statusBreakdown"
            size="small"
            class="status-table"
          >
            <el-table-column label="状态" min-width="120">
              <template #default="{ row }">
                <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="count" label="单数" width="80" align="right" />
          </el-table>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <span>供应商统计</span>
            <el-button type="primary" link @click="router.push('/suppliers')">查看全部</el-button>
          </template>
          <el-row :gutter="12" class="supplier-metrics">
            <el-col v-for="card in supplierCards" :key="card.label" :xs="12" :span="12">
              <div class="supplier-metric">
                <div class="supplier-metric-top">
                  <span>{{ card.label }}</span>
                  <el-icon :style="{ color: card.color }"><component :is="card.icon" /></el-icon>
                </div>
                <div class="supplier-metric-value">{{ card.value }}</div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <span>本月供应商采购排行</span>
            <el-icon class="header-icon"><TrendCharts /></el-icon>
          </template>
          <el-table :data="stats?.topSuppliers || []" stripe empty-text="本月暂无采购数据">
            <el-table-column label="供应商" min-width="140">
              <template #default="{ row }">
                <el-button link type="primary" @click="goSupplier(row.supplierId)">
                  {{ row.supplierName || `#${row.supplierId}` }}
                </el-button>
              </template>
            </el-table-column>
            <el-table-column prop="orderCount" label="单数" width="80" align="center" />
            <el-table-column label="采购额" width="130" align="right">
              <template #default="{ row }">¥{{ fmtMoney(row.totalAmount) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <span>最近供应商订单</span>
            <el-button type="primary" link @click="router.push('/purchase-orders')">查看全部</el-button>
          </template>
          <el-table :data="stats?.recentOrders || []" stripe empty-text="暂无订单">
            <el-table-column label="单号" min-width="140">
              <template #default="{ row }">
                <el-button link type="primary" @click="goPO(row.id)">{{ row.poNo }}</el-button>
              </template>
            </el-table-column>
            <el-table-column prop="supplierName" label="供应商" min-width="100" />
            <el-table-column label="类型" width="90">
              <template #default="{ row }">
                {{ row.fulfillmentType === 'dropship' ? '代发' : '采购' }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="付款" width="80">
              <template #default="{ row }">{{ PAY_STATUS_MAP[row.payStatus] || row.payStatus }}</template>
            </el-table-column>
            <el-table-column label="金额" width="110" align="right">
              <template #default="{ row }">¥{{ fmtMoney(row.totalAmount) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
}
.work-card {
  cursor: pointer;
  margin-bottom: 12px;
}
.work-card :deep(.el-card__body) {
  padding: 16px;
}
.work-inner {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
}
.work-label {
  font-size: 13px;
  color: #606266;
  margin-bottom: 6px;
}
.work-value {
  font-size: 26px;
  font-weight: 700;
  line-height: 1.2;
}
.work-tip {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}
.work-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.metric-card :deep(.el-card__body) {
  padding: 18px 20px;
}
.metric-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #909399;
  margin-bottom: 8px;
}
.metric-value {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
}
.metric-sub {
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}
.po-metrics,
.supplier-metrics {
  margin-bottom: 8px;
}
.po-metric,
.supplier-metric {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 12px;
}
.po-metric-label,
.supplier-metric-top {
  font-size: 13px;
  color: #909399;
}
.supplier-metric-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.po-metric-value,
.supplier-metric-value {
  font-size: 22px;
  font-weight: 700;
  color: #303133;
  margin-top: 4px;
}
.po-metric-sub {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}
.status-table {
  margin-top: 4px;
}
:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-icon {
  color: #909399;
}
</style>
