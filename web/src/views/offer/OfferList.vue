<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, Search } from '@element-plus/icons-vue'
import SkuSearchSelect from '../../components/SkuSearchSelect.vue'
import {
  createSkuOffer,
  deleteSkuOffer,
  fetchSkuOffers,
  fetchSuppliers,
  updateSkuOffer,
  type SkuOffer,
  type Supplier,
} from '../../api/supplier'

const tableData = ref<SkuOffer[]>([])
const suppliers = ref<Supplier[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filterSkuId = ref<number | undefined>()
const filterSupplierId = ref<number | undefined>()
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref<Partial<SkuOffer>>({
  currency: 'CNY',
  minOrderQty: 1,
  supportsSelfStock: true,
  status: 1,
})

async function loadSuppliers() {
  try {
    const data = await fetchSuppliers({ page: 1, pageSize: 200 })
    suppliers.value = data.list
  } catch {
    suppliers.value = []
  }
}

async function loadData() {
  loading.value = true
  try {
    const data = await fetchSkuOffers({
      skuId: filterSkuId.value,
      supplierId: filterSupplierId.value,
      page: page.value,
      pageSize: pageSize.value,
    })
    tableData.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadSuppliers()
  await loadData()
})

watch(filterSkuId, () => {
  page.value = 1
  void loadData()
})

function handleAdd() {
  editing.value = {
    skuId: filterSkuId.value,
    supplierId: filterSupplierId.value,
    supplyPrice: 0,
    currency: 'CNY',
    minOrderQty: 1,
    supportsDropship: false,
    supportsSelfStock: true,
    isPrimary: false,
    priority: 0,
    status: 1,
  }
  dialogVisible.value = true
}

function handleEdit(row: SkuOffer) {
  editing.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (!editing.value.skuId || !editing.value.supplierId) {
      ElMessage.warning('请填写 SKU ID 与供应商')
      return
    }
    if (editing.value.id) {
      await updateSkuOffer(editing.value.id, editing.value)
    } else {
      await createSkuOffer(editing.value)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function handleDelete(row: SkuOffer) {
  try {
    await deleteSkuOffer(row.id)
    ElMessage.success('已删除')
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

function onPageChange(p: number) {
  page.value = p
  loadData()
}
</script>

<template>
  <div class="offer-page">
    <el-card v-loading="loading">
      <template #header>
        <span>SKU 供货报价</span>
        <el-button type="primary" :icon="Plus" @click="handleAdd">添加报价</el-button>
      </template>

      <div class="toolbar">
        <div class="toolbar-sku">
          <SkuSearchSelect v-model="filterSkuId" placeholder="按 SKU 搜索筛选" clearable />
        </div>
        <el-select
          v-model="filterSupplierId"
          placeholder="供应商"
          clearable
          filterable
          style="width: 200px"
          @change="() => { page = 1; loadData() }"
        >
          <el-option v-for="s in suppliers" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-button :icon="Search" @click="loadData">查询</el-button>
      </div>

      <el-table :data="tableData" stripe border>
        <el-table-column prop="skuId" label="SKU ID" width="90" />
        <el-table-column prop="supplierName" label="供应商" min-width="140" />
        <el-table-column prop="supplierSkuCode" label="对方货号" width="120" />
        <el-table-column label="拿货价" width="100" align="right">
          <template #default="{ row }">¥{{ row.supplyPrice.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="shipFromCity" label="发货地" width="100" />
        <el-table-column label="代发" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.supportsDropship ? 'success' : 'info'" size="small">
              {{ row.supportsDropship ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="leadTimeDays" label="交期(天)" width="90" align="center" />
        <el-table-column label="主供" width="70" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isPrimary" type="warning" size="small">主</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row)">
              <template #reference>
                <el-button type="danger" link :icon="Delete">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="onPageChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑报价' : '添加报价'" width="560px">
      <el-form :model="editing" label-width="110px">
        <el-form-item label="商品 SKU" required>
          <SkuSearchSelect v-model="editing.skuId" :disabled="!!editing.id" />
          <div v-if="editing.id" class="hint">编辑时不可更换 SKU</div>
        </el-form-item>
        <el-form-item label="供应商" required>
          <el-select v-model="editing.supplierId" filterable style="width: 100%">
            <el-option v-for="s in suppliers" :key="s.id" :label="`${s.name} (${s.code})`" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="对方货号">
          <el-input v-model="editing.supplierSkuCode" />
        </el-form-item>
        <el-form-item label="拿货价" required>
          <el-input-number v-model="editing.supplyPrice" :min="0" :precision="2" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="发货地址 ID">
          <el-input-number v-model="editing.shipFromAddressId" :min="0" controls-position="right" style="width: 100%" />
          <div class="hint">在供应商详情页维护地址后填写对应 ID</div>
        </el-form-item>
        <el-form-item label="起订量">
          <el-input-number v-model="editing.minOrderQty" :min="1" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="交期(天)">
          <el-input-number v-model="editing.leadTimeDays" :min="0" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="支持代发">
          <el-switch v-model="editing.supportsDropship" />
        </el-form-item>
        <el-form-item label="供货到仓">
          <el-switch v-model="editing.supportsSelfStock" />
        </el-form-item>
        <el-form-item label="主供应商">
          <el-switch v-model="editing.isPrimary" />
        </el-form-item>
        <el-form-item label="排序优先级">
          <el-input-number v-model="editing.priority" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editing.remark" type="textarea" :rows="2" />
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
.offer-page :deep(.el-card__header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  align-items: flex-start;
}
.toolbar-sku {
  width: 320px;
}
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
