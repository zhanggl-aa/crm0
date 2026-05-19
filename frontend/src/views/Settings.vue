<template>
  <div class="settings-page">
    <h2>{{ t('settings.title') }}</h2>
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card :header="t('settings.orgInfo')">
          <el-form :model="tenantForm" label-width="100px">
            <el-form-item :label="t('settings.orgName')"><el-input v-model="tenantForm.name" disabled /></el-form-item>
            <el-form-item :label="t('settings.currentPlan')"><el-tag>{{ tenantForm.plan }}</el-tag></el-form-item>
          </el-form>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card :header="t('settings.personalInfo')">
          <el-form :model="userForm" label-width="100px">
            <el-form-item :label="t('common.name')"><el-input v-model="userForm.name" /></el-form-item>
            <el-form-item :label="t('common.email')"><el-input v-model="userForm.email" disabled /></el-form-item>
            <el-form-item :label="t('settings.role')"><el-tag>{{ userForm.role }}</el-tag></el-form-item>
            <el-form-item><el-button type="primary" @click="handleSaveProfile" :loading="saving">{{ t('settings.saveChanges') }}</el-button></el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()
const saving = ref(false)
const tenantForm = ref({ name: '', plan: '' })
const userForm = ref({ name: '', email: '', role: '' })

onMounted(() => {
  const user = authStore.user
  if (user) { userForm.value = { name: user.name, email: user.email, role: user.role } }
  tenantForm.value = { name: t('settings.myOrg'), plan: 'free' }
})

const handleSaveProfile = async () => {
  saving.value = true
  try { ElMessage.success(t('settings.profileUpdated')) } finally { saving.value = false }
}
</script>

<style scoped>
.settings-page { padding: 20px; }
</style>
