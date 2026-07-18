<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchPurchaseSuggestions, type SuggestionItem } from '../../api/purchaseExt'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const list = ref<SuggestionItem[]>([])

const source = computed<'stockout' | 'warning' | 'no_stock'>(() => {
  if (route.path.includes('warning')) return 'warning'
  if (route.path.includes('no-stock')) return 'no_stock'
  return 'stockout'
})

const title = computed(() => ({
  stockout: '缺货采购',
  warning: '预警采购',
  no_stock: '无库存采购',
}[source.value]))

async function load() {
  loading.value = true
  try {
    list.value = await fetchPurchaseSuggestions(source.value)
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

function goCreatePO(row: SuggestionItem) {
  router.push({
    path: '/purchase-orders/create',
    query: {
      supplierId: String(row.supplierId || ''),
      skuId: String(row.skuId),
      qty: String(row.suggestPurchase || 1),
      offerId: String(row.offerId || ''),
    },
  })
}

watch(source, load)
onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <div>
        <h2>{{ title }}</h2>
        <p class="hint">对接 WarehouseCore 库存后生成缺货 / 预警 / 无库存采购建议</p>
      </div>
      <el-button @click="load">刷新</el-button>
    </div>
    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="skuCode" label="库存SKU" min-width="120" />
      <el-table-column prop="skuName" label="配货名称" min-width="140" />
      <el-table-column prop="supplierName" label="默认供应商" min-width="140" />
      <el-table-column prop="salesQty" label="销售数量" width="100" />
      <el-table-column prop="stockoutQty" label="缺货数量" width="100" />
      <el-table-column prop="suggestPurchase" label="建议采购" width="100" />
      <el-table-column prop="unitPrice" label="采购单价" width="100" />
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" :disabled="!row.supplierId" @click="goCreatePO(row)">
            生成采购订单
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-if="!loading && !list.length" description="暂无采购建议" />
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; }
.toolbar h2 { margin: 0 0 4px; font-size: 18px; }
.hint { margin: 0; color: #909399; font-size: 13px; }
</style>
