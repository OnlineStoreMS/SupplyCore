<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  OfficeBuilding, PriceTag, TrendCharts, Calendar,
} from '@element-plus/icons-vue'
import { fetchDashboardStats, type DashboardStats } from '../api/dashboard'
import { PO_STATUS_MAP, PAY_STATUS_MAP } from '../api/purchase'
import { goPurchaseOrders } from '../utils/poListIntent'

const router = useRouter()
const loading = ref(false)
const stats = ref<DashboardStats | null>(null)

const emptyWorkbench = {
  dropshipPO: 0, stockInPO: 0,
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
    key: 'dropship',
    label: '代发订单',
    tip: '今日代发 · 点击查看',
    value: wb.value.dropshipPO,
    color: '#d48806',
    highlight: true,
    go: () => goPurchaseOrders(router, {
      fulfillmentType: 'dropship',
      today: true,
      excludeStatuses: ['draft', 'cancelled'],
    }),
  },
  {
    key: 'stock_in',
    label: '采购订单',
    tip: '今日采购入仓',
    value: wb.value.stockInPO,
    color: '#1677ff',
    highlight: false,
    go: () => goPurchaseOrders(router, {
      fulfillmentType: 'stock_in',
      today: true,
      excludeStatuses: ['draft', 'cancelled'],
    }),
  },
  {
    key: 'draft',
    label: '草稿待提交',
    tip: '今日草稿',
    value: wb.value.draftPO,
    color: '#64748b',
    highlight: false,
    go: () => goPurchaseOrders(router, { status: 'draft', today: true }),
  },
  {
    key: 'ordered',
    label: '已下单',
    tip: '今日已下单待推进',
    value: wb.value.orderedPO,
    color: '#409eff',
    highlight: false,
    go: () => goPurchaseOrders(router, { status: 'ordered', today: true }),
  },
  {
    key: 'unpaid',
    label: '待付款',
    tip: '今日未付 / 部分付款',
    value: wb.value.unpaidPO,
    color: '#e6a23c',
    highlight: true,
    go: () => goPurchaseOrders(router, {
      today: true,
      payStatuses: ['unpaid', 'partial'],
      excludeStatuses: ['draft', 'cancelled'],
    }),
  },
  {
    key: 'transit',
    label: '在途采购',
    tip: '今日部分发货 / 运输中',
    value: wb.value.inTransitPO,
    color: '#0f766e',
    highlight: false,
    go: () => goPurchaseOrders(router, {
      today: true,
      statuses: ['partial_shipped', 'in_transit'],
    }),
  },
  {
    key: 'partial',
    label: '部分到货',
    tip: '今日待继续收货',
    value: wb.value.partialReceivedPO,
    color: '#722ed1',
    highlight: false,
    go: () => goPurchaseOrders(router, { status: 'partial_received', today: true }),
  },
  {
    key: 'offers',
    label: '有效报价',
    tip: 'SKU 供货报价',
    value: wb.value.activeOffers,
    color: '#67c23a',
    highlight: false,
    go: () => router.push('/sku-offers'),
  },
])

const costCards = computed(() => [
  {
    label: '今日采购额',
    value: cost.value.todayAmount,
    tip: '今日业务日采购金额',
    highlight: true,
  },
  {
    label: '近7日采购额',
    value: cost.value.weekAmount,
    tip: '近 7 日累计',
    highlight: false,
  },
  {
    label: '本月采购额',
    value: cost.value.monthAmount,
    tip: `本年累计 ¥${fmtMoney(cost.value.yearAmount)}`,
    highlight: false,
  },
  {
    label: '待付金额',
    value: cost.value.unpaidAmount,
    tip: '未付 / 部分付款合计',
    highlight: false,
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
    <div class="section-head">工作场景 · 今日</div>
    <div class="work-cards">
      <button
        v-for="card in workCards"
        :key="card.key"
        type="button"
        class="work-card"
        :class="{ highlight: card.highlight && card.value > 0 }"
        :style="{ '--accent': card.color }"
        @click="card.go()"
      >
        <div class="work-label">{{ card.label }}</div>
        <div class="work-value">{{ card.value }}</div>
        <div class="work-tip">{{ card.tip }}</div>
      </button>
    </div>

    <div class="section-head">采购成本统计</div>
    <div class="metric-row">
      <div
        v-for="card in costCards"
        :key="card.label"
        class="metric-card"
        :class="{ highlight: card.highlight }"
      >
        <div class="metric-label">{{ card.label }}</div>
        <div class="metric-value">¥{{ fmtMoney(card.value) }}</div>
        <div class="metric-tip">{{ card.tip }}</div>
      </div>
    </div>

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
  gap: 14px;
}
.section-head {
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
  margin-top: 2px;
}

.work-cards {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}
.work-card {
  text-align: left;
  border: 1px solid #e8edf3;
  background: #fff;
  border-radius: 10px;
  padding: 14px 16px;
  cursor: pointer;
  border-top: 3px solid var(--accent, #1677ff);
  transition: box-shadow 0.15s, border-color 0.15s, transform 0.15s;
}
.work-card:hover {
  box-shadow: 0 4px 14px rgba(15, 39, 68, 0.08);
  transform: translateY(-1px);
}
.work-card.highlight {
  border-color: color-mix(in srgb, var(--accent) 35%, #e8edf3);
  background: linear-gradient(180deg, color-mix(in srgb, var(--accent) 10%, #fff) 0%, #fff 65%);
}
.work-label {
  font-size: 13px;
  color: #64748b;
}
.work-value {
  margin-top: 6px;
  font-size: 28px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.1;
}
.work-tip {
  margin-top: 6px;
  font-size: 12px;
  color: #94a3b8;
}

.metric-row {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}
.metric-card {
  text-align: left;
  background: #fff;
  border: 1px solid #eef0f3;
  border-radius: 10px;
  padding: 14px 16px;
  transition: box-shadow 0.15s;
}
.metric-card.highlight {
  border-color: #99f6e4;
  background: linear-gradient(180deg, #f0fdfa 0%, #fff 60%);
}
.metric-label {
  font-size: 13px;
  color: #64748b;
}
.metric-value {
  margin-top: 4px;
  font-size: 24px;
  font-weight: 700;
  color: #0f172a;
}
.metric-tip {
  margin-top: 4px;
  font-size: 12px;
  color: #94a3b8;
}

.po-metrics,
.supplier-metrics {
  margin-bottom: 8px;
}
.po-metric,
.supplier-metric {
  background: #f8fafc;
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 12px;
  border: 1px solid #eef0f3;
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
  color: #0f172a;
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

@media (max-width: 1100px) {
  .work-cards,
  .metric-row {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 640px) {
  .work-cards,
  .metric-row {
    grid-template-columns: 1fr;
  }
}
</style>
