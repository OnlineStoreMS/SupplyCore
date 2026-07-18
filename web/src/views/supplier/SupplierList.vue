<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, Search, Setting } from '@element-plus/icons-vue'
import {
  createSupplier,
  createSupplierCategory,
  deleteSupplier,
  deleteSupplierCategory,
  fetchSupplierCategories,
  fetchSuppliers,
  supplierMobile,
  updateSupplier,
  updateSupplierCategory,
  type Supplier,
  type SupplierCategory,
} from '../../api/supplier'

const router = useRouter()
const tableData = ref<Supplier[]>([])
const categories = ref<SupplierCategory[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const selectedCategoryId = ref(0)
const loading = ref(false)
const dialogVisible = ref(false)
const categoryDialogVisible = ref(false)
const categoryManageVisible = ref(false)
const editing = ref<Partial<Supplier>>({})
const editingCategory = ref<Partial<SupplierCategory>>({})

const categoryOptions = computed(() =>
  categories.value.filter((c) => c.status === 1),
)

function defaultSupplier(): Partial<Supplier> {
  return {
    code: '',
    name: '',
    status: 1,
    cutOffTime: '00:01',
    categoryId: selectedCategoryId.value || undefined,
  }
}

async function loadCategories() {
  try {
    categories.value = await fetchSupplierCategories()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载类别失败')
  }
}

async function loadData() {
  loading.value = true
  try {
    const data = await fetchSuppliers({
      keyword: keyword.value || undefined,
      categoryId: selectedCategoryId.value || undefined,
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
  await loadCategories()
  await loadData()
})

function selectCategory(id: number) {
  selectedCategoryId.value = id
  page.value = 1
  loadData()
}

function handleAdd() {
  editing.value = defaultSupplier()
  dialogVisible.value = true
}

function handleEdit(row: Supplier) {
  editing.value = {
    ...row,
    mobile: row.mobile || row.phone,
    cutOffTime: row.cutOffTime || '00:01',
  }
  dialogVisible.value = true
}

function openDetail(row: Supplier) {
  router.push(`/suppliers/${row.id}`)
}

async function handleSave() {
  try {
    const payload = { ...editing.value }
    if (payload.mobile && !payload.phone) {
      payload.phone = payload.mobile
    }
    if (payload.id) {
      await updateSupplier(payload.id, payload)
    } else {
      await createSupplier(payload)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function handleDelete(row: Supplier) {
  try {
    await deleteSupplier(row.id)
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

function handleAddCategory() {
  editingCategory.value = { name: '', status: 1, sort: 0, parentId: 0 }
  categoryDialogVisible.value = true
}

function handleEditCategory(row: SupplierCategory) {
  editingCategory.value = { ...row }
  categoryDialogVisible.value = true
}

async function handleSaveCategory() {
  try {
    if (editingCategory.value.id) {
      await updateSupplierCategory(editingCategory.value.id, editingCategory.value)
    } else {
      await createSupplierCategory(editingCategory.value)
    }
    ElMessage.success('已保存')
    categoryDialogVisible.value = false
    await loadCategories()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function handleDeleteCategory(row: SupplierCategory) {
  try {
    await deleteSupplierCategory(row.id)
    ElMessage.success('已删除')
    if (selectedCategoryId.value === row.id) {
      selectedCategoryId.value = 0
      await loadData()
    }
    await loadCategories()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}
</script>

<template>
  <div class="supplier-page">
    <div class="layout">
      <el-card class="category-panel" shadow="never">
        <template #header>
          <span>供应商类别</span>
          <el-button type="primary" link :icon="Setting" @click="categoryManageVisible = true">
            设置
          </el-button>
        </template>
        <ul class="category-list">
          <li
            :class="{ active: selectedCategoryId === 0 }"
            @click="selectCategory(0)"
          >
            全部类别
          </li>
          <li
            v-for="cat in categories"
            :key="cat.id"
            :class="{ active: selectedCategoryId === cat.id }"
            @click="selectCategory(cat.id)"
          >
            {{ cat.name }}
          </li>
        </ul>
      </el-card>

      <el-card v-loading="loading" class="main-panel">
        <template #header>
          <span>供应商</span>
          <el-button type="primary" :icon="Plus" @click="handleAdd">新增供应商</el-button>
        </template>

        <div class="toolbar">
          <el-input
            v-model="keyword"
            placeholder="搜索名称/编码"
            :prefix-icon="Search"
            clearable
            style="width: 260px"
            @change="() => { page = 1; loadData() }"
          />
        </div>

        <el-table :data="tableData" stripe border>
          <el-table-column prop="name" label="名称" min-width="140" fixed="left">
            <template #default="{ row }">
              <el-link type="primary" @click="openDetail(row)">{{ row.name }}</el-link>
            </template>
          </el-table-column>
          <el-table-column prop="code" label="编码" width="110" />
          <el-table-column prop="categoryName" label="类别" width="100" />
          <el-table-column prop="wangwangId" label="旺旺ID" width="110" show-overflow-tooltip />
          <el-table-column prop="contactName" label="联系人" width="90" />
          <el-table-column label="手机" width="120">
            <template #default="{ row }">{{ supplierMobile(row) }}</template>
          </el-table-column>
          <el-table-column prop="address" label="地址" min-width="140" show-overflow-tooltip />
          <el-table-column prop="website" label="网址" width="120" show-overflow-tooltip />
          <el-table-column prop="buyerName" label="采购员" width="90" />
          <el-table-column prop="cutOffTime" label="截单时间" width="90" align="center" />
          <el-table-column prop="paymentDays" label="账期天数" width="90" align="center" />
          <el-table-column prop="arrivalDays" label="到货天数" width="90" align="center" />
          <el-table-column label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
                {{ row.status === 1 ? '启用' : '停用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="openDetail(row)">详情</el-button>
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
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="editing.id ? '编辑供应商' : '增加供应商'"
      width="720px"
    >
      <el-form :model="editing" label-width="80px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="类别">
              <el-select v-model="editing.categoryId" clearable placeholder="选择类别" style="width: 100%">
                <el-option
                  v-for="cat in categoryOptions"
                  :key="cat.id"
                  :label="cat.name"
                  :value="cat.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12" />
          <el-col :span="12">
            <el-form-item label="名称" required>
              <el-input v-model="editing.name" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="编码" required>
              <el-input v-model="editing.code" :disabled="!!editing.id" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="采购员">
              <el-input v-model="editing.buyerName" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="截单时间">
              <el-input v-model="editing.cutOffTime" placeholder="HH:mm" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="到货天数">
              <el-input-number v-model="editing.arrivalDays" :min="0" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="账期天数">
              <el-input-number v-model="editing.paymentDays" :min="0" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="联系人">
              <el-input v-model="editing.contactName" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="地址">
              <el-input v-model="editing.address" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="办公电话">
              <el-input v-model="editing.officePhone" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机">
              <el-input v-model="editing.mobile" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="旺旺ID">
              <el-input v-model="editing.wangwangId" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="QQ">
              <el-input v-model="editing.qq" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱">
              <el-input v-model="editing.email" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="网址">
              <el-input v-model="editing.website" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="账号">
              <el-input v-model="editing.bankAccount" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-switch v-model="editing.status" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="开户行">
              <el-input v-model="editing.bankName" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="户名">
              <el-input v-model="editing.accountName" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="简称">
              <el-input v-model="editing.shortName" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="账期说明">
              <el-input v-model="editing.defaultPaymentTerms" placeholder="可选文字说明" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="editing.remark" type="textarea" :rows="2" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="categoryManageVisible" title="设置供应商类别" width="480px">
      <div class="category-toolbar">
        <el-button type="primary" :icon="Plus" size="small" @click="handleAddCategory">新增类别</el-button>
      </div>
      <el-table :data="categories" stripe border size="small">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="sort" label="排序" width="70" align="center" />
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEditCategory(row)">编辑</el-button>
            <el-popconfirm title="确定删除？有供应商的类别不可删" @confirm="handleDeleteCategory(row)">
              <template #reference>
                <el-button type="danger" link size="small">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog
      v-model="categoryDialogVisible"
      :title="editingCategory.id ? '编辑类别' : '新增类别'"
      width="400px"
    >
      <el-form :model="editingCategory" label-width="70px">
        <el-form-item label="名称" required>
          <el-input v-model="editingCategory.name" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="editingCategory.sort" :min="0" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="editingCategory.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editingCategory.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="categoryDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveCategory">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.layout {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}
.category-panel {
  width: 200px;
  flex-shrink: 0;
}
.category-panel :deep(.el-card__header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
}
.category-list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.category-list li {
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 4px;
  font-size: 14px;
}
.category-list li:hover {
  background: var(--el-fill-color-light);
}
.category-list li.active {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-weight: 500;
}
.main-panel {
  flex: 1;
  min-width: 0;
}
.main-panel :deep(.el-card__header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.toolbar {
  margin-bottom: 12px;
}
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.category-toolbar {
  margin-bottom: 12px;
}
</style>
