<script setup lang="ts">
import { OfficeBuilding, PriceTag, ShoppingCart } from '@element-plus/icons-vue'
</script>

<template>
  <div class="dashboard">
    <el-row :gutter="16">
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-card">
            <el-icon :size="32" color="#409eff"><OfficeBuilding /></el-icon>
            <div>
              <div class="stat-title">供应商管理 (VMS)</div>
              <div class="stat-desc">维护供应商档案、发货地址与 SKU 供货报价</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-card">
            <el-icon :size="32" color="#67c23a"><PriceTag /></el-icon>
            <div>
              <div class="stat-title">供货比价</div>
              <div class="stat-desc">按 SKU 查看各供应商拿货价、发货地与代发能力</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-card">
            <el-icon :size="32" color="#e6a23c"><ShoppingCart /></el-icon>
            <div>
              <div class="stat-title">采购单 (PMS)</div>
              <div class="stat-desc">草稿 → 已下单 → 已付款 → 完成；可从供货报价一键带入明细</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.stat-card {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}
.stat-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 8px;
}
.stat-desc {
  font-size: 13px;
  color: #909399;
  line-height: 1.5;
}
.muted {
  opacity: 0.85;
}
</style>
