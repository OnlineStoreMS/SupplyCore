<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import type { PurchaseOrder } from '../../api/purchase'
import {
  fetchShipments, createShipment, updateShipmentStatus, deleteShipment,
  SHIPMENT_STATUS_MAP, type Shipment,
} from '../../api/poTracking'

const props = defineProps<{ poId: number; po: PurchaseOrder; readonly: boolean }>()
const emit = defineEmits<{ refresh: [] }>()

const loading = ref(false)
const list = ref<Shipment[]>([])
const dialogVisible = ref(false)
const form = ref({
  carrierName: '',
  trackingNo: '',
  expectedArrivalDate: '',
  receiverName: '',
  receiverPhone: '',
  receiverAddress: '',
  remark: '',
})

async function loadData() {
  loading.value = true
  try {
    list.value = await fetchShipments(props.poId)
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function openCreate() {
  form.value = {
    carrierName: '', trackingNo: '', expectedArrivalDate: '',
    receiverName: '', receiverPhone: '', receiverAddress: '', remark: '',
  }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    await createShipment(props.poId, { ...form.value })
    ElMessage.success('已添加发货批次')
    dialogVisible.value = false
    await loadData()
    emit('refresh')
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function changeStatus(row: Shipment, status: string) {
  try {
    await updateShipmentStatus(props.poId, row.id, status)
    ElMessage.success('状态已更新')
    await loadData()
    emit('refresh')
  } catch (e) {
    ElMessage.error((e as Error).message || '更新失败')
  }
}

async function handleDelete(row: Shipment) {
  try {
    await ElMessageBox.confirm('确定删除此发货批次？', '确认')
  } catch {
    return
  }
  try {
    await deleteShipment(props.poId, row.id)
    ElMessage.success('已删除')
    await loadData()
    emit('refresh')
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}
</script>

<template>
  <div v-loading="loading">
    <div v-if="!readonly" class="toolbar">
      <el-button type="primary" :icon="Plus" @click="openCreate">添加发货</el-button>
    </div>
    <el-table :data="list" border stripe>
      <el-table-column prop="shipmentNo" label="批次号" width="140" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">{{ SHIPMENT_STATUS_MAP[row.status] || row.status }}</template>
      </el-table-column>
      <el-table-column prop="carrierName" label="快递" width="100" />
      <el-table-column prop="trackingNo" label="物流单号" min-width="140" />
      <el-table-column prop="expectedArrivalDate" label="预计到货" width="110" />
      <el-table-column prop="shippedAt" label="发货时间" width="150" />
      <el-table-column v-if="!readonly" label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status === 'pending'" link type="primary" @click="changeStatus(row, 'shipped')">已发货</el-button>
          <el-button v-if="row.status === 'shipped'" link type="primary" @click="changeStatus(row, 'in_transit')">运输中</el-button>
          <el-button v-if="row.status === 'in_transit' || row.status === 'shipped'" link type="success" @click="changeStatus(row, 'delivered')">已签收</el-button>
          <el-button link type="danger" :icon="Delete" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" title="添加发货批次" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="快递公司">
          <el-input v-model="form.carrierName" placeholder="如：顺丰速运" />
        </el-form-item>
        <el-form-item label="物流单号">
          <el-input v-model="form.trackingNo" />
        </el-form-item>
        <el-form-item label="预计到货">
          <el-date-picker v-model="form.expectedArrivalDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="收货人">
          <el-input v-model="form.receiverName" />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="form.receiverPhone" />
        </el-form-item>
        <el-form-item label="收货地址">
          <el-input v-model="form.receiverAddress" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  margin-bottom: 12px;
}
</style>
