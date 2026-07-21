<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  HomeFilled, OfficeBuilding, ShoppingCart,
  Box, Van, RefreshLeft, Warning, User,
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const collapsed = defineModel<boolean>('collapsed', { default: false })

const openMenus = computed(() => {
  const p = route.path
  if (p.startsWith('/suggestions')) return ['suggestions']
  if (p.startsWith('/package-receives') || p.startsWith('/scan-inbound')) return ['package']
  if (p.startsWith('/purchase-inbounds') || p.startsWith('/inbound-sort')) return ['inbound']
  if (p.startsWith('/purchase-returns')) return ['returns']
  if (p.startsWith('/suppliers') || p.startsWith('/sku-offers')) return ['supplier']
  if (p.startsWith('/purchase-orders')) return ['supplier-orders']
  return []
})

const logoText = computed(() => (collapsed.value ? 'SC' : 'SupplyCore'))

const activeMenu = computed(() => {
  if (route.path.startsWith('/purchase-orders')) {
    const ft = route.query.fulfillmentType
    if (ft === 'dropship') return '/purchase-orders?fulfillmentType=dropship'
    if (ft === 'stock_in') return '/purchase-orders?fulfillmentType=stock_in'
    return '/purchase-orders'
  }
  return route.path
})

function navigate(path: string) {
  router.push(path)
}
</script>

<template>
  <aside class="sidebar" :class="{ collapsed }">
    <div class="logo">{{ logoText }}</div>
    <el-menu
      :key="activeMenu"
      :default-active="activeMenu"
      :default-openeds="openMenus"
      :collapse="collapsed"
      background-color="#001529"
      text-color="#ffffffa6"
      active-text-color="#fff"
    >
      <el-menu-item index="/dashboard" @click="navigate('/dashboard')">
        <el-icon><HomeFilled /></el-icon>
        <span>工作台</span>
      </el-menu-item>

      <el-sub-menu index="suggestions">
        <template #title>
          <el-icon><Warning /></el-icon>
          <span>采购建议</span>
        </template>
        <el-menu-item index="/suggestions/stockout" @click="navigate('/suggestions/stockout')">缺货采购</el-menu-item>
        <el-menu-item index="/suggestions/warning" @click="navigate('/suggestions/warning')">预警采购</el-menu-item>
        <el-menu-item index="/suggestions/no-stock" @click="navigate('/suggestions/no-stock')">无库存采购</el-menu-item>
      </el-sub-menu>

      <el-menu-item index="/purchase-accounts" @click="navigate('/purchase-accounts')">
        <el-icon><User /></el-icon>
        <span>采购账号</span>
      </el-menu-item>

      <el-sub-menu index="supplier-orders">
        <template #title>
          <el-icon><ShoppingCart /></el-icon>
          <span>供应商订单</span>
        </template>
        <el-menu-item index="/purchase-orders" @click="navigate('/purchase-orders')">全部订单</el-menu-item>
        <el-menu-item
          index="/purchase-orders?fulfillmentType=dropship"
          @click="navigate('/purchase-orders?fulfillmentType=dropship')"
        >
          代发订单
        </el-menu-item>
        <el-menu-item
          index="/purchase-orders?fulfillmentType=stock_in"
          @click="navigate('/purchase-orders?fulfillmentType=stock_in')"
        >
          采购订单
        </el-menu-item>
      </el-sub-menu>

      <el-sub-menu index="package">
        <template #title>
          <el-icon><Van /></el-icon>
          <span>收包入库</span>
        </template>
        <el-menu-item index="/package-receives" @click="navigate('/package-receives')">收货记录</el-menu-item>
        <el-menu-item index="/scan-inbound" @click="navigate('/scan-inbound')">包裹扫描入库</el-menu-item>
      </el-sub-menu>

      <el-sub-menu index="inbound">
        <template #title>
          <el-icon><Box /></el-icon>
          <span>采购入库</span>
        </template>
        <el-menu-item index="/purchase-inbounds" @click="navigate('/purchase-inbounds')">采购入库单</el-menu-item>
        <el-menu-item index="/inbound-sort" @click="navigate('/inbound-sort')">采购入库分拣</el-menu-item>
      </el-sub-menu>

      <el-menu-item index="/purchase-returns" @click="navigate('/purchase-returns')">
        <el-icon><RefreshLeft /></el-icon>
        <span>采购退回单</span>
      </el-menu-item>

      <el-sub-menu index="supplier">
        <template #title>
          <el-icon><OfficeBuilding /></el-icon>
          <span>供应商</span>
        </template>
        <el-menu-item index="/suppliers" @click="navigate('/suppliers')">供应商信息</el-menu-item>
        <el-menu-item index="/sku-offers" @click="navigate('/sku-offers')">SKU 供货报价</el-menu-item>
      </el-sub-menu>
    </el-menu>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 220px;
  background: #001529;
  transition: width 0.2s;
  flex-shrink: 0;
  overflow: auto;
}
.sidebar.collapsed {
  width: 64px;
}
.logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 600;
  font-size: 16px;
  border-bottom: 1px solid #ffffff14;
}
.sidebar :deep(.el-menu) {
  border-right: none;
}
</style>
