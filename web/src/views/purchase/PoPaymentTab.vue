<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import type { PurchaseOrder } from '../../api/purchase'
import { fetchPayments, createPayment, deletePayment, type Payment } from '../../api/poTracking'

const props = defineProps<{ poId: number; po: PurchaseOrder; readonly: boolean }>()
const emit = defineEmits<{ refresh: [] }>()

const loading = ref(false)
const list = ref<Payment[]>([])
const dialogVisible = ref(false)
const form = ref({
  payAmount: 0,
  payMethod: 'bank',
  payAccount: '',
  payeeAccount: '',
  payeeName: '',
  remark: '',
})

const paidSum = computed(() => list.value.filter((p) => p.payStatus === 'paid').reduce((s, p) => s + p.payAmount, 0))

async function loadData() {
  loading.value = true
  try {
    list.value = await fetchPayments(props.poId)
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function openCreate() {
  const remain = Math.max(0, props.po.totalAmount - paidSum.value)
  form.value = {
    payAmount: remain,
    payMethod: 'bank',
    payAccount: '',
    payeeAccount: '',
    payeeName: '',
    remark: '',
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (form.value.payAmount <= 0) {
    ElMessage.warning('请输入付款金额')
    return
  }
  try {
    await createPayment(props.poId, { ...form.value, payStatus: 'paid' })
    ElMessage.success('已记录付款')
    dialogVisible.value = false
    await loadData()
    emit('refresh')
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function handleDelete(row: Payment) {
  try {
    await ElMessageBox.confirm('确定删除此付款记录？', '确认')
  } catch {
    return
  }
  try {
    await deletePayment(props.poId, row.id)
    ElMessage.success('已删除')
    await loadData()
    emit('refresh')
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

const payMethodLabel: Record<string, string> = {
  bank: '银行转账',
  alipay: '支付宝',
  wechat: '微信',
  other: '其他',
}
</script>

<template>
  <div v-loading="loading">
    <div class="summary">
      采购总额 ¥{{ po.totalAmount.toFixed(2) }} · 已付 ¥{{ paidSum.toFixed(2) }} · 付款状态 {{ po.payStatus === 'partial' ? '部分付款' : po.payStatus === 'paid' ? '已付清' : '未付清' }}
    </div>
    <div v-if="!readonly" class="toolbar">
      <el-button type="primary" :icon="Plus" @click="openCreate">记录付款</el-button>
    </div>
    <el-table :data="list" border stripe>
      <el-table-column label="金额" width="120" align="right">
        <template #default="{ row }">¥{{ row.payAmount.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="方式" width="100">
        <template #default="{ row }">{{ payMethodLabel[row.payMethod || ''] || row.payMethod || '—' }}</template>
      </el-table-column>
      <el-table-column prop="payAccount" label="打款账号" min-width="120" />
      <el-table-column prop="payeeAccount" label="收款账号" min-width="120" />
      <el-table-column prop="payeeName" label="收款户名" width="100" />
      <el-table-column prop="paidAt" label="打款时间" width="150" />
      <el-table-column prop="remark" label="备注" min-width="100" />
      <el-table-column v-if="!readonly" label="操作" width="80">
        <template #default="{ row }">
          <el-button link type="danger" :icon="Delete" @click="handleDelete(row)" />
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" title="记录付款" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="付款金额" required>
          <el-input-number v-model="form.payAmount" :min="0.01" :precision="2" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="付款方式">
          <el-select v-model="form.payMethod" style="width: 100%">
            <el-option label="银行转账" value="bank" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="微信" value="wechat" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="打款账号">
          <el-input v-model="form.payAccount" />
        </el-form-item>
        <el-form-item label="收款账号">
          <el-input v-model="form.payeeAccount" />
        </el-form-item>
        <el-form-item label="收款户名">
          <el-input v-model="form.payeeName" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" />
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
.summary {
  margin-bottom: 12px;
  color: #606266;
  font-size: 14px;
}
.toolbar {
  margin-bottom: 12px;
}
</style>
