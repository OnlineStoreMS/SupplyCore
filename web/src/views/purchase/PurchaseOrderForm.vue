<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Plus, Delete } from '@element-plus/icons-vue'
import {
  createPurchaseOrder,
  fetchPurchaseOrder,
  updatePurchaseOrder,
  type PurchaseOrderInput,
} from '../../api/purchase'
import {
  createSkuOffer,
  fetchSuppliers,
  fetchSkuOffers,
  type Supplier,
  type SkuOffer,
} from '../../api/supplier'
import SkuSearchSelect from '../../components/SkuSearchSelect.vue'
import OrderSearchSelect from '../../components/OrderSearchSelect.vue'
import {
  resolveProductSkus,
  type ProductSkuSearchItem,
} from '../../api/productSku'
import type { OrderBrief } from '../../api/order'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => route.name === 'PurchaseOrderEdit')
const poId = computed(() => (isEdit.value ? Number(route.params.id) : 0))

const loading = ref(false)
const saving = ref(false)
const suppliers = ref<Supplier[]>([])
const offers = ref<SkuOffer[]>([])
const offerSkuMap = ref<Map<number, ProductSkuSearchItem>>(new Map())

const offerDialogVisible = ref(false)
const offerSaving = ref(false)
const offerLineIndex = ref(-1)
const offerDraft = ref({
  supplierSkuCode: '',
  supplyPrice: 0,
  supportsDropship: true,
  supportsSelfStock: false,
})

