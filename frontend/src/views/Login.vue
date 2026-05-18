<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const loginForm = reactive({
  email: '',
  password: ''
})

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ]
}

async function handleLogin() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await authStore.login({
        email: loginForm.email,
        password: loginForm.password
      })
      ElMessage.success('登录成功')
    } catch {
      // error handled by interceptor
    } finally {
      loading.value = false
    }
  })
}

function goToRegister() {
  router.push('/register')
}
</script>

<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1 class="login-title">CRM0</h1>
        <p class="login-subtitle">SaaS 智能CRM系统</p>
      </div>

      <el-form
        ref="formRef"
        :model="loginForm"
        :rules="rules"
        label-position="top"
        size="large"
        @keyup.enter="handleLogin"
      >
        <el-form-item label="邮箱地址" prop="email">
          <el-input
            v-model="loginForm.email"
            placeholder="请输入邮箱地址"
            prefix-icon="Message"
          />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="Lock"
            show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            class="login-btn"
            @click="handleLogin"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <span>还没有账号？</span>
        <el-link type="primary" @click="goToRegister">立即注册</el-link>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 420px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-title {
  font-size: 36px;
  font-weight: 700;
  color: #409eff;
  margin: 0 0 8px 0;
  letter-spacing: 4px;
}

.login-subtitle {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

.login-btn {
  width: 100%;
}

.login-footer {
  text-align: center;
  margin-top: 16px;
  color: #909399;
  font-size: 14px;
}

.login-footer .el-link {
  font-size: 14px;
}
</style>
