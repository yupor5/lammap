import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

export interface Product {
  id: number
  name: string
  sku: string
  category: string
  description: string
  material: string
  size: string
  color: string
  process: string
  packaging: string
  price: number
  moq: number
  leadTime: string
  paymentTerms: string
  attachments: number
  createdAt: string
  updatedAt: string
}

export const useProductStore = defineStore('product', () => {
  const products = ref<Product[]>([])
  const total = ref(0)
  const loading = ref(false)

  async function fetchProducts(params?: { search?: string; page?: number; pageSize?: number }) {
    loading.value = true
    try {
      const res: any = await request.get('/products', { params })
      products.value = res.data.items || []
      total.value = res.data.total || 0
      return res.data
    } finally {
      loading.value = false
    }
  }

  async function fetchProduct(id: number) {
    const res: any = await request.get(`/products/${id}`)
    return res.data as Product
  }

  async function createProduct(data: Partial<Product>) {
    const res: any = await request.post('/products', data)
    return res.data as Product
  }

  async function updateProduct(id: number, data: Partial<Product>) {
    const res: any = await request.put(`/products/${id}`, data)
    return res.data as Product
  }

  async function deleteProduct(id: number) {
    await request.delete(`/products/${id}`)
    products.value = products.value.filter((p) => p.id !== id)
  }

  async function importProducts(data: Partial<Product>[]) {
    const res: any = await request.post('/products/import', data)
    return res.data
  }

  async function matchProducts(params: { productName?: string; material?: string; size?: string; color?: string; model?: string }) {
    const res: any = await request.post('/products/match', params)
    return res.data
  }

  return { products, total, loading, fetchProducts, fetchProduct, createProduct, updateProduct, deleteProduct, importProducts, matchProducts }
})
