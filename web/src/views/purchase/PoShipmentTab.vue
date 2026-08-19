<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import type { PurchaseOrder, PurchaseOrderItem } from '../../api/purchase'
import {
  fetchShipments, createShipment, updateShipmentStatus, deleteShipment,
  syncShipmentsFromOrders, splitPurchaseOrderItem,
  fetchAttachments, createAttachment,
  SHIPMENT_STATUS_MAP, type Shipment, type Attachment,
} from '../../api/poTracking'
import { fetchOrder, formatOrderReceiverAddress, shipCallbackTip, shipOrder } from '../../api/order'
import { EXPRESS_COMPANIES, findExpressCompany } from '../../constants/expressCompanies'
import ScanImageUpload from '../../components/ScanImageUpload.vue'

const props = defineProps<{ poId: number; po: PurchaseOrder; readonly: boolean }>()
const emit = defineEmits<{ refresh: [] }>()

const isDropship = computed(() => props.po.fulfillmentType === 'dropship')

const loading = ref(false)
const syncing = ref(false)
const callbacking = ref(false)
const saving = ref(false)
const list = ref<Shipment[]>([])
const attachments = ref<Attachment[]>([])
const dialogVisible = ref(false)
const callbackVisible = ref(false)
const splitVisible = ref(false)
const splitSaving = ref(false)
const splitParent = ref<LinePick | null>(null)
const splitLines = ref<{ skuName: string; qty: number; shipPlanLineId?: number }[]>([])
const photoVisible = ref(false)
const photoSaving = ref(false)
const photoTarget = ref<Shipment | null>(null)
const photoUrls = ref<string[]>([])
const pendingPhotoUrls = ref<string[]>([])
const loadingAddr = ref(false)
const form = ref({
  carrierCode: '',
  carrierName: '',
  trackingNo: '',
  expectedArrivalDate: '',
  receiverName: '',
  receiverPhone: '',
  receiverAddress: '',
  remark: '',
})
const callbackForm = ref({
  refSoId: 0,
  refOrderNo: '',
  carrierCode: '',
  carrierName: '',
  trackingNo: '',
  remark: '',
})

function onCarrierChange(code: string, target: 'form' | 'callback') {
  const hit = findExpressCompany(code)
  const name = hit?.name || code || ''
  const resolvedCode = hit?.code || code || ''
  if (target === 'form') {
    form.value.carrierCode = resolvedCode
    form.value.carrierName = name
  } else {
    callbackForm.value.carrierCode = resolvedCode
    callbackForm.value.carrierName = name
  }
}

function onFormCarrierChange(v: string | number | null | undefined) {
  onCarrierChange(String(v || ''), 'form')
}

function onCallbackCarrierChange(v: string | number | null | undefined) {
  onCarrierChange(String(v || ''), 'callback')
}

function fileNameFromUrl(url: string, fallback = '物流照片.jpg') {
  try {
    const path = url.split('?')[0]
    const name = path.split('/').pop() || ''
    return decodeURIComponent(name) || fallback
  } catch {
    return fallback
  }
}

const photosByShipment = computed(() => {
  const map = new Map<number, Attachment[]>()
  for (const a of attachments.value) {
    if (a.fileType !== 'shipment_photo' || !a.shipmentId) continue
    const arr = map.get(a.shipmentId) || []
    arr.push(a)
    map.set(a.shipmentId, arr)
  }
  return map
})

async function bindShipmentPhotos(shipmentId: number, urls: string[]) {
  for (const url of urls) {
    await createAttachment(props.poId, {
      fileType: 'shipment_photo',
      fileName: fileNameFromUrl(url),
      fileUrl: url,
      shipmentId,
      remark: '发货记录/物流单号照片',
    })
  }
}

interface LinePick {
  poItemId: number
  productName: string
  skuCode: string
  skuSpecs: string
  picUrl?: string
  qty: number
  shippedQty: number
  remaining: number
  selected: boolean
  shipQty: number
  refSoId?: number
  refOrderNo?: string
  refOrderItemId?: number
  parentPoItemId?: number
  splitKind?: string
  shipPlanLineId?: number
  shippable: boolean
  isSplitChild: boolean
  isSplitParent: boolean
}

interface SalesOrderGroup {
  key: string
  refSoId: number
  refOrderNo: string
  /** 未挂订单中心销售单（手工代发） */
  unlinked: boolean
  lines: LinePick[]
  remainingQty: number
  receiverHint: string
  shipStatusHint: string
}

const linePicks = ref<LinePick[]>([])
const soGroups = ref<SalesOrderGroup[]>([])
const activeGroupKey = ref('')
const addressHint = ref('')

const itemLabelMap = computed(() => {
  const map = new Map<number, PurchaseOrderItem>()
  for (const it of props.po.items || []) {
    if (it.id) map.set(it.id, it)
  }
  return map
})

