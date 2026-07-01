<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  formatSkuOptionLabel,
  searchProductSkus,
  type ProductSkuSearchItem,
} from '../api/productSku'

const model = defineModel<number | undefined>({ default: undefined })

withDefaults(
  defineProps<{
    placeholder?: string
    disabled?: boolean
    clearable?: boolean
  }>(),
  {
    placeholder: '输入 SKU 编码、规格值或商品名搜索',
    disabled: false,
    clearable: true,
  },
)

const loading = ref(false)
const options = ref<ProductSkuSearchItem[]>([])
const selectedLabel = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | undefined

function mergeOptions(items: ProductSkuSearchItem[]) {
  const map = new Map<number, ProductSkuSearchItem>()
  for (const item of options.value) {
    map.set(item.skuId, item)
  }
  for (const item of items) {
    map.set(item.skuId, item)
  }
  options.value = Array.from(map.values())
}

async function remoteSearch(keyword: string) {
  const q = keyword.trim()
  if (!q) {
    return
  }
  loading.value = true
  try {
    const data = await searchProductSkus({ keyword: q, page: 1, pageSize: 20 })
    mergeOptions(data.list)
  } catch (e) {
    ElMessage.error((e as Error).message || 'SKU 搜索失败')
  } finally {
    loading.value = false
  }
}

function onSearch(keyword: string) {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    void remoteSearch(keyword)
  }, 300)
}

async function loadSelected(id?: number) {
  if (!id) {
    selectedLabel.value = ''
    return
  }
  const existing = options.value.find((item) => item.skuId === id)
  if (existing) {
    selectedLabel.value = formatSkuOptionLabel(existing)
    return
  }
  loading.value = true
  try {
    const data = await searchProductSkus({ keyword: String(id), page: 1, pageSize: 5 })
    const hit = data.list.find((item) => item.skuId === id) || data.list[0]
    if (hit) {
      mergeOptions([hit])
      selectedLabel.value = formatSkuOptionLabel(hit)
    } else {
      selectedLabel.value = `#${id}`
    }
  } catch {
    selectedLabel.value = `#${id}`
  } finally {
    loading.value = false
  }
}

function onChange(value: number | undefined) {
  if (!value) {
    selectedLabel.value = ''
    return
  }
  const hit = options.value.find((item) => item.skuId === value)
  if (hit) {
    selectedLabel.value = formatSkuOptionLabel(hit)
  }
}

onMounted(() => {
  void loadSelected(model.value)
})

watch(
  () => model.value,
  (id) => {
    void loadSelected(id)
  },
)
</script>

<template>
  <el-select
    v-model="model"
    filterable
    remote
    reserve-keyword
    :remote-method="onSearch"
    :loading="loading"
    :placeholder="placeholder"
    :disabled="disabled"
    :clearable="clearable"
    style="width: 100%"
    @change="onChange"
  >
    <el-option
      v-for="item in options"
      :key="item.skuId"
      :label="formatSkuOptionLabel(item)"
      :value="item.skuId"
    >
      <div class="sku-option">
        <span class="sku-option-main">{{ item.skuCode || `#${item.skuId}` }}</span>
        <span class="sku-option-spec">{{ item.specLabel || '-' }}</span>
        <span class="sku-option-sub">{{ item.productName }}</span>
      </div>
    </el-option>
  </el-select>
  <div v-if="model && selectedLabel" class="selected-preview">
    已选：{{ selectedLabel }}
  </div>
</template>

<style scoped>
.sku-option {
  display: flex;
  flex-direction: column;
  line-height: 1.35;
  padding: 2px 0;
}
.sku-option-main {
  font-weight: 500;
  color: var(--el-text-color-primary);
}
.sku-option-spec {
  font-size: 12px;
  color: var(--el-text-color-regular);
}
.sku-option-sub {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.selected-preview {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
