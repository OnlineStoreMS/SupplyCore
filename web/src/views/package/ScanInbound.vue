<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createInboundFromPackage, scanPackage, type PackageReceive } from '../../api/purchaseExt'
import { fetchPurchaseOrders, type PurchaseOrderListItem } from '../../api/purchase'
import { listWarehouses, type Warehouse } from '../../api/warehouse'

const router = useRouter()
const form = reactive({
  trackingNo: '',
  carrier: '',
  warehouseId: undefined as number | undefined,
  warehouseName: '',
  packageType: 'normal',
  poId: undefined as number | undefined,
  poNo: '',
})
const last = ref<PackageReceive | null>(null)
const poList = ref<PurchaseOrderListItem[]>([])
const warehouses = ref<Warehouse[]>([])
const scanning = ref(false)

async function loadPOs() {
  try {
    const data = await fetchPurchaseOrders({ page: 1, pageSize: 100 })
    poList.value = data.list
  } catch {
    poList.value = []
  }
}

async function loadWarehouses() {
  try {
    const data = await listWarehouses({ page: 1, pageSize: 200 })
    warehouses.value = data.list.filter((w) => w.status === 1)
    const def = warehouses.value.find((w) => w.isDefault === 1)
    if (def && !form.warehouseId) {
      form.warehouseId = def.id
      form.warehouseName = def.name
    }
  } catch {
    warehouses.value = []
  }
}

function onWarehouseChange(id: number) {
  const wh = warehouses.value.find((w) => w.id === id)
  form.warehouseName = wh?.name || ''
}

async function onScan() {
  if (!form.trackingNo.trim()) {
    ElMessage.warning('请输入物流单号')
    return
  }
  scanning.value = true
  try {
    const po = poList.value.find((p) => p.id === form.poId)
    last.value = await scanPackage({
      trackingNo: form.trackingNo.trim(),
      carrier: form.carrier,
      warehouseId: form.warehouseId,
      warehouseName: form.warehouseName,
      packageType: form.packageType,
      poId: form.poId,
      poNo: po?.poNo,
    })
    ElMessage.success('扫描成功')
    form.trackingNo = ''
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    scanning.value = false
  }
}

async function genInbound() {
  if (!last.value) return
  try {
    const inbound = await createInboundFromPackage(last.value.id)
    ElMessage.success(`已生成入库单 ${inbound.inboundNo}`)
    await router.push(`/purchase-inbounds/${inbound.id}`)
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

onMounted(() => {
  void loadPOs()
  void loadWarehouses()
})
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>包裹扫描入库</h2>
      <el-button @click="router.push('/package-receives')">查看收货记录</el-button>
    </div>
    <div class="scan-panel">
      <el-form label-width="100px">
        <el-form-item label="扫描单号" required>
          <el-input
            v-model="form.trackingNo"
            placeholder="物流单号 / 采购单号"
            style="width: 360px"
            @keyup.enter="onScan"
          />
        </el-form-item>
        <el-form-item label="快递公司">
          <el-input v-model="form.carrier" style="width: 360px" />
        </el-form-item>
        <el-form-item label="仓库">
          <el-select
            v-model="form.warehouseId"
            filterable
            clearable
            placeholder="选择 WarehouseCore 仓库"
            style="width: 360px"
            @change="onWarehouseChange"
          >
            <el-option
              v-for="w in warehouses"
              :key="w.id"
              :label="`${w.name} (${w.code})`"
              :value="w.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="关联采购单">
          <el-select v-model="form.poId" clearable filterable style="width: 360px">
            <el-option v-for="p in poList" :key="p.id" :label="`${p.poNo} · ${p.supplierName}`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="包裹类型">
          <el-select v-model="form.packageType" style="width: 360px">
            <el-option label="普通" value="normal" />
            <el-option label="紧急" value="urgent" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="scanning" @click="onScan">扫描入库</el-button>
          <el-button v-if="last?.poId && !last.inboundId" type="success" @click="genInbound">
            生成采购入库单
          </el-button>
        </el-form-item>
      </el-form>
      <div v-if="last" class="last-scan">
        <h3>上次扫描</h3>
        <p>物流单号：{{ last.trackingNo }}</p>
        <p>仓库：{{ last.warehouseName || '-' }}</p>
        <p>采购单号：{{ last.poNo || '-' }}</p>
        <p>入库单号：{{ last.inboundNo || '未生成' }}</p>
        <p>扫描时间：{{ last.createdAt }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.scan-panel {
  display: flex;
  gap: 32px;
}
.last-scan {
  min-width: 280px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}
</style>
