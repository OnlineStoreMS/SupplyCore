<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createPurchaseAccount, deletePurchaseAccount, fetchPurchaseAccounts, updatePurchaseAccount,
  type PurchaseAccount,
} from '../../api/purchaseExt'

const loading = ref(false)
const list = ref<PurchaseAccount[]>([])
const total = ref(0)
const page = ref(1)
const keyword = ref('')
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({
  channel: 'alibaba1688',
  accountAlias: '',
  accountName: '',
  isPrimary: false,
  status: 'active',
  authStatus: 'unauthorized',
  remark: '',
})

async function load() {
  loading.value = true
  try {
    const data = await fetchPurchaseAccounts(keyword.value || undefined, page.value, 20)
    list.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, {
    channel: 'alibaba1688', accountAlias: '', accountName: '',
    isPrimary: false, status: 'active', authStatus: 'unauthorized', remark: '',
  })
  dialogVisible.value = true
}

function openEdit(row: PurchaseAccount) {
  editingId.value = row.id
  Object.assign(form, {
    channel: row.channel, accountAlias: row.accountAlias, accountName: row.accountName,
    isPrimary: row.isPrimary, status: row.status, authStatus: row.authStatus, remark: row.remark,
  })
  dialogVisible.value = true
}

async function submit() {
  try {
    if (editingId.value) {
      await updatePurchaseAccount(editingId.value, { ...form })
    } else {
      await createPurchaseAccount({ ...form })
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function onDelete(row: PurchaseAccount) {
  await ElMessageBox.confirm(`确定删除账号「${row.accountAlias}」？`, '确认')
  try {
    await deletePurchaseAccount(row.id)
    ElMessage.success('已删除')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

const channelLabel = (c: string) => ({ alibaba1688: '1688', taogongxiao: '淘供销', other: '其他' }[c] || c)

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>采购账号</h2>
      <div class="actions">
        <el-input v-model="keyword" clearable placeholder="账号简称/名称" style="width: 200px" @keyup.enter="load" />
        <el-button @click="load">查询</el-button>
        <el-button type="primary" @click="openCreate">新增账号</el-button>
      </div>
    </div>
    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="accountAlias" label="账号简称" width="140" />
      <el-table-column prop="accountName" label="采购账号" min-width="160" />
      <el-table-column label="渠道" width="100">
        <template #default="{ row }">{{ channelLabel(row.channel) }}</template>
      </el-table-column>
      <el-table-column label="主账号" width="80">
        <template #default="{ row }">{{ row.isPrimary ? '是' : '否' }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="90" />
      <el-table-column prop="authStatus" label="授权状态" width="120" />
      <el-table-column prop="operatorName" label="操作人" width="100" />
      <el-table-column prop="createdAt" label="创建时间" width="170" />
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="page"
      class="pager"
      layout="total, prev, pager, next"
      :total="total"
      @current-change="load"
    />

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑采购账号' : '新增采购账号'" width="480px">
      <el-form label-width="90px">
        <el-form-item label="渠道" required>
          <el-select v-model="form.channel" style="width: 100%">
            <el-option label="1688" value="alibaba1688" />
            <el-option label="淘供销" value="taogongxiao" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="账号简称" required>
          <el-input v-model="form.accountAlias" />
        </el-form-item>
        <el-form-item label="采购账号" required>
          <el-input v-model="form.accountName" />
        </el-form-item>
        <el-form-item label="主账号">
          <el-switch v-model="form.isPrimary" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="启用" value="active" />
            <el-option label="停用" value="disabled" />
            <el-option label="作废" value="revoked" />
          </el-select>
        </el-form-item>
        <el-form-item label="授权状态">
          <el-select v-model="form.authStatus" style="width: 100%">
            <el-option label="未授权" value="unauthorized" />
            <el-option label="已授权" value="authorized" />
            <el-option label="已过期" value="expired" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.toolbar h2 { margin: 0; font-size: 18px; }
.actions { display: flex; gap: 8px; }
.pager { margin-top: 16px; justify-content: flex-end; }
</style>
