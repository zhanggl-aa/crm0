<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'

const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const registerForm = reactive({
  name: '',
  email: '',
  password: '',
  tenant_name: ''
})

const rules: FormRules = {
  name: [
    { required: true, message: () => t('auth.nameRequired'), trigger: 'blur' }
  ],
  email: [
    { required: true, message: () => t('auth.emailRequired'), trigger: 'blur' },
    { type: 'email', message: () => t('auth.emailInvalid'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: () => t('auth.passwordRequired'), trigger: 'blur' },
    { min: 6, message: () => t('auth.passwordMin'), trigger: 'blur' }
  ],
  tenant_name: [
    { required: true, message: () => t('auth.tenantNameRequired'), trigger: 'blur' }
  ]
}

async function handleRegister() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await authStore.register({
        name: registerForm.name,
        email: registerForm.email,
        password: registerForm.password,
        tenant_name: registerForm.tenant_name
      })
      ElMessage.success(t('auth.registerSuccess'))
    } catch {
      // error handled by interceptor
    } finally {
      loading.value = false
    }
  })
}

function goToLogin() {
  router.push('/login')
}
</script>

<template>
  <div class="register-container">
    <div class="register-card">
      <div class="register-header">
        <h1 class="register-title">CRM0</h1>
        <p class="register-subtitle">{{ t('auth.createAccount') }}</p>
      </div>

      <el-form
        ref="formRef"
        :model="registerForm"
        :rules="rules"
        label-position="top"
        size="large"
        @keyup.enter="handleRegister"
      >
        <el-form-item :label="t('auth.nameLabel')" prop="name">
          <el-input
            v-model="registerForm.name"
            :placeholder="t('auth.namePlaceholder')"
            prefix-icon="User"
          />
        </el-form-item>

        <el-form-item :label="t('auth.emailLabel')" prop="email">
          <el-input
            v-model="registerForm.email"
            :placeholder="t('auth.emailPlaceholder')"
            prefix-icon="Message"
          />
        </el-form-item>

        <el-form-item :label="t('auth.passwordLabel')" prop="password">
          <el-input
            v-model="registerForm.password"
            type="password"
            :placeholder="t('auth.passwordMinPlaceholder')"
            prefix-icon="Lock"
            show-password
          />
        </el-form-item>

        <el-form-item :label="t('auth.tenantNameLabel')" prop="tenant_name">
          <el-input
            v-model="registerForm.tenant_name"
            :placeholder="t('auth.tenantNamePlaceholder')"
            prefix-icon="OfficeBuilding"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            class="register-btn"
            @click="handleRegister"
          >
            {{ t('auth.registerButton') }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="register-footer">
        <span>{{ t('auth.hasAccount') }}</span>
        <el-link type="primary" @click="goToLogin">{{ t('auth.loginNow') }}</el-link>
      </div>
    </div>
  </div>
</template>

<style scoped>
.register-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.register-card {
  width: 420px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.register-header {
  text-align: center;
  margin-bottom: 32px;
}

.register-title {
  font-size: 36px;
  font-weight: 700;
  color: #409eff;
  margin: 0 0 8px 0;
  letter-spacing: 4px;
}

.register-subtitle {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

.register-btn {
  width: 100%;
}

.register-footer {
  text-align: center;
  margin-top: 16px;
  color: #909399;
  font-size: 14px;
}

.register-footer .el-link {
  font-size: 14px;
}
</style>
