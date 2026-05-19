<template>
  <div class="customers-page">
    <div class="page-header">
      <h2>{{ t('customers.title') }}</h2>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon> {{ t('customers.addCustomer') }}
      </el-button>
    </div>

    <div class="toolbar">
      <el-input v-model="search" :placeholder="t('customers.searchPlaceholder')" clearable style="width: 300px" @clear="loadCustomers" @keyup.enter="loadCustomers">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-select v-model="statusFilter" :placeholder="t('customers.statusFilter')" clearable style="width: 150px" @change="loadCustomers">
        <el-option :label="t('common.active')" value="active" />
        <el-option :label="t('common.inactive')" value="inactive" />
        <el-option :label="t('common.churned')" value="churned" />
      </el-select>
    </div>

    <el-table :data="customerStore.customers" v-loading="customerStore.loading" stripe>
      <el-table-column prop="name" :label="t('common.name')" min-width="120" />
      <el-table-column prop="email" :label="t('common.email')" min-width="180" />
      <el-table-column prop="company" :label="t('common.company')" min-width="150" />
      <el-table-column prop="status" :label="t('common.status')" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="acquired_channel" :label="t('customers.acquiredChannel')" width="120" />
      <el-table-column :label="t('common.actions')" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row.id)">{{ t('common.detail') }}</el-button>
          <el-button link type="warning" @click="editCustomer(row)">{{ t('common.edit') }}</el-button>
          <el-popconfirm :title="t('customers.deleteConfirm')" @confirm="handleDelete(row.id)">
            <template #reference>
              <el-button link type="danger">{{ t('common.delete') }}</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="customerStore.total" layout="total, prev, pager, next" style="margin-top: 16px; justify-content: flex-end" @current-change="loadCustomers" />

    <el-dialog v-model="showCreateDialog" :title="editingCustomer ? t('customers.editCustomer') : t('customers.addCustomer')" width="500">
      <el-form :model="form" label-width="80px">
        <el-form-item :label="t('common.name')" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item :label="t('common.email')" required><el-input v-model="form.email" /></el-form-item>
        <el-form-item :label="t('common.company')"><el-input v-model="form.company" /></el-form-item>
        <el-form-item :label="t('common.phone')"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="form.status">
            <el-option :label="t('common.active')" value="active" />
            <el-option :label="t('common.inactive')" value="inactive" />
            <el-option :label="t('common.churned')" value="churned" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('customers.acquiredChannel')"><el-input v-model="form.acquired_channel" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { useCustomerStore } from '../stores/customer'
import type { Customer } from '../api/customer'

const router = useRouter()
const { t } = useI18n()
const customerStore = useCustomerStore()

const page = ref(1)
const pageSize = ref(20)
const search = ref('')
const statusFilter = ref('')
const showCreateDialog = ref(false)
const editingCustomer = ref<Customer | null>(null)
const saving = ref(false)

const form = ref({ name: '', email: '', company: '', phone: '', status: 'active', acquired_channel: '' })

const statusType = (s: string) => s === 'active' ? 'success' : s === 'inactive' ? 'warning' : 'danger'
const statusLabel = (s: string) => s === 'active' ? t('common.active') : s === 'inactive' ? t('common.inactive') : t('common.churned')

const loadCustomers = () => {
  customerStore.fetchCustomers({
    page: page.value,
    page_size: pageSize.value,
    search: search.value || undefined,
    status: statusFilter.value || undefined
  })
}

const goDetail = (id: string) => router.push(`/customers/${id}`)

const editCustomer = (c: Customer) => {
  editingCustomer.value = c
  form.value = { name: c.name, email: c.email, company: c.company || '', phone: c.phone || '', status: c.status, acquired_channel: c.acquired_channel || '' }
  showCreateDialog.value = true
}

const handleSave = async () => {
  saving.value = true
  try {
    if (editingCustomer.value) {
      await customerStore.updateCustomer(editingCustomer.value.id, form.value)
      ElMessage.success(t('customers.customerUpdated'))
    } else {
      await customerStore.createCustomer(form.value)
      ElMessage.success(t('customers.customerCreated'))
    }
    showCreateDialog.value = false
    editingCustomer.value = null
    form.value = { name: '', email: '', company: '', phone: '', status: 'active', acquired_channel: '' }
    loadCustomers()
  } catch {
    ElMessage.error(t('common.saveFailed'))
  } finally {
    saving.value = false
  }
}

const handleDelete = async (id: string) => {
  await customerStore.deleteCustomer(id)
  ElMessage.success(t('customers.customerDeleted'))
  loadCustomers()
}

onMounted(loadCustomers)
</script>

<style scoped>
.customers-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
</style>