function nowOrderedAt() {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const form = ref<PurchaseOrderInput>({
  supplierId: 0,
  fulfillmentType: 'stock_in',
  currency: 'CNY',
  remark: '',
  orderedAt: nowOrderedAt(),
  items: [{ qty: 1, unitPrice: 0 }],
})

function onOrderSelect(item: OrderBrief | undefined) {
  if (!item) {
    form.value.refSoId = undefined
    form.value.refTraceId = undefined
    return
  }
  form.value.refTraceId = item.orderNo
  form.value.refSoId = item.id || undefined
  if (item.payAmount != null && item.payAmount > 0) {
    form.value.saleAmount = item.payAmount
  }
}

async function loadSuppliers() {
  const data = await fetchSuppliers({ page: 1, pageSize: 200 })
  suppliers.value = data.list
}

async function loadOffers(supplierId: number) {
  if (!supplierId) {
    offers.value = []
    offerSkuMap.value = new Map()
    return
  }
  try {
    const data = await fetchSkuOffers({ supplierId, page: 1, pageSize: 500 })
    offers.value = data.list
    offerSkuMap.value = await resolveProductSkus(data.list.map((o) => o.skuId))
  } catch {
    offers.value = []
    offerSkuMap.value = new Map()
  }
}

function offerLabel(o: SkuOffer) {
  const info = offerSkuMap.value.get(o.skuId)
  const code = info?.skuCode?.trim() || '未编码'
  const bits = [code, `¥${Number(o.supplyPrice || 0).toFixed(2)}`]
  if (o.supplierSkuCode) bits.push(`对方 ${o.supplierSkuCode}`)
  return bits.join(' · ')
}

async function loadPO() {
  if (!isEdit.value || !poId.value) return
  loading.value = true
  try {
    const po = await fetchPurchaseOrder(poId.value)
    if (po.status !== 'draft') {
      ElMessage.warning('仅草稿可编辑')
      router.replace(`/purchase-orders/${poId.value}`)
      return
    }
    form.value = {
      supplierId: po.supplierId,
      fulfillmentType: po.fulfillmentType,
      currency: po.currency,
      expectedArrivalDate: po.expectedArrivalDate,
      warehouseId: po.warehouseId,
      refSoId: po.refSoId,
      refTraceId: po.refTraceId,
      orderedAt: po.orderedAt || nowOrderedAt(),
      remark: po.remark,
      items: po.items.map((it) => ({
        skuId: it.skuId || undefined,
        offerId: it.offerId || undefined,
        productName: it.productName,
        skuCode: it.skuCode,
        skuSpecs: it.skuSpecs,
        picUrl: it.picUrl,
        supplierSkuCode: it.supplierSkuCode,
        qty: it.qty,
        saleUnitPrice: it.saleUnitPrice,
        saleAmount: it.saleAmount,
        unitPrice: it.unitPrice,
        remark: it.remark,
      })),
    }
    await loadOffers(po.supplierId)
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  if (!isEdit.value) {
    const ft = route.query.fulfillmentType
    if (ft === 'dropship' || ft === 'stock_in') {
      form.value.fulfillmentType = ft
    }
  }
  await loadSuppliers()
  await loadPO()
})

watch(() => form.value.supplierId, (id) => {
  void loadOffers(id)
})

function addLine() {
  form.value.items.push({ qty: 1, unitPrice: 0 })
}

function removeLine(index: number) {
  if (form.value.items.length <= 1) return
  form.value.items.splice(index, 1)
}

function applyOffer(index: number, offerId: number) {
  const offer = offers.value.find((o) => o.id === offerId)
  if (!offer) return
  const line = form.value.items[index]
  line.offerId = offer.id
  line.skuId = offer.skuId
  line.supplierSkuCode = offer.supplierSkuCode
  line.unitPrice = offer.supplyPrice
  const info = offerSkuMap.value.get(offer.skuId)
  if (info?.productName) line.productName = info.productName
  if (info?.skuCode) line.skuCode = info.skuCode
  if (info?.specLabel) line.skuSpecs = info.specLabel
  if (info?.pic || info?.productPic) line.picUrl = info.pic || info.productPic
}

function onSkuPicked(index: number, item: ProductSkuSearchItem | undefined) {
  const line = form.value.items[index]
  if (!item) {
    // 清空商家编码选择时，保留 OMS 已带入的商品名/图片/规格，避免编辑页被冲掉
    return
  }
  line.productName = item.productName
  line.skuCode = item.skuCode
  line.skuSpecs = item.specLabel || line.skuSpecs
  line.picUrl = item.pic || item.productPic || line.picUrl
  if (line.offerId) {
    const offer = offers.value.find((o) => o.id === line.offerId)
    if (offer && offer.skuId !== item.skuId) {
      line.offerId = undefined
    }
  }
}

function openCreateOffer(index: number) {
  const line = form.value.items[index]
  if (!form.value.supplierId) {
    ElMessage.warning('请先选择供应商')
    return
  }
  if (!line.skuId) {
    ElMessage.warning('请先选择商家编码对应的商品')
    return
  }
  offerLineIndex.value = index
  offerDraft.value = {
    supplierSkuCode: line.supplierSkuCode || '',
    supplyPrice: Number(line.unitPrice || 0),
    supportsDropship: form.value.fulfillmentType === 'dropship',
    supportsSelfStock: form.value.fulfillmentType !== 'dropship',
  }
  offerDialogVisible.value = true
}

async function saveInlineOffer() {
  const index = offerLineIndex.value
  const line = form.value.items[index]
  if (!line?.skuId || !form.value.supplierId) return
  if (offerDraft.value.supplyPrice < 0) {
    ElMessage.warning('请填写拿货价')
    return
  }
  offerSaving.value = true
  try {
    const created = await createSkuOffer({
      skuId: line.skuId,
      supplierId: form.value.supplierId,
      supplierSkuCode: offerDraft.value.supplierSkuCode,
      supplyPrice: offerDraft.value.supplyPrice,
      currency: 'CNY',
      minOrderQty: 1,
      supportsDropship: offerDraft.value.supportsDropship,
      supportsSelfStock: offerDraft.value.supportsSelfStock,
      status: 1,
    })
    await loadOffers(form.value.supplierId)
    line.offerId = created.id
    line.unitPrice = created.supplyPrice
    line.supplierSkuCode = created.supplierSkuCode || offerDraft.value.supplierSkuCode
    offerDialogVisible.value = false
    ElMessage.success('已保存到 SKU 供货报价并应用到本行')
  } catch (e) {
    ElMessage.error((e as Error).message || '保存报价失败')
  } finally {
    offerSaving.value = false
  }
}

const lineTotal = computed(() =>
  form.value.items.reduce((sum, it) => sum + it.qty * (it.unitPrice || 0), 0),
)

async function handleSave() {
  if (!form.value.supplierId) {
    ElMessage.warning('请选择供应商')
    return
  }
  const dropship = form.value.fulfillmentType === 'dropship'
  for (const it of form.value.items) {
    if (it.qty <= 0) {
      ElMessage.warning('请完善明细数量')
      return
    }
    if (!dropship && !it.skuId) {
      ElMessage.warning('请选择商家编码')
      return
    }
    if (!dropship && (it.unitPrice == null || it.unitPrice <= 0) && !it.offerId) {
      ElMessage.warning('请填写采购单价或选择供货报价')
      return
    }
    if (dropship && !it.skuId && !it.productName) {
      ElMessage.warning('代发明细需有商品名称或商家编码')
      return
    }
  }
  saving.value = true
  try {
    if (isEdit.value) {
      await updatePurchaseOrder(poId.value, form.value)
      ElMessage.success('已保存')
      router.push(`/purchase-orders/${poId.value}`)
    } else {
      const po = await createPurchaseOrder(form.value)
      ElMessage.success('已创建')
      router.push(`/purchase-orders/${po.id}`)
    }
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div v-loading="loading" class="po-form">
    <el-button :icon="ArrowLeft" text @click="router.push('/purchase-orders')">返回列表</el-button>

    <el-card>
      <template #header>{{ isEdit ? '编辑供应商订单' : '新建供应商订单' }}</template>

      <el-form label-width="100px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="供应商" required>
              <el-select v-model="form.supplierId" filterable placeholder="选择供应商" style="width: 100%">
                <el-option v-for="s in suppliers" :key="s.id" :label="`${s.name} (${s.code})`" :value="s.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="订单类型">
              <el-select v-model="form.fulfillmentType" style="width: 100%">
                <el-option label="采购入仓" value="stock_in" />
                <el-option label="代发直邮" value="dropship" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="预计到货">
              <el-date-picker
                v-model="form.expectedArrivalDate"
                type="date"
                value-format="YYYY-MM-DD"
                placeholder="可选"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="关联订单">
              <OrderSearchSelect
                v-model="form.refTraceId"
                @select="onOrderSelect"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="下单时间">
              <el-date-picker
                v-model="form.orderedAt"
                type="datetime"
                value-format="YYYY-MM-DD HH:mm:ss"
                format="YYYY-MM-DD HH:mm"
                placeholder="默认当前时间"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="form.remark" type="textarea" :rows="2" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <div class="lines-header">
        <span>采购明细</span>
        <el-button type="primary" link :icon="Plus" @click="addLine">添加行</el-button>
      </div>
      <div class="hint-bar">
        可不选供货报价，直接填写采购单价；也可「存为报价」写入 SKU 供货报价后继续下单。
      </div>

      <el-table :data="form.items" border size="small" class="lines-table">
        <el-table-column label="图片" width="72" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.picUrl"
              :src="row.picUrl"
              :preview-src-list="[row.picUrl]"
              fit="cover"
              class="sku-pic"
              preview-teleported
            />
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="商品" min-width="140">
          <template #default="{ row }">
            <div class="name">{{ row.productName || '—' }}</div>
            <div v-if="row.remark" class="sub">{{ row.remark }}</div>
          </template>
        </el-table-column>
        <el-table-column label="规格" width="120">
          <template #default="{ row }">
            <el-input v-model="row.skuSpecs" placeholder="规格" />
          </template>
        </el-table-column>
        <el-table-column label="供货报价" width="200">
          <template #default="{ row, $index }">
            <el-select
              :model-value="row.offerId"
              placeholder="可选"
              clearable
              filterable
              style="width: 100%"
              @update:model-value="(v: number) => applyOffer($index, v)"
            >
              <el-option v-for="o in offers" :key="o.id" :label="offerLabel(o)" :value="o.id" />
            </el-select>
            <el-button type="primary" link size="small" @click="openCreateOffer($index)">存为报价</el-button>
          </template>
        </el-table-column>
        <el-table-column label="商家编码" min-width="200">
          <template #default="{ row, $index }">
            <SkuSearchSelect
              v-model="row.skuId"
              :show-preview="false"
              placeholder="搜索商家编码（可选）"
              @select="(item) => onSkuPicked($index, item)"
            />
          </template>
        </el-table-column>
        <el-table-column label="对方货号" width="110">
          <template #default="{ row }">
            <el-input v-model="row.supplierSkuCode" placeholder="供应商货号" />
          </template>
        </el-table-column>
        <el-table-column label="数量" width="90">
          <template #default="{ row }">
            <el-input-number v-model="row.qty" :min="1" controls-position="right" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="实付金额" width="100" align="right">
          <template #default="{ row }">
            <span v-if="row.saleAmount">¥{{ Number(row.saleAmount).toFixed(2) }}</span>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="采购单价" width="120">
          <template #default="{ row }">
            <el-input-number
              v-model="row.unitPrice"
              :min="0"
              :precision="2"
              controls-position="right"
              style="width: 100%"
            />
          </template>
        </el-table-column>
        <el-table-column label="采购小计" width="90" align="right">
          <template #default="{ row }">¥{{ (row.qty * (row.unitPrice || 0)).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column width="50" align="center">
          <template #default="{ $index }">
            <el-button type="danger" link :icon="Delete" @click="removeLine($index)" />
          </template>
        </el-table-column>
      </el-table>

      <div class="footer">
        <span class="total">合计：¥{{ lineTotal.toFixed(2) }}</span>
        <div>
          <el-button @click="router.push('/purchase-orders')">取消</el-button>
          <el-button type="primary" :loading="saving" @click="handleSave">保存草稿</el-button>
        </div>
      </div>
    </el-card>

    <el-dialog v-model="offerDialogVisible" title="现场新增供货报价" width="440px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="对方货号">
          <el-input v-model="offerDraft.supplierSkuCode" placeholder="供应商侧货号（可选）" />
        </el-form-item>
        <el-form-item label="拿货价" required>
          <el-input-number
            v-model="offerDraft.supplyPrice"
            :min="0"
            :precision="2"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="支持代发">
          <el-switch v-model="offerDraft.supportsDropship" />
        </el-form-item>
        <el-form-item label="供货到仓">
          <el-switch v-model="offerDraft.supportsSelfStock" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="offerDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="offerSaving" @click="saveInlineOffer">保存并应用</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.po-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.lines-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 16px 0 4px;
  font-weight: 600;
}
.hint-bar {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}
.sku-pic {
  width: 40px;
  height: 40px;
  border-radius: 4px;
}
.name {
  font-size: 13px;
  line-height: 1.35;
}
.sub {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}
.muted {
  color: #c0c4cc;
}
.footer {
  margin-top: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.total {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}
.lines-table :deep(.el-table__cell) {
  vertical-align: top;
}
</style>
