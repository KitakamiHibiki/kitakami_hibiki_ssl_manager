import axios from 'axios'
import { useAuth } from '../stores/auth'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

api.interceptors.request.use((config) => {
  const { state } = useAuth()
  if (state.token) {
    config.headers.Authorization = `Bearer ${state.token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      const { logout } = useAuth()
      logout()
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export interface Domain {
  id: number
  domain: string
  email: string
  challenge: string
  created_at: string
}

export interface Certificate {
  id: number
  domain_id: number
  domain: string
  status: string
  issued_at: string
  expires_at: string
  created_at: string
}

export interface PlatformInfo {
  os: string
  arch: string
}

// Auth
export function authLogin(username: string, password: string) {
  return api.post('/auth/login', { username, password })
}

export function authRegister(username: string, password: string) {
  return api.post('/auth/register', { username, password })
}

// Domains
export function getDomains() {
  return api.get<Domain[]>('/domains')
}

export function createDomain(data: { domain: string; email: string; challenge?: string }) {
  return api.post<Domain>('/domains', data)
}

export function deleteDomain(id: number) {
  return api.delete(`/domains/${id}`)
}

// Certificates
export function getCertificates() {
  return api.get<Certificate[]>('/certs')
}

export function getCertificate(id: number) {
  return api.get<Certificate>(`/certs/${id}`)
}

export function applyCertificate(data: { domain_id?: number; domain?: string; email?: string }) {
  return api.post('/certs/apply', data)
}

export function renewCertificate(certificate_id: number) {
  return api.post('/certs/renew', { certificate_id })
}

export function deployCertificate(certificate_id: number, target: string) {
  return api.post('/deploy', { certificate_id, target })
}

// Platform
export function getPlatform() {
  return api.get<PlatformInfo>('/platform')
}

export default api
