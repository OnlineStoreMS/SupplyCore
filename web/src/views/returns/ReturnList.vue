<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  approveReturn, approveReturnFinance, createPurchaseReturn,
  fetchPurchaseReturns, voidReturn, RETURN_STATUS_MAP,
  fetchPurchaseInbound, type PurchaseReturnListItem,
} from '../../api/purchaseExt'

const router = useRouter()
const loading = ref(false)
const list = ref<PurchaseReturnListItem[]>([])
const total = ref(0)
const page = ref(1)
const status = ref('')
const keyword = ref('')
const dialogVisible = ref(false)
const form = reactive({
  inboundId: undefined as number | undefined,
  warehouseName: '',
  trackingNo: '',
  remark: '',
})
const items = ref<{
  inboundItemId: number
  skuId: number
  skuCode: string
  skuName: string
  originalQty: number
  returnQty: number
  unitPrice: number
}[]>([])

async function load() {
  loading.value = true
  try {
    const data = await fetchPurchaseReturns({
      status: status.value || undefined,
      keyword: keyword.value || undefined,
      page: page.value,
      pageSize: 20,
    })
    list.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function openCreate() {
  form.inboundId = undefined
  form.warehouseName = ''
  form.trackingNo = ''
  form.remark = ''
  items.value = []
  dialogVisible.value = true
}

async function loadInboundItems() {
  if (!form.inboundId) return
  try {
    const detail = await fetchPurchaseInbound(form.inboundId)
    form.warehouseName = detail.warehouseName
    items.value = detail.items.map((it) => ({
      inboundItemId: it.id,
      skuId: it.skuId,
      skuCode: it.skuCode,
      skuName: it.skuName,
      originalQty: it.inboundQty,
      returnQty: it.inboundQty,
      unitPrice: it.unitPrice,
    }))
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function submit() {
  if (!form.inboundId) {
    ElMessage.warning('请填写原入库单 ID')
    return
  }
  try {
    const created = await createPurchaseReturn({
      inboundId: form.inboundId,
      warehouseName: form.warehouseName,
      trackingNo: form.trackingNo,
      remark: form.remark,
      items: items.value,
    })
    ElMessage.success('退回单已创建')
    dialogVisible.value = false
    await router.push(`/purchase-returns/${created.id}`)
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function onApprove(row: PurchaseReturnListItem) {
  await ElMessageBox.confirm(`审核退回单 ${row.returnNo}？`, '确认')
  try {
    await approveReturn(row.id)
    ElMessage.success('退回审核完成')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function onApproveFin(row: PurchaseReturnListItem) {
  await ElMessageBox.confirm(`财务审核退回单 ${row.returnNo}？`, '确认')
  try {
    await approveReturnFinance(row.id)
    ElMessage.success('财务审核完成')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function onVoid(row: PurchaseReturnListItem) {
  await ElMessageBox.confirm(`作废退回单 ${row.returnNo}？`, '确认')
  try {
    await voidReturn(row.id)
    ElMessage.success('已作废')
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
      <h2>采购退回单</h2>
      <div class="actions">
        <el-select v-model="status" clearable placeholder="状态" style="width: 140px" @change="load">
          <el-option v-for="(v, k) in RETURN_STATUS_MAP" :key="k" :label="v.label" :value="k" />
        </el-select>
        <el-input v-model="keyword" clearable placeholder="单号" style="width: 160px" @keyup.enter="load" />
        <el-button @click="load">查询</el-button>
        <el-button type="primary" @click="openCreate">新增采购退回单</el-button>
      </div>
    </div>
    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="returnNo" label="退回单号" width="160" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="(RETURN_STATUS_MAP[row.status]?.type as any) || 'info'" size="small">
            {{ RETURN_STATUS_MAP[row.status]?.label || row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="inboundNo" label="原入库单" width="140" />
      <el-table-column prop="supplierName" label="供应商" min-width="140" />
      <el-table-column prop="totalQty" label="退回数量" width="100" />
      <el-table-column prop="totalAmount" label="退回金额" width="100" />
      <el-table-column prop="creatorName" label="制单人" width="100" />
      <el-table-column prop="createdAt" label="制单时间" width="170" />
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push(`/purchase-returns/${row.id}`)">详情</el-button>
          <el-button
            v-if="row.status === 'draft' || row.status === 'pending_return'"
            link
            type="warning"
            @click="onApprove(row)"
          >
            退回审核
          </el-button>
          <el-button
            v-if="row.status === 'pending_finance'"
            link
            type="success"
            @click="onApproveFin(row)"
          >
            财务审核
          </el-button>
          <el-button
            v-if="row.status !== 'completed' && row.status !== 'void'"
            link
            type="danger"
            @click="onVoid(row)"
          >
            作废
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

    <el-dialog v-model="dialogVisible" title="新增采购退回单" width="720px">
      <el-form label-width="100px">
        <el-form-item label="原入库单ID" required>
          <el-input-number v-model="form.inboundId" :min="1" @change="loadInboundItems" />
          <el-button class="ml" @click="loadInboundItems">加载明细</el-button>
        </el-form-item>
        <el-form-item label="仓库">
          <el-input v-model="form.warehouseName" />
        </el-form-item>
        <el-form-item label="物流单号">
          <el-input v-model="form.trackingNo" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" />
        </el-form-item>
      </el-form>
      <el-table :data="items" stripe max-height="280">
        <el-table-column prop="skuCode" label="SKU" min-width="120" />
        <el-table-column prop="originalQty" label="原入库" width="90" />
        <el-table-column label="退回数量" width="140">
          <template #default="{ row }">
            <el-input-number v-model="row.returnQty" :min="0" :max="row.originalQty" size="small" />
          </template>
        </el-table-column>
        <el-table-column prop="unitPrice" label="单价" width="90" />
      </el-table>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.toolbar h2 { margin: 0; font-size: 18px; }
.actions { display: flex; gap: 8px; }
.pager { margin-top: 16px; justify-content: flex-end; }
.ml { margin-left: 8px; }
</style>
