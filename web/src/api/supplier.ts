import client, { unwrap, type PageData } from './client'

export interface SupplierCategory {
  id: number
  name: string
  parentId?: number
  sort?: number
  status: number
  remark?: string
}

export interface Supplier {
  id: number
  categoryId?: number
  categoryName?: string
  code: string
  name: string
  shortName?: string
  status: number
  buyerName?: string
  cutOffTime?: string
  arrivalDays?: number
  paymentDays?: number
  contactName?: string
  address?: string
  officePhone?: string
  mobile?: string
  phone?: string
  wangwangId?: string
  qq?: string
  email?: string
  website?: string
  remark?: string
  defaultPaymentTerms?: string
  bankName?: string
  bankAccount?: string
  accountName?: string
  createdAt?: string
}

export interface SupplierAddress {
  id: number
  supplierId: number
  label: string
  contactName?: string
  phone?: string
  province?: string
  city?: string
  district?: string
  address?: string
  isDefault: boolean
  status: number
}

export interface SkuOffer {
  id: number
  skuId: number
  supplierId: number
  supplierName?: string
  supplierCode?: string
  supplierSkuCode?: string
  supplyPrice: number
  currency: string
  minOrderQty: number
  leadTimeDays: number
  shipFromAddressId?: number
  shipFromLabel?: string
  shipFromCity?: string
  supportsDropship: boolean
  supportsSelfStock: boolean
  isPrimary: boolean
  priority: number
  status: number
  remark?: string
}

export async function fetchSupplierCategories() {
  return unwrap<SupplierCategory[]>(await client.get('/supplier-categories'))
}

export async function createSupplierCategory(data: Partial<SupplierCategory>) {
  return unwrap<SupplierCategory>(await client.post('/supplier-categories', data))
}

export async function updateSupplierCategory(id: number, data: Partial<SupplierCategory>) {
  return unwrap<SupplierCategory>(await client.put(`/supplier-categories/${id}`, data))
}

export async function deleteSupplierCategory(id: number) {
  return unwrap(await client.delete(`/supplier-categories/${id}`))
}

export async function fetchSuppliers(params?: {
  keyword?: string
  categoryId?: number
  page?: number
  pageSize?: number
}) {
  const { keyword, categoryId, page = 1, pageSize = 20 } = params ?? {}
  const res = await client.get('/suppliers', {
    params: { keyword, categoryId: categoryId || undefined, page, pageSize },
  })
  return unwrap<PageData<Supplier>>(res)
}

export async function fetchSupplier(id: number) {
  return unwrap<Supplier>(await client.get(`/suppliers/${id}`))
}

export async function createSupplier(data: Partial<Supplier>) {
  return unwrap<Supplier>(await client.post('/suppliers', data))
}

export async function updateSupplier(id: number, data: Partial<Supplier>) {
  return unwrap<Supplier>(await client.put(`/suppliers/${id}`, data))
}

export async function deleteSupplier(id: number) {
  return unwrap(await client.delete(`/suppliers/${id}`))
}

export async function fetchSupplierAddresses(supplierId: number) {
  return unwrap<SupplierAddress[]>(await client.get(`/suppliers/${supplierId}/addresses`))
}

export async function createSupplierAddress(supplierId: number, data: Partial<SupplierAddress>) {
  return unwrap<SupplierAddress>(await client.post(`/suppliers/${supplierId}/addresses`, data))
}

export async function updateSupplierAddress(supplierId: number, addressId: number, data: Partial<SupplierAddress>) {
  return unwrap<SupplierAddress>(await client.put(`/suppliers/${supplierId}/addresses/${addressId}`, data))
}

export async function deleteSupplierAddress(supplierId: number, addressId: number) {
  return unwrap(await client.delete(`/suppliers/${supplierId}/addresses/${addressId}`))
}

export async function fetchSkuOffers(params: { skuId?: number; supplierId?: number; page?: number; pageSize?: number }) {
  const res = await client.get('/sku-offers', { params })
  return unwrap<PageData<SkuOffer>>(res)
}

export async function createSkuOffer(data: Partial<SkuOffer>) {
  return unwrap<SkuOffer>(await client.post('/sku-offers', data))
}

export async function updateSkuOffer(id: number, data: Partial<SkuOffer>) {
  return unwrap<SkuOffer>(await client.put(`/sku-offers/${id}`, data))
}

export async function deleteSkuOffer(id: number) {
  return unwrap(await client.delete(`/sku-offers/${id}`))
}

/** Display mobile with fallback to legacy phone field */
export function supplierMobile(row: Supplier): string {
  return row.mobile || row.phone || ''
}
