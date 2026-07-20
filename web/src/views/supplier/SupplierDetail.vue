<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, ArrowLeft } from '@element-plus/icons-vue'
import {
  fetchSupplier,
  fetchSupplierAddresses,
  createSupplierAddress,
  updateSupplierAddress,
  deleteSupplierAddress,
  supplierMobile,
  type Supplier,
  type SupplierAddress,
} from '../../api/supplier'

const route = useRoute()
const router = useRouter()
const supplierId = computed(() => Number(route.params.id))

const supplier = ref<Supplier | null>(null)
const addresses = ref<SupplierAddress[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref<Partial<SupplierAddress>>({})

async function loadData() {
  loading.value = true
  try {
    supplier.value = await fetchSupplier(supplierId.value)
    addresses.value = await fetchSupplierAddresses(supplierId.value)
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function handleAddAddress() {
  editing.value = { label: '', status: 1, isDefault: false }
  dialogVisible.value = true
}

function handleEditAddress(row: SupplierAddress) {
  editing.value = { ...row }
  dialogVisible.value = true
}

async function handleSaveAddress() {
  try {
    if (editing.value.id) {
      await updateSupplierAddress(supplierId.value, editing.value.id, editing.value)
    } else {
      await createSupplierAddress(supplierId.value, editing.value)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    addresses.value = await fetchSupplierAddresses(supplierId.value)
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function handleDeleteAddress(row: SupplierAddress) {
  try {
    await deleteSupplierAddress(supplierId.value, row.id)
    ElMessage.success('已删除')
    addresses.value = await fetchSupplierAddresses(supplierId.value)
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}
</script>

<template>
  <div v-loading="loading" class="detail-page">
    <el-button :icon="ArrowLeft" text @click="router.push('/suppliers')">返回列表</el-button>

    <el-card v-if="supplier" class="info-card">
      <template #header>
        <span>{{ supplier.name }}（{{ supplier.code }}）</span>
      </template>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="编码">{{ supplier.code }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ supplier.name }}</el-descriptions-item>
        <el-descriptions-item label="简称">{{ supplier.shortName || '—' }}</el-descriptions-item>
        <el-descriptions-item label="类别">{{ supplier.categoryName || '—' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="supplier.status === 1 ? 'success' : 'info'" size="small">
            {{ supplier.status === 1 ? '启用' : '停用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="采购员">{{ supplier.buyerName || '—' }}</el-descriptions-item>
        <el-descriptions-item label="截单时间">{{ supplier.cutOffTime || '—' }}</el-descriptions-item>
        <el-descriptions-item label="到货天数">{{ supplier.arrivalDays ? `${supplier.arrivalDays} 天` : '—' }}</el-descriptions-item>
        <el-descriptions-item label="账期天数">{{ supplier.paymentDays ? `${supplier.paymentDays} 天` : '—' }}</el-descriptions-item>
        <el-descriptions-item label="联系人">{{ supplier.contactName || '—' }}</el-descriptions-item>
        <el-descriptions-item label="手机">{{ supplierMobile(supplier) || '—' }}</el-descriptions-item>
        <el-descriptions-item label="办公电话">{{ supplier.officePhone || '—' }}</el-descriptions-item>
        <el-descriptions-item label="旺旺ID">{{ supplier.wangwangId || '—' }}</el-descriptions-item>
        <el-descriptions-item label="QQ">{{ supplier.qq || '—' }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ supplier.email || '—' }}</el-descriptions-item>
        <el-descriptions-item label="网址" :span="2">{{ supplier.website || '—' }}</el-descriptions-item>
        <el-descriptions-item label="地址" :span="3">{{ supplier.address || '—' }}</el-descriptions-item>
        <el-descriptions-item label="开户行">{{ supplier.bankName || '—' }}</el-descriptions-item>
        <el-descriptions-item label="账号">{{ supplier.bankAccount || '—' }}</el-descriptions-item>
        <el-descriptions-item label="户名">{{ supplier.accountName || '—' }}</el-descriptions-item>
        <el-descriptions-item label="账期说明" :span="3">{{ supplier.defaultPaymentTerms || '—' }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="3">{{ supplier.remark || '—' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card class="addr-card">
      <template #header>
        <span>发货地址</span>
        <el-button type="primary" :icon="Plus" @click="handleAddAddress">添加地址</el-button>
      </template>
      <el-table :data="addresses" stripe border>
        <el-table-column prop="label" label="标签" width="120" />
        <el-table-column label="地区" min-width="180">
          <template #default="{ row }">
            {{ [row.province, row.city, row.district].filter(Boolean).join(' ') }}
          </template>
        </el-table-column>
        <el-table-column prop="address" label="详细地址" min-width="200" />
        <el-table-column prop="contactName" label="联系人" width="100" />
        <el-table-column prop="phone" label="电话" width="120" />
        <el-table-column label="默认" width="70" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isDefault" type="success" size="small">默认</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="handleEditAddress(row)">编辑</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDeleteAddress(row)">
              <template #reference>
                <el-button type="danger" link :icon="Delete">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑地址' : '添加地址'" width="520px">
      <el-form :model="editing" label-width="90px">
        <el-form-item label="标签" required>
          <el-input v-model="editing.label" placeholder="如：深圳仓" />
        </el-form-item>
        <el-form-item label="省">
          <el-input v-model="editing.province" />
        </el-form-item>
        <el-form-item label="市">
          <el-input v-model="editing.city" />
        </el-form-item>
        <el-form-item label="区">
          <el-input v-model="editing.district" />
        </el-form-item>
        <el-form-item label="详细地址">
          <el-input v-model="editing.address" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="editing.contactName" />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="editing.phone" />
        </el-form-item>
        <el-form-item label="默认发货地">
          <el-switch v-model="editing.isDefault" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveAddress">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.detail-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.info-card :deep(.el-card__header),
.addr-card :deep(.el-card__header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
