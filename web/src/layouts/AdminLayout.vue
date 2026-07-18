<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowDown, Expand, Fold } from '@element-plus/icons-vue'
import Sidebar from './Sidebar.vue'
import TenantSwitcher from '../components/TenantSwitcher.vue'
import { portalAppsUrl, portalLoginUrl } from '../utils/auth'
import { useSessionStore } from '../stores/session'

const route = useRoute()
const collapsed = ref(false)
const sessionStore = useSessionStore()

const userInitial = computed(() => {
  const name = sessionStore.session?.user.displayName?.trim()
  return name ? name[0].toUpperCase() : '?'
})

onMounted(() => {
  void sessionStore.load(true)
})

function backToPortal() {
  window.location.href = portalAppsUrl()
}

function logout() {
  sessionStore.clear()
  window.location.href = portalLoginUrl()
}

const breadcrumbs = computed(() => {
  const title = (route.meta.title as string) || 'SupplyCore'
  if (route.path.startsWith('/suggestions')) return ['采购建议', title]
  if (route.path.startsWith('/purchase-accounts')) return ['采购', '采购账号']
  if (route.path.startsWith('/package-receives')) return ['收包入库', '收货记录']
  if (route.path.startsWith('/scan-inbound')) return ['收包入库', '包裹扫描入库']
  if (route.path.startsWith('/purchase-inbounds/create')) return ['采购入库', '新增入库单']
  if (route.path.startsWith('/purchase-inbounds/')) return ['采购入库', '入库单详情']
  if (route.path.startsWith('/purchase-inbounds')) return ['采购入库', '采购入库单']
  if (route.path.startsWith('/inbound-sort')) return ['采购入库', '入库分拣']
  if (route.path.startsWith('/purchase-returns/')) return ['采购退回', '退回单详情']
  if (route.path.startsWith('/purchase-returns')) return ['采购退回', '采购退回单']
  if (route.path.startsWith('/suppliers/') && route.params.id) return ['供应商', '供应商详情']
  if (route.path.startsWith('/suppliers')) return ['供应商', '供应商信息']
  if (route.path.startsWith('/sku-offers')) return ['供应商', 'SKU 供货报价']
  if (route.path.startsWith('/purchase-orders/create')) return ['采购订单', '新建']
  if (route.path.includes('/edit') && route.path.includes('purchase-orders')) return ['采购订单', '编辑']
  if (route.path.startsWith('/purchase-orders/')) return ['采购订单', '详情']
  if (route.path.startsWith('/purchase-orders')) return ['采购订单', '列表']
  return ['首页', title]
})
</script>

<template>
  <div class="admin-layout">
    <Sidebar v-model:collapsed="collapsed" />
    <div class="main-area">
      <header class="header">
        <div class="header-left">
          <el-button :icon="collapsed ? Expand : Fold" text @click="collapsed = !collapsed" />
          <el-breadcrumb separator="/">
            <el-breadcrumb-item v-for="(item, i) in breadcrumbs" :key="i">{{ item }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <TenantSwitcher />
          <el-dropdown trigger="click" @command="(cmd: string) => cmd === 'logout' ? logout() : backToPortal()">
            <div class="user-trigger">
              <el-avatar :size="32" class="user-avatar">{{ userInitial }}</el-avatar>
              <div v-if="sessionStore.session" class="user-meta">
                <span class="user-name">{{ sessionStore.session.user.displayName }}</span>
                <span class="tenant-name">{{ sessionStore.session.tenant.name }}</span>
              </div>
              <span v-else class="user-loading">加载中…</span>
              <el-icon class="user-arrow"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="portal">返回应用中心</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin-layout {
  display: flex;
  height: 100vh;
  background: #f0f2f5;
}
.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.header {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.header-right {
  display: flex;
  align-items: center;
}
.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
}
.user-trigger:hover {
  background: #f5f7fa;
}
.user-meta {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}
.user-name {
  font-size: 14px;
  color: #303133;
}
.tenant-name {
  font-size: 12px;
  color: #909399;
}
.user-loading {
  font-size: 13px;
  color: #909399;
}
.user-avatar {
  background: #409eff;
  color: #fff;
}
.content {
  flex: 1;
  overflow: auto;
  padding: 16px;
}
</style>
