<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createInboundFromPackage, fetchPackageReceives, type PackageReceive,
} from '../../api/purchaseExt'

const loading = ref(false)
const list = ref<PackageReceive[]>([])
const total = ref(0)
const page = ref(1)
const keyword = ref('')

async function load() {
  loading.value = true
  try {
    const data = await fetchPackageReceives(keyword.value || undefined, page.value, 20)
    list.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function genInbound(row: PackageReceive) {
  try {
    const inbound = await createInboundFromPackage(row.id)
    ElMessage.success(`已生成入库单 ${inbound.inboundNo}`)
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
      <h2>收货记录</h2>
      <div class="actions">
        <el-input v-model="keyword" clearable placeholder="物流单号/采购单号" style="width: 200px" @keyup.enter="load" />
        <el-button @click="load">查询</el-button>
      </div>
    </div>
    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="trackingNo" label="物流单号" min-width="140" />
      <el-table-column prop="carrier" label="快递公司" width="120" />
      <el-table-column prop="packageType" label="包裹类型" width="100" />
      <el-table-column prop="warehouseName" label="仓库" width="120" />
      <el-table-column prop="poNo" label="采购单号" width="140" />
      <el-table-column prop="inboundNo" label="入库单号" width="140" />
      <el-table-column prop="scannerName" label="扫描人" width="100" />
      <el-table-column prop="createdAt" label="扫描时间" width="170" />
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="!row.inboundId && row.poId"
            link
            type="primary"
            @click="genInbound(row)"
          >
            生成入库单
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
