<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, Search } from '@element-plus/icons-vue'
import {
  createSupplier, deleteSupplier, fetchSuppliers, updateSupplier, type Supplier,
} from '../../api/supplier'

const router = useRouter()
const tableData = ref<Supplier[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref<Partial<Supplier>>({})

async function loadData() {
  loading.value = true
  try {
    const data = await fetchSuppliers(keyword.value || undefined, page.value, pageSize.value)
    tableData.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function handleAdd() {
  editing.value = { code: '', name: '', status: 1 }
  dialogVisible.value = true
}

function handleEdit(row: Supplier) {
  editing.value = { ...row }
  dialogVisible.value = true
}

function openDetail(row: Supplier) {
  router.push(`/suppliers/${row.id}`)
}

async function handleSave() {
  try {
    if (editing.value.id) {
      await updateSupplier(editing.value.id, editing.value)
    } else {
      await createSupplier(editing.value)
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
</script>

<template>
  <div class="supplier-page">
    <el-card v-loading="loading">
      <template #header>
        <span>供应商</span>
        <el-button type="primary" :icon="Plus" @click="handleAdd">添加供应商</el-button>
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
        <el-table-column prop="code" label="编码" width="120" />
        <el-table-column prop="name" label="名称" min-width="160">
          <template #default="{ row }">
            <el-link type="primary" @click="openDetail(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="shortName" label="简称" width="120" />
        <el-table-column prop="contactName" label="联系人" width="100" />
        <el-table-column prop="phone" label="电话" width="130" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
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

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑供应商' : '添加供应商'" width="520px">
      <el-form :model="editing" label-width="90px">
        <el-form-item label="编码" required>
          <el-input v-model="editing.code" :disabled="!!editing.id" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="editing.name" />
        </el-form-item>
        <el-form-item label="简称">
          <el-input v-model="editing.shortName" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="editing.contactName" />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="editing.phone" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editing.email" />
        </el-form-item>
        <el-form-item label="账期说明">
          <el-input v-model="editing.defaultPaymentTerms" />
        </el-form-item>
        <el-form-item label="开户行">
          <el-input v-model="editing.bankName" />
        </el-form-item>
        <el-form-item label="银行账号">
          <el-input v-model="editing.bankAccount" />
        </el-form-item>
        <el-form-item label="户名">
          <el-input v-model="editing.accountName" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="editing.status" :active-value="1" :inactive-value="0" />
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
.supplier-page :deep(.el-card__header) {
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
</style>