function shippedQtyByItem(): Map<number, number> {
  const map = new Map<number, number>()
  for (const sh of list.value) {
    for (const it of sh.items || []) {
      map.set(it.poItemId, (map.get(it.poItemId) || 0) + it.qty)
    }
  }
  return map
}

function isSplitChildItem(it: PurchaseOrderItem) {
  return !!(it.splitKind || (it.parentPoItemId && it.parentPoItemId > 0))
}

function isShippableItem(it: PurchaseOrderItem, all: PurchaseOrderItem[]) {
  if (it.cancelled) return false
  if (isSplitChildItem(it)) return true
  const hasFull = all.some((x) => !x.cancelled && x.splitKind === 'full')
  if (hasFull) return false
  return !all.some((x) => !x.cancelled && x.splitKind === 'partial' && x.parentPoItemId === it.id)
}

function rebuildLinePicks() {
  const shipped = shippedQtyByItem()
  const all = (props.po.items || []).filter((it) => it.id && !it.cancelled)
  linePicks.value = all.map((it) => {
    const shippedQty = shipped.get(it.id!) || 0
    const remaining = Math.max(0, it.qty - shippedQty)
    const shippable = isShippableItem(it, all)
    const isSplitChild = isSplitChildItem(it)
    const isSplitParent = !isSplitChild && all.some((x) => x.splitKind === 'partial' && x.parentPoItemId === it.id)
    return {
      poItemId: it.id!,
      productName: it.productName || '—',
      skuCode: it.skuCode || '',
      skuSpecs: it.skuSpecs || '',
      picUrl: it.picUrl,
      qty: it.qty,
      shippedQty,
      remaining,
      selected: false,
      shipQty: remaining > 0 ? remaining : 1,
      refSoId: it.refSoId || 0,
      refOrderNo: it.refOrderNo || '',
      refOrderItemId: it.refOrderItemId || 0,
      parentPoItemId: it.parentPoItemId || 0,
      splitKind: it.splitKind || '',
      shipPlanLineId: it.shipPlanLineId || 0,
      shippable,
      isSplitChild,
      isSplitParent,
    }
  })
}

/** 销售单展示名：优先单号，其次订单中心 ID；都没有则「未关联」（手工代发） */
function formatSalesOrderLabel(orderNo: string, soId: number): { text: string; unlinked: boolean } {
  const no = (orderNo || '').trim()
  if (no) return { text: no, unlinked: false }
  if (soId > 0) return { text: `订单中心 #${soId}`, unlinked: false }
  return { text: '未关联', unlinked: true }
}

function rebuildSoGroups() {
  rebuildLinePicks()
  const headerSoId = Number(props.po.refSoId || 0)
  const headerOrderNo = (props.po.refTraceId || '').trim()
  const map = new Map<string, SalesOrderGroup>()
  for (const line of linePicks.value) {
    // 入仓按明细行展示；代发按销售单聚合
    const orderNo = isDropship.value
      ? ((line.refOrderNo || '').trim() || (headerOrderNo.includes(',') ? '' : headerOrderNo))
      : ''
    const soId = isDropship.value ? (line.refSoId || headerSoId || 0) : 0
    const key = isDropship.value
      ? (soId > 0 ? `id:${soId}` : orderNo ? `no:${orderNo}` : `item:${line.poItemId}`)
      : `item:${line.poItemId}`
    let g = map.get(key)
    if (!g) {
      const label = formatSalesOrderLabel(orderNo, soId)
      g = {
        key,
        refSoId: soId,
        refOrderNo: label.text,
        unlinked: label.unlinked,
        lines: [],
        remainingQty: 0,
        receiverHint: '',
        shipStatusHint: '',
      }
      map.set(key, g)
    }
    g.lines.push(line)
    if (line.shippable) g.remainingQty += line.remaining
  }
  const groups = [...map.values()]
  for (const g of groups) {
    const shippedAll = g.lines.every((l) => l.remaining <= 0)
    const partial = g.lines.some((l) => l.shippedQty > 0) && !shippedAll
    g.shipStatusHint = shippedAll ? '已登记物流' : partial ? '部分发货' : '待发货'
  }
  soGroups.value = groups
}

