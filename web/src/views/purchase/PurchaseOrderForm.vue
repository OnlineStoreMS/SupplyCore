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
import { fetchSuppliers, fetchSkuOffers, type Supplier, type SkuOffer } from '../../api/supplier'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => route.name === 'PurchaseOrderEdit')
const poId = computed(() => (isEdit.value ? Number(route.params.id) : 0))

const loading = ref(false)
const saving = ref(false)
const suppliers = ref<Supplier[]>([])
const offers = ref<SkuOffer[]>([])

const form = ref<PurchaseOrderInput>({
  supplierId: 0,
  fulfillmentType: 'stock_in',
  currency: 'CNY',
  remark: '',
  items: [{ skuId: 0, qty: 1, unitPrice: 0 }],
})

async function loadSuppliers() {
  const data = await fetchSuppliers(undefined, 1, 200)
  suppliers.value = data.list
}

async function loadOffers(supplierId: number) {
  if (!supplierId) {
    offers.value = []
    return
  }
  try {
    const data = await fetchSkuOffers({ supplierId, page: 1, pageSize: 500 })
    offers.value = data.list
  } catch {
    offers.value = []
  }
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
      remark: po.remark,
      items: po.items.map((it) => ({
        skuId: it.skuId,
        offerId: it.offerId || undefined,
        supplierSkuCode: it.supplierSkuCode,
        qty: it.qty,
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
  await loadSuppliers()
  await loadPO()
})

watch(() => form.value.supplierId, (id) => {
  void loadOffers(id)
})

function addLine() {
  form.value.items.push({ skuId: 0, qty: 1, unitPrice: 0 })
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
}

const lineTotal = computed(() =>
  form.value.items.reduce((sum, it) => sum + it.qty * (it.unitPrice || 0), 0),
)

async function handleSave() {
  if (!form.value.supplierId) {
    ElMessage.warning('请选择供应商')
    return
  }
  if (form.value.items.some((it) => !it.skuId || it.qty <= 0)) {
    ElMessage.warning('请完善明细行')
    return
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
      <template #header>{{ isEdit ? '编辑采购单' : '新建采购单' }}</template>

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
            <el-form-item label="履约类型">
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
            <el-form-item label="外部订单 ID">
              <el-input-number v-model="form.refSoId" :min="0" controls-position="right" style="width: 100%" />
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

      <el-table :data="form.items" border size="small">
        <el-table-column label="供货报价" width="220">
          <template #default="{ row, $index }">
            <el-select
              :model-value="row.offerId"
              placeholder="从报价带入"
              clearable
              filterable
              style="width: 100%"
              @update:model-value="(v: number) => applyOffer($index, v)"
            >
              <el-option
                v-for="o in offers"
                :key="o.id"
                :label="`SKU ${o.skuId} · ¥${o.supplyPrice}`"
                :value="o.id"
              />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="SKU ID" width="120">
          <template #default="{ row }">
            <el-input-number v-model="row.skuId" :min="1" controls-position="right" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="对方货号" width="120">
          <template #default="{ row }">
            <el-input v-model="row.supplierSkuCode" />
          </template>
        </el-table-column>
        <el-table-column label="数量" width="100">
          <template #default="{ row }">
            <el-input-number v-model="row.qty" :min="1" controls-position="right" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="单价" width="120">
          <template #default="{ row }">
            <el-input-number v-model="row.unitPrice" :min="0" :precision="2" controls-position="right" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="小计" width="100" align="right">
          <template #default="{ row }">¥{{ (row.qty * (row.unitPrice || 0)).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="备注" min-width="120">
          <template #default="{ row }">
            <el-input v-model="row.remark" />
          </template>
        </el-table-column>
        <el-table-column width="60" align="center">
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
  margin: 16px 0 8px;
  font-weight: 600;
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
</style>
