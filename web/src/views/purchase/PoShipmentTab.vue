<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import type { PurchaseOrder, PurchaseOrderItem } from '../../api/purchase'
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

interface LinePick {
  poItemId: number
  productName: string
  skuCode: string
  skuSpecs: string
  picUrl?: string
  qty: number
  shippedQty: number
  remaining: number
  selected: boolean
  shipQty: number
}

const linePicks = ref<LinePick[]>([])

const itemLabelMap = computed(() => {
  const map = new Map<number, PurchaseOrderItem>()
  for (const it of props.po.items || []) {
    if (it.id) map.set(it.id, it)
  }
  return map
})

function shippedQtyByItem(): Map<number, number> {
  const map = new Map<number, number>()
  for (const sh of list.value) {
    for (const it of sh.items || []) {
      map.set(it.poItemId, (map.get(it.poItemId) || 0) + it.qty)
    }
  }
  return map
}

function rebuildLinePicks() {
  const shipped = shippedQtyByItem()
  linePicks.value = (props.po.items || [])
    .filter((it) => it.id)
    .map((it) => {
      const shippedQty = shipped.get(it.id!) || 0
      const remaining = Math.max(0, it.qty - shippedQty)
      return {
        poItemId: it.id!,
        productName: it.productName || '—',
        skuCode: it.skuCode || '',
        skuSpecs: it.skuSpecs || '',
        picUrl: it.picUrl,
        qty: it.qty,
        shippedQty,
        remaining,
        selected: remaining > 0,
        shipQty: remaining > 0 ? remaining : 1,
      }
    })
}

function shipmentProductText(row: Shipment) {
  const items = row.items || []
  if (!items.length) return '—'
  return items
    .map((it) => {
      const poItem = itemLabelMap.value.get(it.poItemId)
      const name = poItem?.productName || `明细#${it.poItemId}`
      const specs = poItem?.skuSpecs ? `（${poItem.skuSpecs}）` : ''
      return `${name}${specs} ×${it.qty}`
    })
    .join('；')
}

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
watch(() => props.po.items, () => {
  if (dialogVisible.value) rebuildLinePicks()
})

function openCreate() {
  form.value = {
    carrierName: '', trackingNo: '', expectedArrivalDate: '',
    receiverName: '', receiverPhone: '', receiverAddress: '', remark: '',
  }
  rebuildLinePicks()
  dialogVisible.value = true
}

async function handleSave() {
  const selected = linePicks.value.filter((l) => l.selected && l.remaining > 0)
  if (!selected.length) {
    ElMessage.warning('请选择本物流单对应的商品')
    return
  }
  for (const line of selected) {
    if (!line.shipQty || line.shipQty < 1 || line.shipQty > line.remaining) {
      ElMessage.warning(`「${line.productName}」发货数量需在 1 ~ ${line.remaining} 之间`)
      return
    }
  }
  if (!form.value.trackingNo?.trim()) {
    ElMessage.warning('请填写物流单号')
    return
  }
  try {
    await createShipment(props.poId, {
      ...form.value,
      items: selected.map((l) => ({ poItemId: l.poItemId, qty: l.shipQty })),
    })
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
      <span class="hint">每个物流单号需勾选对应商品，一单可含多件</span>
    </div>
    <el-table :data="list" border stripe>
      <el-table-column prop="shipmentNo" label="批次号" width="140" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">{{ SHIPMENT_STATUS_MAP[row.status] || row.status }}</template>
      </el-table-column>
      <el-table-column prop="carrierName" label="快递" width="100" />
      <el-table-column prop="trackingNo" label="物流单号" min-width="140" />
      <el-table-column label="对应商品" min-width="220" show-overflow-tooltip>
        <template #default="{ row }">{{ shipmentProductText(row) }}</template>
      </el-table-column>
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

    <el-dialog v-model="dialogVisible" title="添加发货（按商品）" width="720px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="快递公司" required>
          <el-input v-model="form.carrierName" placeholder="如：顺丰速运" />
        </el-form-item>
        <el-form-item label="物流单号" required>
          <el-input v-model="form.trackingNo" placeholder="与下方勾选商品对应" />
        </el-form-item>
        <el-form-item label="发货商品" required>
          <el-table :data="linePicks" border size="small" max-height="280">
            <el-table-column width="48" align="center">
              <template #default="{ row }">
                <el-checkbox v-model="row.selected" :disabled="row.remaining <= 0" />
              </template>
            </el-table-column>
            <el-table-column label="图片" width="56" align="center">
              <template #default="{ row }">
                <el-image
                  v-if="row.picUrl"
                  :src="row.picUrl"
                  fit="cover"
                  style="width: 32px; height: 32px; border-radius: 4px"
                />
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="商品" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">
                <div>{{ row.productName }}</div>
                <div v-if="row.skuSpecs" class="sub">{{ row.skuSpecs }}</div>
              </template>
            </el-table-column>
            <el-table-column label="商家编码" width="110" show-overflow-tooltip>
              <template #default="{ row }">{{ row.skuCode || '—' }}</template>
            </el-table-column>
            <el-table-column label="可发" width="70" align="center">
              <template #default="{ row }">
                <span :class="{ muted: row.remaining <= 0 }">{{ row.remaining }}/{{ row.qty }}</span>
              </template>
            </el-table-column>
            <el-table-column label="本单数量" width="110" align="center">
              <template #default="{ row }">
                <el-input-number
                  v-model="row.shipQty"
                  :min="1"
                  :max="Math.max(1, row.remaining)"
                  :disabled="!row.selected || row.remaining <= 0"
                  size="small"
                  controls-position="right"
                  style="width: 96px"
                />
              </template>
            </el-table-column>
          </el-table>
          <div v-if="!linePicks.some((l) => l.remaining > 0)" class="hint warn">
            全部明细已关联物流，如需改绑请先删除旧发货批次
          </div>
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
  display: flex;
  align-items: center;
  gap: 12px;
}
.hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.hint.warn {
  margin-top: 8px;
  color: var(--el-color-warning);
}
.sub {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.muted {
  color: var(--el-text-color-placeholder);
}
</style>
