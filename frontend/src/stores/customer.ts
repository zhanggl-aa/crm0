import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  list as listApi,
  get as getApi,
  create as createApi,
  update as updateApi,
  remove as removeApi,
  getInsights as getInsightsApi,
  type Customer,
  type CustomerListParams,
  type CreateCustomerData,
  type UpdateCustomerData,
  type CustomerInsight
} from '../api/customer'

export const useCustomerStore = defineStore('customer', () => {
  const customers = ref<Customer[]>([])
  const total = ref(0)
  const currentCustomer = ref<Customer | null>(null)
  const currentInsight = ref<CustomerInsight | null>(null)
  const loading = ref(false)

  async function fetchCustomers(params: CustomerListParams) {
    loading.value = true
    try {
      const res = await listApi(params)
      customers.value = res.items
      total.value = res.total
    } finally {
      loading.value = false
    }
  }

  async function fetchCustomer(id: string) {
    loading.value = true
    try {
      currentCustomer.value = await getApi(id)
    } finally {
      loading.value = false
    }
  }

  async function fetchInsights(id: string) {
    loading.value = true
    try {
      currentInsight.value = await getInsightsApi(id)
      if (currentInsight.value) {
        currentCustomer.value = currentInsight.value.customer
      }
    } finally {
      loading.value = false
    }
  }

  async function createCustomer(data: CreateCustomerData) {
    const customer = await createApi(data)
    customers.value.unshift(customer)
    total.value += 1
    return customer
  }

  async function updateCustomer(id: string, data: UpdateCustomerData) {
    const customer = await updateApi(id, data)
    const index = customers.value.findIndex((c) => c.id === id)
    if (index !== -1) {
      customers.value[index] = customer
    }
    if (currentCustomer.value && currentCustomer.value.id === id) {
      currentCustomer.value = customer
    }
    return customer
  }

  async function deleteCustomer(id: string) {
    await removeApi(id)
    customers.value = customers.value.filter((c) => c.id !== id)
    total.value -= 1
    if (currentCustomer.value && currentCustomer.value.id === id) {
      currentCustomer.value = null
    }
  }

  return {
    customers,
    total,
    currentCustomer,
    currentInsight,
    loading,
    fetchCustomers,
    fetchCustomer,
    fetchInsights,
    createCustomer,
    updateCustomer,
    deleteCustomer
  }
})
