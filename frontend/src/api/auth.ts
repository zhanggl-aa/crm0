import api from './index'

export interface LoginData {
  email: string
  password: string
}

export interface RegisterData {
  name: string
  email: string
  password: string
  tenant_name: string
}

export interface AuthResponse {
  token: string
  expires_at: string
  user: {
    id: string
    tenant_id: string
    email: string
    name: string
    role: string
  }
}

export function login(data: LoginData): Promise<AuthResponse> {
  return api.post('/auth/login', data)
}

export function register(data: RegisterData): Promise<AuthResponse> {
  return api.post('/auth/register', data)
}

export function refreshToken(): Promise<AuthResponse> {
  return api.post('/auth/refresh')
}