/** 待发明细树：根行 + └ 拆分子行（对齐订单中心展示） */
const soGroupRows = computed(() => {
  const rows: {
    key: string
    group: SalesOrderGroup
    line: LinePick
    splitParent?: LinePick
    showSalesOrder: boolean
    showSplitBtn: boolean
    showCallbackBtn: boolean
    lineStatus: string
    isSplitChild: boolean
    isSplitParent: boolean
  }[] = []
  for (const g of soGroups.value) {
    const childrenByParent = new Map<number, LinePick[]>()
    const roots: LinePick[] = []
    for (const line of g.lines) {
      if (line.isSplitChild && line.parentPoItemId) {
        const list = childrenByParent.get(line.parentPoItemId) || []
        list.push(line)
        childrenByParent.set(line.parentPoItemId, list)
        continue
      }
      roots.push(line)
    }
    const pushLine = (
      line: LinePick,
      opts: {
        isChild: boolean
        isParent: boolean
        splitParent?: LinePick
        showSalesOrder: boolean
        showSplitBtn: boolean
        showCallbackBtn: boolean
        kids?: LinePick[]
      },
    ) => {
      const kids = opts.kids || []
      const done = line.shippable
        ? line.remaining <= 0
        : (kids.length > 0 && kids.every((k) => k.remaining <= 0))
      const partial = line.shippable
        ? (line.shippedQty > 0 && line.remaining > 0)
        : kids.some((k) => k.shippedQty > 0) && !kids.every((k) => k.remaining <= 0)
      let status = '待发货'
      if (!line.shippable && kids.length) status = done ? '已登记物流' : partial ? '部分发货' : '已拆分'
      else if (done) status = '已登记物流'
      else if (partial) status = '部分发货'
      rows.push({
        key: `${g.key}:${line.poItemId}`,
        group: g,
        line,
        splitParent: opts.splitParent,
        showSalesOrder: opts.showSalesOrder,
        showSplitBtn: opts.showSplitBtn,
        showCallbackBtn: opts.showCallbackBtn,
        lineStatus: status,
        isSplitChild: opts.isChild,
        isSplitParent: opts.isParent,
      })
    }
    for (const root of roots) {
      const kids = childrenByParent.get(root.poItemId) || []
      if (kids.length > 0) {
        // 树形：父行（已拆分、不可发）+ └ 子规格（可发）
        pushLine(root, {
          isChild: false,
          isParent: true,
          showSalesOrder: true,
          showSplitBtn: true,
          showCallbackBtn: true,
          kids,
        })
        for (const ch of kids) {
          pushLine(ch, {
            isChild: true,
            isParent: false,
            splitParent: root,
            showSalesOrder: false,
            showSplitBtn: false,
            showCallbackBtn: false,
            kids: [],
          })
        }
      } else {
        pushLine(root, {
          isChild: false,
          isParent: false,
          showSalesOrder: true,
          showSplitBtn: true,
          showCallbackBtn: true,
        })
      }
    }
  }
  return rows
})

function formatSpecLabel(specs?: string, qty?: number) {
  const s = (specs || '').trim() || '—'
  return qty != null ? `${s} ×${qty}` : s
}

function shipmentSpecText(row: Shipment) {
  const items = row.items || []
  if (!items.length) return '—'
  return items
    .map((it) => {
      const poItem = itemLabelMap.value.get(it.poItemId)
      return formatSpecLabel(poItem?.skuSpecs, it.qty)
    })
    .join('；')
}

function shipmentSalesOrders(row: Shipment) {
  const nos = new Set<string>()
  const headerOrderNo = (props.po.refTraceId || '').trim()
  const headerSoId = Number(props.po.refSoId || 0)
  for (const it of row.items || []) {
    const poItem = itemLabelMap.value.get(it.poItemId)
    const no = poItem?.refOrderNo?.trim()
    if (no) nos.add(no)
    else if (poItem?.refSoId) nos.add(formatSalesOrderLabel('', poItem.refSoId).text)
  }
  if (!nos.size && headerOrderNo && !headerOrderNo.includes(',')) nos.add(headerOrderNo)
  else if (!nos.size && headerSoId) nos.add(formatSalesOrderLabel('', headerSoId).text)
  return [...nos].join('、') || '未关联'
}

