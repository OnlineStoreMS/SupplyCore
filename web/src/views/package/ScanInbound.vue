<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createInboundFromPackage, scanPackage, type PackageReceive } from '../../api/purchaseExt'
import { fetchPurchaseOrders, type PurchaseOrderListItem } from '../../api/purchase'

const router = useRouter()
const form = reactive({
  trackingNo: '',
  carrier: '',
  warehouseName: '',
  packageType: 'normal',
  poId: undefined as number | undefined,
  poNo: '',
})
const last = ref<PackageReceive | null>(null)
const poList = ref<PurchaseOrderListItem[]>([])
const scanning = ref(false)

async function loadPOs() {
  try {
    const data = await fetchPurchaseOrders({ page: 1, pageSize: 100 })
    poList.value = data.list
  } catch {
    poList.value = []
  }
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

void loadPOs()
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
          <el-input v-model="form.warehouseName" style="width: 360px" />
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
        <p>采购单号：{{ last.poNo || '-' }}</p>
        <p>入库单号：{{ last.inboundNo || '未生成' }}</p>
        <p>扫描时间：{{ last.createdAt }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.toolbar h2 { margin: 0; font-size: 18px; }
.scan-panel {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 24px;
}
.last-scan {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
  line-height: 1.8;
}
.last-scan h3 { margin: 0 0 8px; font-size: 15px; }
.last-scan p { margin: 0; color: #606266; font-size: 13px; }
</style>
