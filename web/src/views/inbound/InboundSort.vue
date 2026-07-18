<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const skuCode = ref('')
const scanQty = ref(1)
const scanned = ref<{ skuCode: string; scanQty: number; time: string }[]>([])

function onScan() {
  if (!skuCode.value.trim()) {
    ElMessage.warning('请输入/扫描 SKU')
    return
  }
  scanned.value.unshift({
    skuCode: skuCode.value.trim(),
    scanQty: scanQty.value,
    time: new Date().toLocaleString(),
  })
  skuCode.value = ''
  ElMessage.success('已记录扫描（分拣结果将对接订单核单）')
}
</script>

<template>
  <div>
    <div class="toolbar">
      <div>
        <h2>采购入库分拣</h2>
        <p class="hint">扫描入库 SKU，统计可发订单数量（对齐普源「采购入库分拣」作业台）</p>
      </div>
    </div>
    <div class="panel">
      <el-form inline>
        <el-form-item label="物品SKU">
          <el-input v-model="skuCode" placeholder="扫描或输入 SKU" style="width: 220px" @keyup.enter="onScan" />
        </el-form-item>
        <el-form-item label="扫描数量">
          <el-input-number v-model="scanQty" :min="1" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onScan">扫描</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="scanned" stripe>
        <el-table-column prop="skuCode" label="库存SKU" min-width="140" />
        <el-table-column prop="scanQty" label="扫描数量" width="120" />
        <el-table-column prop="time" label="时间" width="180" />
      </el-table>
      <el-empty v-if="!scanned.length" description="暂无扫描记录，请先扫描 SKU" />
    </div>
  </div>
</template>

<style scoped>
.toolbar { margin-bottom: 16px; }
.toolbar h2 { margin: 0 0 4px; font-size: 18px; }
.hint { margin: 0; color: #909399; font-size: 13px; }
.panel { background: #fff; padding: 16px; border-radius: 8px; }
</style>
