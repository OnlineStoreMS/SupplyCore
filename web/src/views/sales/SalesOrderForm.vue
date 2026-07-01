<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Plus, Delete } from '@element-plus/icons-vue'
import {
  createSalesOrder,
  fetchSalesOrder,
  updateSalesOrder,
  type SalesOrderInput,
} from '../../api/salesOrder'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => route.name === 'SalesOrderEdit')
const soId = computed(() => (isEdit.value ? Number(route.params.id) : 0))

const loading = ref(false)
const saving = ref(false)

const form = ref<SalesOrderInput>({
  sourceChannel: 'manual',
  receiverName: '',
  receiverPhone: '',
  province: '',
  city: '',
  district: '',
  receiverAddress: '',
  remark: '',
  items: [{ skuId: 0, qty: 1, fulfillmentMode: 'dropship' }],
})

async function loadSO() {
  if (!isEdit.value || !soId.value) return
  loading.value = true
  try {
    const so = await fetchSalesOrder(soId.value)
    if (so.status !== 'draft') {
      ElMessage.warning('仅草稿可编辑')
      router.replace(`/sales-orders/${soId.value}`)
      return
    }
    form.value = {
      sourceChannel: so.sourceChannel,
      receiverName: so.receiverName,
      receiverPhone: so.receiverPhone,
      province: so.province,
      city: so.city,
      district: so.district,
      receiverAddress: so.receiverAddress,
      remark: so.remark,
      items: so.items.map((it) => ({
        skuId: it.skuId,
        qty: it.qty,
        fulfillmentMode: it.fulfillmentMode || 'dropship',
        remark: it.remark,
      })),
    }
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadSO)

function addLine() {
  form.value.items.push({ skuId: 0, qty: 1, fulfillmentMode: 'dropship' })
}

function removeLine(index: number) {
  if (form.value.items.length <= 1) return
  form.value.items.splice(index, 1)
}

async function save() {
  if (!form.value.items.every((it) => it.skuId > 0 && it.qty > 0)) {
    ElMessage.warning('请填写 SKU 与数量')
    return
  }
  saving.value = true
  try {
    if (isEdit.value) {
      await updateSalesOrder(soId.value, form.value)
      ElMessage.success('已保存')
      router.push(`/sales-orders/${soId.value}`)
    } else {
      const so = await createSalesOrder(form.value)
      ElMessage.success('已创建')
      router.push(`/sales-orders/${so.id}`)
    }
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div v-loading="loading" class="so-form-page">
    <el-page-header :icon="ArrowLeft" @back="router.back()">
      <template #content>{{ isEdit ? '编辑销售单' : '新建销售单' }}</template>
    </el-page-header>

    <el-card class="form-card">
      <el-form label-width="100px">
        <el-divider content-position="left">收货信息</el-divider>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="收货人">
              <el-input v-model="form.receiverName" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="电话">
              <el-input v-model="form.receiverPhone" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="来源">
              <el-input v-model="form.sourceChannel" placeholder="manual" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8"><el-form-item label="省"><el-input v-model="form.province" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="市"><el-input v-model="form.city" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="区"><el-input v-model="form.district" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="详细地址">
          <el-input v-model="form.receiverAddress" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>

        <el-divider content-position="left">明细</el-divider>
        <el-table :data="form.items" border size="small">
          <el-table-column label="SKU ID" width="140">
            <template #default="{ row }">
              <el-input-number v-model="row.skuId" :min="1" controls-position="right" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="数量" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.qty" :min="1" controls-position="right" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="履约方式" width="140">
            <template #default="{ row }">
              <el-select v-model="row.fulfillmentMode" style="width: 100%">
                <el-option label="代发" value="dropship" />
                <el-option label="自发货" value="self" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="备注">
            <template #default="{ row }">
              <el-input v-model="row.remark" />
            </template>
          </el-table-column>
          <el-table-column label="" width="60">
            <template #default="{ $index }">
              <el-button link type="danger" :icon="Delete" @click="removeLine($index)" />
            </template>
          </el-table-column>
        </el-table>
        <el-button class="add-line" :icon="Plus" @click="addLine">添加行</el-button>

        <div class="actions">
          <el-button @click="router.back()">取消</el-button>
          <el-button type="primary" :loading="saving" @click="save">保存</el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.so-form-page {
  max-width: 960px;
}
.form-card {
  margin-top: 16px;
}
.add-line {
  margin-top: 12px;
}
.actions {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
