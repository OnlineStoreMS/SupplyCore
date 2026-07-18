<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createPurchaseInbound } from '../../api/purchaseExt'
import { fetchPurchaseOrders, fetchPurchaseOrder, type PurchaseOrderListItem } from '../../api/purchase'

const router = useRouter()
const saving = ref(false)
const poList = ref<PurchaseOrderListItem[]>([])
const form = reactive({
  poId: undefined as number | undefined,
  warehouseName: '',
  trackingNo: '',
  platformOrderNo: '',
  remark: '',
})
const items = ref<{
  poItemId: number
  skuId: number
  skuCode: string
  skuName: string
  purchaseQty: number
  inboundQty: number
  unitPrice: number
}[]>([])

async function loadPOs() {
  const data = await fetchPurchaseOrders({ page: 1, pageSize: 100 })
  poList.value = data.list.filter((p) => !['completed', 'cancelled'].includes(p.status))
}

async function onPoChange(poId: number) {
  const detail = await fetchPurchaseOrder(poId)
  form.poId = poId
  items.value = detail.items.map((it) => ({
    poItemId: it.id || 0,
    skuId: it.skuId,
    skuCode: it.supplierSkuCode || String(it.skuId),
    skuName: '',
    purchaseQty: it.qty,
    inboundQty: Math.max(it.qty - (it.receivedQty || 0), 0),
    unitPrice: it.unitPrice,
  })).filter((it) => it.inboundQty > 0)
}

async function submit() {
  if (!form.poId) {
    ElMessage.warning('请选择采购单')
    return
  }
  if (!items.value.length) {
    ElMessage.warning('没有可入库明细')
    return
  }
  saving.value = true
  try {
    const created = await createPurchaseInbound({
      poId: form.poId,
      warehouseName: form.warehouseName,
      trackingNo: form.trackingNo,
      platformOrderNo: form.platformOrderNo,
      remark: form.remark,
      items: items.value,
    })
    ElMessage.success('入库单已创建')
    await router.push(`/purchase-inbounds/${created.id}`)
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    saving.value = false
  }
}

onMounted(loadPOs)
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>新增采购入库单</h2>
      <div>
        <el-button @click="router.back()">返回</el-button>
        <el-button type="primary" :loading="saving" @click="submit">保存</el-button>
      </div>
    </div>
    <el-form label-width="100px" class="form">
      <el-form-item label="采购订单" required>
        <el-select v-model="form.poId" filterable style="width: 360px" @change="onPoChange">
          <el-option v-for="p in poList" :key="p.id" :label="`${p.poNo} · ${p.supplierName}`" :value="p.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="入库仓库">
        <el-input v-model="form.warehouseName" style="width: 360px" />
      </el-form-item>
      <el-form-item label="物流单号">
        <el-input v-model="form.trackingNo" style="width: 360px" />
      </el-form-item>
      <el-form-item label="平台单号">
        <el-input v-model="form.platformOrderNo" style="width: 360px" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" style="width: 360px" />
      </el-form-item>
    </el-form>
    <el-table :data="items" stripe>
      <el-table-column prop="skuCode" label="SKU" min-width="120" />
      <el-table-column prop="purchaseQty" label="采购数" width="90" />
      <el-table-column label="本次入库" width="140">
        <template #default="{ row }">
          <el-input-number v-model="row.inboundQty" :min="0" :max="row.purchaseQty" size="small" />
        </template>
      </el-table-column>
      <el-table-column prop="unitPrice" label="单价" width="100" />
    </el-table>
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.toolbar h2 { margin: 0; font-size: 18px; }
.form { background: #fff; padding: 16px; margin-bottom: 16px; border-radius: 8px; }
</style>