async function loadData() {
  loading.value = true
  try {
    const [shipments, files] = await Promise.all([
      fetchShipments(props.poId),
      fetchAttachments(props.poId),
    ])
    list.value = shipments
    attachments.value = files
    rebuildSoGroups()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
watch(() => props.po.items, () => {
  rebuildSoGroups()
})
watch(isDropship, () => {
  rebuildSoGroups()
})

function resetForm() {
  form.value = {
    carrierCode: '', carrierName: '', trackingNo: '', expectedArrivalDate: '',
    receiverName: '', receiverPhone: '', receiverAddress: '', remark: '',
  }
  pendingPhotoUrls.value = []
  addressHint.value = ''
}

async function fillReceiverFromOrder(soId: number) {
  if (!soId) {
    addressHint.value = '该明细未关联销售单 ID，请手工填写收件人'
    return
  }
  loadingAddr.value = true
  addressHint.value = ''
  try {
    const order = await fetchOrder(soId)
    const addr = order.address
    const name = addr?.name?.trim() || order.buyerName?.trim() || ''
    const phone = addr?.phone?.trim() || order.buyerPhone?.trim() || ''
    const address = formatOrderReceiverAddress(addr)
    form.value.receiverName = name
    form.value.receiverPhone = phone
    form.value.receiverAddress = address
    if (!name && !phone && !address) {
      addressHint.value = '订单中心暂无明文收件人（可能已脱敏），请到订单中心解密后重试，或手工填写'
    } else {
      addressHint.value = `已从订单 ${order.orderNo} 带入收件人`
    }
  } catch (e) {
    addressHint.value = (e as Error).message || '拉取订单收件人失败，请手工填写'
  } finally {
    loadingAddr.value = false
  }
}

async function openCreateDropship(group: SalesOrderGroup, focus?: LinePick) {
  if (group.remainingQty <= 0) {
    ElMessage.warning('该销售单明细已全部关联物流')
    return
  }
  resetForm()
  activeGroupKey.value = group.key
  const focusId = focus?.shippable ? focus.poItemId : 0
  linePicks.value = group.lines
    .filter((l) => l.shippable)
    .map((l) => ({
      ...l,
      selected: focusId ? l.poItemId === focusId && l.remaining > 0 : l.remaining > 0,
      shipQty: l.remaining > 0 ? l.remaining : 1,
    }))
  dialogVisible.value = true
  await fillReceiverFromOrder(group.refSoId)
}

/** 入仓：按明细行发货（对齐代发操作入口，无销售单/回传） */
function openCreateStockInLine(line: LinePick) {
  if (line.remaining <= 0) {
    ElMessage.warning('该明细已全部关联物流')
    return
  }
  resetForm()
  activeGroupKey.value = `item:${line.poItemId}`
  rebuildLinePicks()
  linePicks.value = linePicks.value.map((l) => ({
    ...l,
    selected: l.poItemId === line.poItemId && l.remaining > 0,
    shipQty: l.remaining > 0 ? l.remaining : 1,
  }))
  addressHint.value = '采购入仓发货：登记供应商发往仓库的物流'
  dialogVisible.value = true
}

async function handleSave() {
  const selected = linePicks.value.filter((l) => l.selected && l.remaining > 0)
  if (!selected.length) {
    ElMessage.warning(isDropship.value ? '该销售单没有可发货明细' : '请选择本物流单对应的商品')
    return
  }
  for (const line of selected) {
    if (!line.shipQty || line.shipQty < 1 || line.shipQty > line.remaining) {
      ElMessage.warning(`「${line.productName}」发货数量需在 1 ~ ${line.remaining} 之间`)
      return
    }
  }
  if (!form.value.trackingNo?.trim()) {
    ElMessage.warning('请填写物流单号')
    return
  }
  if (!form.value.carrierName?.trim() && form.value.carrierCode) {
    onCarrierChange(form.value.carrierCode, 'form')
  }
  if (!form.value.carrierName?.trim()) {
    ElMessage.warning('请选择快递公司')
    return
  }
  saving.value = true
  try {
    const created = await createShipment(props.poId, {
      ...form.value,
      items: selected.map((l) => ({ poItemId: l.poItemId, qty: l.shipQty })),
    })
    if (pendingPhotoUrls.value.length && created?.id) {
      await bindShipmentPhotos(created.id, pendingPhotoUrls.value)
    }

    // 代发：同一弹窗内回传订单中心 → StoreSyncAgent → 快递助手手动发货
    let callbackMsg = ''
    const refSoId = activeGroup.value?.refSoId || selected.find((l) => l.refSoId)?.refSoId
    const trackingNo = form.value.trackingNo.trim()
    if (isDropship.value && refSoId) {
      try {
        const shipItems = selected
          .filter((l) => l.refOrderItemId && l.refOrderItemId > 0)
          .map((l) => ({ orderItemId: l.refOrderItemId!, qty: l.shipQty }))
        const shipped = await shipOrder(refSoId, {
          expressCompany: form.value.carrierName,
          expressNo: trackingNo,
          remark: form.value.remark || `代发采购单 ${props.po.poNo || props.poId} 发货回传`,
          callback: true,
          items: shipItems.length ? shipItems : undefined,
        })
        const tip = shipCallbackTip(shipped, trackingNo)
        if (!tip.ok) {
          await ElMessageBox.alert(tip.message, '回传失败（快递助手/平台原始报错）', {
            type: 'error',
            confirmButtonText: '知道了',
          })
          dialogVisible.value = false
          await loadData()
          emit('refresh')
          return
        }
        callbackMsg = tip.message ? `，${tip.message}` : '，并已回传电商平台'
        await handleSyncFromOrders(refSoId)
      } catch (e) {
        ElMessage.warning(`本地发货已保存，但回传失败：${(e as Error).message || '未知错误'}`)
        dialogVisible.value = false
        await loadData()
        emit('refresh')
        return
      }
    }

    ElMessage.success(`已添加发货批次${callbackMsg}`)
    dialogVisible.value = false
    await loadData()
    emit('refresh')
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  } finally {
    saving.value = false
  }
}

function openUploadPhotos(row: Shipment) {
  photoTarget.value = row
  photoUrls.value = []
  photoVisible.value = true
}

async function handleSavePhotos() {
  if (!photoTarget.value) return
  if (!photoUrls.value.length) {
    ElMessage.warning('请先上传照片')
    return
  }
  photoSaving.value = true
  try {
    await bindShipmentPhotos(photoTarget.value.id, photoUrls.value)
    ElMessage.success('已上传物流照片')
    photoVisible.value = false
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    photoSaving.value = false
  }
}

async function openCallback(group: SalesOrderGroup) {
  if (!group.refSoId) {
    ElMessage.warning('该销售单缺少订单中心 ID，无法回传')
    return
  }
  // 优先带出本销售单已登记的物流，避免再填一次
  const itemIds = new Set(group.lines.map((l) => l.poItemId))
  const withTracking = list.value.find((s) => {
    if (!s.trackingNo?.trim()) return false
    return (s.items || []).some((it) => itemIds.has(it.poItemId))
  }) || list.value.find((s) => !!s.trackingNo?.trim())

  callbackForm.value = {
    refSoId: group.refSoId,
    refOrderNo: group.refOrderNo,
    carrierCode: withTracking?.carrierCode || '',
    carrierName: withTracking?.carrierName || '',
    trackingNo: withTracking?.trackingNo || '',
    remark: '',
  }
  if (callbackForm.value.carrierCode && !callbackForm.value.carrierName) {
    onCarrierChange(callbackForm.value.carrierCode, 'callback')
  }
  callbackVisible.value = true
}

async function handleCallbackShip() {
  if (!callbackForm.value.carrierName?.trim() && callbackForm.value.carrierCode) {
    onCarrierChange(callbackForm.value.carrierCode, 'callback')
  }
  if (!callbackForm.value.carrierName?.trim()) {
    ElMessage.warning('请选择快递公司')
    return
  }
  if (!callbackForm.value.trackingNo?.trim()) {
    ElMessage.warning('请填写物流单号')
    return
  }
  callbacking.value = true
  try {
    const expressNo = callbackForm.value.trackingNo.trim()
    const shipped = await shipOrder(callbackForm.value.refSoId, {
      expressCompany: callbackForm.value.carrierName,
      expressNo,
      remark: callbackForm.value.remark || `代发采购单 ${props.po.poNo || props.poId} 回传`,
      callback: true,
    })
    const tip = shipCallbackTip(shipped, expressNo)
    if (!tip.ok) {
      await ElMessageBox.alert(tip.message, '回传失败（快递助手/平台原始报错）', {
        type: 'error',
        confirmButtonText: '知道了',
      })
    } else {
      ElMessage.success(tip.message || '已回传订单中心')
      callbackVisible.value = false
    }
    // 回传后补同步本采购单物流展示（成功或失败都刷新本地展示）
    await handleSyncFromOrders(callbackForm.value.refSoId)
  } catch (e) {
    ElMessage.error((e as Error).message || '回传失败')
  } finally {
    callbacking.value = false
  }
}

async function handleSyncFromOrders(refSoId?: number) {
  syncing.value = true
  try {
    const res = await syncShipmentsFromOrders(props.poId, refSoId)
    const parts = [
      res.created ? `新建 ${res.created}` : '',
      res.updated ? `更新 ${res.updated}` : '',
      res.skipped ? `跳过 ${res.skipped}` : '',
    ].filter(Boolean)
    ElMessage.success(parts.length ? `同步完成：${parts.join('，')}` : '同步完成')
    if (res.errors?.length) {
      ElMessage.warning(res.errors.slice(0, 3).join('；'))
    }
    await loadData()
    emit('refresh')
  } catch (e) {
    ElMessage.error((e as Error).message || '同步物流失败')
  } finally {
    syncing.value = false
  }
}

async function changeStatus(row: Shipment, status: string) {
  try {
    await updateShipmentStatus(props.poId, row.id, status)
    ElMessage.success('状态已更新')
    await loadData()
    emit('refresh')
  } catch (e) {
    ElMessage.error((e as Error).message || '更新失败')
  }
}

async function handleDelete(row: Shipment) {
  try {
    await ElMessageBox.confirm('确定删除此发货批次？', '确认')
  } catch {
    return
  }
  try {
    await deleteShipment(props.poId, row.id)
    ElMessage.success('已删除')
    await loadData()
    emit('refresh')
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}


function openSplit(line: LinePick, parent?: LinePick) {
  const target = parent || (line.isSplitChild && line.parentPoItemId
    ? linePicks.value.find((l) => l.poItemId === line.parentPoItemId)
    : line)
  if (!target || target.isSplitChild) {
    ElMessage.warning('请在父商品上编辑拆分')
    return
  }
  splitParent.value = target
  const kids = linePicks.value.filter((l) => l.parentPoItemId === target.poItemId && l.isSplitChild)
  if (kids.length) {
    splitLines.value = kids.map((k) => ({
      skuName: k.skuSpecs || k.productName,
      qty: k.qty,
      shipPlanLineId: k.shipPlanLineId || undefined,
    }))
  } else {
    // 自由添加：默认空表，由用户点「+ 添加规格」
    splitLines.value = []
  }
  splitVisible.value = true
}

function addSplitLine() {
  splitLines.value.push({ skuName: '', qty: 1 })
}

function removeSplitLine(idx: number) {
  splitLines.value.splice(idx, 1)
}

async function handleSaveSplit() {
  if (!splitParent.value) return
  const clearing = splitLines.value.length === 0
  if (!clearing) {
    for (let i = 0; i < splitLines.value.length; i++) {
      const row = splitLines.value[i]
      if (!row.skuName?.trim()) {
        ElMessage.warning(`第 ${i + 1} 行请填写规格名称`)
        return
      }
      if (!row.qty || row.qty < 1) {
        ElMessage.warning(`第 ${i + 1} 行数量须大于 0`)
        return
      }
    }
  } else {
    try {
      await ElMessageBox.confirm(
        '将删除该商品全部未发拆分行，恢复为未拆分。已发拆分行不可取消。确认继续？',
        '取消拆分',
        { type: 'warning', confirmButtonText: '确认取消拆分', cancelButtonText: '返回' },
      )
    } catch {
      return
    }
  }
  splitSaving.value = true
  try {
    const res = await splitPurchaseOrderItem(
      props.poId,
      splitParent.value.poItemId,
      splitLines.value.map((l) => ({
        skuName: l.skuName.trim(),
        qty: l.qty,
        shipPlanLineId: l.shipPlanLineId,
      })),
    )
    if (res.syncWarning) {
      ElMessage.warning(`${clearing ? '已取消拆分' : '拆分已保存'}：${res.syncWarning}`)
    } else if (clearing) {
      ElMessage.success(res.syncedToOrderCore ? '已取消拆分并同步订单中心' : '已取消拆分')
    } else if (res.syncedToOrderCore) {
      ElMessage.success('已拆分并同步订单中心')
    } else {
      ElMessage.success('已拆分')
    }
    splitVisible.value = false
    emit('refresh')
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || (clearing ? '取消拆分失败' : '拆分失败'))
  } finally {
    splitSaving.value = false
  }
}

const dialogTitle = computed(() =>
  isDropship.value ? '按销售单发货' : '发货',
)

const activeGroup = computed(() =>
  soGroups.value.find((g) => g.key === activeGroupKey.value) || null,
)
</script>

<template>
  <div v-loading="loading">
    <div v-if="isDropship && !readonly" class="toolbar">
      <el-button type="primary" plain :loading="syncing" @click="handleSyncFromOrders()">同步物流</el-button>
      <span class="hint">代发「发货」会同时回传订单中心与快递助手；也可「回传单号」单独补传</span>
    </div>

    <!-- 待发明细：代发含销售单/回传；入仓仅规格+待发+状态+发货 -->
    <el-table :data="soGroupRows" border stripe class="so-group-table" row-key="key">
      <el-table-column v-if="isDropship" label="销售单" width="160" show-overflow-tooltip>
        <template #default="{ row }">
          <template v-if="row.showSalesOrder">
            <el-tag v-if="row.group.unlinked" type="info" effect="plain" size="small">未关联</el-tag>
            <span v-else>{{ row.group.refOrderNo }}</span>
          </template>
          <span v-else class="muted">—</span>
        </template>
      </el-table-column>
      <el-table-column label="规格" min-width="240" show-overflow-tooltip>
        <template #default="{ row }">
          <div class="spec-cell" :class="{ child: row.isSplitChild }">
            <span v-if="row.isSplitChild" class="tree-prefix">└ </span>
            <span>{{ formatSpecLabel(row.line.skuSpecs || row.line.productName, row.line.qty) }}</span>
            <el-tag v-if="row.isSplitChild" size="small" type="warning" class="split-tag">拆分</el-tag>
            <el-tag v-else-if="row.isSplitParent" size="small" type="info" class="split-tag">已拆分</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="待发" width="80" align="center">
        <template #default="{ row }">
          <span v-if="row.line.shippable" :class="{ muted: row.line.remaining <= 0 }">{{ row.line.remaining }}</span>
          <span v-else class="muted">—</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">{{ row.lineStatus }}</template>
      </el-table-column>
      <el-table-column v-if="!readonly" label="操作" :width="isDropship ? 220 : 140" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="row.showSplitBtn"
            type="warning"
            link
            @click="openSplit(row.line, row.splitParent)"
          >
            {{ row.isSplitParent ? '编辑拆分' : '拆分' }}
          </el-button>
          <el-button
            v-if="isDropship && row.line.shippable"
            type="primary"
            link
            :disabled="row.line.remaining <= 0"
            @click="openCreateDropship(row.group, row.line)"
          >
            发货
          </el-button>
          <el-button
            v-else-if="!isDropship && row.line.shippable"
            type="primary"
            link
            :disabled="row.line.remaining <= 0"
            @click="openCreateStockInLine(row.line)"
          >
            发货
          </el-button>
          <el-button
            v-if="isDropship && row.showCallbackBtn"
            type="success"
            link
            :disabled="!row.group.refSoId"
            @click="openCallback(row.group)"
          >
            回传单号
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <h4 class="section-title">发货批次</h4>

    <el-table :data="list" border stripe>
      <el-table-column prop="shipmentNo" label="批次号" width="140" />
      <el-table-column v-if="isDropship" label="销售单" width="150" show-overflow-tooltip>
        <template #default="{ row }">{{ shipmentSalesOrders(row) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">{{ SHIPMENT_STATUS_MAP[row.status] || row.status }}</template>
      </el-table-column>
      <el-table-column prop="carrierName" label="快递" width="100" />
      <el-table-column prop="trackingNo" label="物流单号" min-width="140" />
      <el-table-column label="发货照片" width="140">
        <template #default="{ row }">
          <div v-if="photosByShipment.get(row.id)?.length" class="shots">
            <el-image
              v-for="a in photosByShipment.get(row.id)"
              :key="a.id"
              :src="a.fileUrl"
              :preview-src-list="(photosByShipment.get(row.id) || []).map((x) => x.fileUrl)"
              fit="cover"
              class="shot"
              preview-teleported
            />
          </div>
          <span v-else class="muted">—</span>
        </template>
      </el-table-column>
      <el-table-column label="对应规格" min-width="220" show-overflow-tooltip>
        <template #default="{ row }">{{ shipmentSpecText(row) }}</template>
      </el-table-column>
      <el-table-column label="收件人" width="100" show-overflow-tooltip>
        <template #default="{ row }">{{ row.receiverName || '—' }}</template>
      </el-table-column>
      <el-table-column prop="expectedArrivalDate" label="预计到货" width="110" />
      <el-table-column prop="shippedAt" label="发货时间" width="150" />
      <el-table-column v-if="!readonly" label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openUploadPhotos(row)">上传照片</el-button>
          <el-button v-if="row.status === 'pending'" link type="primary" @click="changeStatus(row, 'shipped')">已发货</el-button>
          <el-button v-if="row.status === 'shipped'" link type="primary" @click="changeStatus(row, 'in_transit')">运输中</el-button>
          <el-button v-if="row.status === 'in_transit' || row.status === 'shipped'" link type="success" @click="changeStatus(row, 'delivered')">已签收</el-button>
          <el-button link type="danger" :icon="Delete" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="720px">
      <el-form v-loading="loadingAddr" :model="form" label-width="90px">
        <el-form-item v-if="isDropship && activeGroup" label="销售单">
          <el-tag v-if="activeGroup.unlinked" type="info" effect="plain" size="small">未关联</el-tag>
          <span v-else>{{ activeGroup.refOrderNo }}</span>
        </el-form-item>
        <el-form-item label="快递公司" required>
          <el-select
            v-model="form.carrierCode"
            filterable
            allow-create
            default-first-option
            clearable
            placeholder="选择或搜索快递公司"
            style="width: 100%"
            @change="onFormCarrierChange"
          >
            <el-option
              v-for="c in EXPRESS_COMPANIES"
              :key="c.code"
              :label="c.name"
              :value="c.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="物流单号" required>
          <el-input v-model="form.trackingNo" :placeholder="isDropship ? '填写该销售单对应物流单号' : '与下方勾选商品对应'" />
        </el-form-item>
        <el-form-item label="发货照片">
          <ScanImageUpload
            v-model="pendingPhotoUrls"
            subdir="po/shipments"
            tip="本机上传"
            scan-title="手机扫码上传发货/单号照片"
          />
          <div class="hint" style="margin-top: 6px">可上传发货记录、物流面单/单号照片等</div>
        </el-form-item>
        <el-form-item :label="isDropship ? '发货明细' : '发货商品'" required>
          <el-table :data="linePicks" border size="small" max-height="280">
            <el-table-column v-if="!isDropship" width="48" align="center">
              <template #default="{ row }">
                <el-checkbox v-model="row.selected" :disabled="row.remaining <= 0" />
              </template>
            </el-table-column>
            <el-table-column label="图片" width="56" align="center">
              <template #default="{ row }">
                <el-image
                  v-if="row.picUrl"
                  :src="row.picUrl"
                  fit="cover"
                  style="width: 32px; height: 32px; border-radius: 4px"
                />
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="规格" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">{{ formatSpecLabel(row.skuSpecs) }}</template>
            </el-table-column>
            <el-table-column label="商家编码" width="110" show-overflow-tooltip>
              <template #default="{ row }">{{ row.skuCode || '—' }}</template>
            </el-table-column>
            <el-table-column label="可发" width="70" align="center">
              <template #default="{ row }">
                <span :class="{ muted: row.remaining <= 0 }">{{ row.remaining }}/{{ row.qty }}</span>
              </template>
            </el-table-column>
            <el-table-column label="本单数量" width="110" align="center">
              <template #default="{ row }">
                <el-input-number
                  v-model="row.shipQty"
                  :min="1"
                  :max="Math.max(1, row.remaining)"
                  :disabled="!row.selected || row.remaining <= 0"
                  size="small"
                  controls-position="right"
                  style="width: 96px"
                />
              </template>
            </el-table-column>
          </el-table>
          <div v-if="!linePicks.some((l) => l.remaining > 0)" class="hint warn">
            全部明细已关联物流，如需改绑请先删除旧发货批次
          </div>
        </el-form-item>
        <el-form-item label="预计到货">
          <el-date-picker v-model="form.expectedArrivalDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="收货人">
          <el-input v-model="form.receiverName" />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="form.receiverPhone" />
        </el-form-item>
        <el-form-item label="收货地址">
          <el-input v-model="form.receiverAddress" type="textarea" :rows="2" />
          <div v-if="addressHint" class="hint addr-hint">{{ addressHint }}</div>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ isDropship ? '发货并回传' : '保存' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="callbackVisible" title="回传单号" width="480px">
      <el-form :model="callbackForm" label-width="90px">
        <el-form-item label="销售单">
          <span>{{ callbackForm.refOrderNo }}</span>
        </el-form-item>
        <el-form-item label="快递公司" required>
          <el-select
            v-model="callbackForm.carrierCode"
            filterable
            allow-create
            default-first-option
            clearable
            placeholder="选择或搜索快递公司"
            style="width: 100%"
            @change="onCallbackCarrierChange"
          >
            <el-option
              v-for="c in EXPRESS_COMPANIES"
              :key="c.code"
              :label="c.name"
              :value="c.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="物流单号" required>
          <el-input v-model="callbackForm.trackingNo" placeholder="运单号" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="callbackForm.remark" />
        </el-form-item>
        <div class="hint">已自动带出本单物流（如有）；将写入订单中心并回传快递助手手动发货</div>
      </el-form>
      <template #footer>
        <el-button @click="callbackVisible = false">取消</el-button>
        <el-button type="primary" :loading="callbacking" @click="handleCallbackShip">确认回传</el-button>
      </template>
    </el-dialog>


    <el-dialog v-model="splitVisible" title="拆分规格" width="560px">
      <div v-if="splitParent" class="hint" style="margin-bottom: 12px">
        父商品：{{ formatSpecLabel(splitParent.skuSpecs || splitParent.productName, splitParent.qty) }}
        <span v-if="isDropship"> · 保存后同步订单中心拆分明细</span>
      </div>
      <el-table v-if="splitLines.length" :data="splitLines" border size="small">
        <el-table-column label="规格名称" min-width="200">
          <template #default="{ row }">
            <el-input v-model="row.skuName" placeholder="如：红色 / L" />
          </template>
        </el-table-column>
        <el-table-column label="数量" width="120" align="center">
          <template #default="{ row }">
            <el-input-number v-model="row.qty" :min="1" size="small" controls-position="right" style="width: 100px" />
          </template>
        </el-table-column>
        <el-table-column label="" width="70" align="center">
          <template #default="{ $index }">
            <el-button link type="danger" @click="removeSplitLine($index)">删</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-else class="hint" style="margin: 8px 0 4px">
        暂无拆分行。保存将取消拆分并恢复为未拆分；或点击下方添加规格。
      </div>
      <el-button class="add-split" type="primary" link @click="addSplitLine">+ 添加规格</el-button>
      <template #footer>
        <el-button @click="splitVisible = false">取消</el-button>
        <el-button type="primary" :loading="splitSaving" @click="handleSaveSplit">
          {{ splitLines.length ? '保存拆分' : '取消拆分' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="photoVisible" title="上传发货/物流照片" width="520px">
      <div v-if="photoTarget" class="hint" style="margin-bottom: 12px">
        批次 {{ photoTarget.shipmentNo }} · {{ photoTarget.carrierName || '—' }} {{ photoTarget.trackingNo || '' }}
      </div>
      <ScanImageUpload
        v-model="photoUrls"
        subdir="po/shipments"
        tip="本机上传"
        scan-title="手机扫码上传发货/单号照片"
      />
      <template #footer>
        <el-button @click="photoVisible = false">取消</el-button>
        <el-button type="primary" :loading="photoSaving" @click="handleSavePhotos">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.hint.warn {
  margin-top: 8px;
  color: var(--el-color-warning);
}
.addr-hint {
  margin-top: 6px;
}
.section-title {
  margin: 16px 0 10px;
  font-size: 14px;
  font-weight: 600;
}
.so-group-table {
  margin-bottom: 8px;
}
.sub {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.muted {
  color: var(--el-text-color-placeholder);
}
.shots {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.shot {
  width: 36px;
  height: 36px;
  border-radius: 4px;
}
.spec-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}
.spec-cell.child {
  padding-left: 8px;
  color: #475569;
}
.tree-prefix {
  color: #8f959e;
  margin-right: 2px;
  flex-shrink: 0;
}
.split-tag {
  flex-shrink: 0;
  margin-left: 2px;
  vertical-align: middle;
}
.add-split {
  margin-top: 10px;
}
</style>
