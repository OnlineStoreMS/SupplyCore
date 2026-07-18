import client, { unwrap, type PageData } from './client'

export interface Warehouse {
  id: number
  code: string
  name: string
  status: number
  isDefault: number
}

export interface WarehouseLocation {
  id: number
  warehouseId: number
  code: string
  name: string
  status: number
}

export async function listWarehouses(params?: { keyword?: string; page?: number; pageSize?: number }) {
  const res = await client.get('/warehouses', { params })
  return unwrap<PageData<Warehouse>>(res)
}

export async function listLocations(warehouseId: number, params?: { page?: number; pageSize?: number }) {
  const res = await client.get(`/warehouses/${warehouseId}/locations`, { params })
  return unwrap<PageData<WarehouseLocation>>(res)
}
