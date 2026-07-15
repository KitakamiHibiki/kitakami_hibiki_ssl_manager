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
  (res) => {
    const body = res.data
    if (body && body.code !== undefined && body.code !== 200) {
      if (body.code === 401 && !window.location.pathname.includes("/login")) {
        const { logout } = useAuth()
        logout()
        window.location.href = '/login'
      }
      const err: any = new Error(body.msg || "request failed")
      err.response = res
      err.code = body.code
      return Promise.reject(err)
    }
    res.data = body?.data !== undefined ? body.data : body
    return res
  },
  (err) => {
    return Promise.reject(err)
  }
)

export interface PlatformInfo {
  os: string
  arch: string
  proxy_url: string
}

export interface User {
  id: number
  username: string
  role: string
  created_at: string
}

// Auth
export function authLogin(email: string, password: string) {
  return api.post('/auth/login', { email, password })
}

export function authRegister(email: string, password: string) {
  return api.post('/auth/register', { email, password })
}

// Certificates
export function applyCertificate(domain: string, email: string, extraDomains?: string[]) {
  return api.post('/certs/apply', { domain, email, extra_domains: extraDomains || [] })
}

export function getCertificate(id: number) {
  return api.get('/certs/detail', { params: { id } })
}

export function getCertStatus(certId: number) {
  return api.get('/certs/status', { params: { cert_id: certId } })
}

export function verifyDNS(domain: string) {
  return api.post('/certs/verify-dns', { domain })
}

export function verifyHTTPProxy(domain: string, domainHash?: string) {
  return api.post('/certs/verify-http-proxy', { domain, domain_hash: domainHash || '' })
}

export function getChallengeValue(domain: string) {
  return api.get('/certs/challenge-value', { params: { domain } })
}

// Nodes
export function getNodes() {
  return api.get<any[]>('/nodes')
}

export function createNode(data: any) {
  return api.post('/nodes', data)
}

export function deleteNode(id: number) {
  return api.delete('/nodes', { params: { id } })
}

export function updateNode(id: number, data: any) {
  return api.put('/nodes', { id, ...data })
}

export function testNode(id: number) {
  return api.get('/nodes/test', { params: { id } })
}

// Deploy
export function deployCert(certId: number) {
  return api.post('/certs/deploy', { cert_id: certId })
}

export function getDeployLogs(params: { cert_id?: number; page?: number; page_size?: number }) {
  return api.get('/certs/deploy-logs', { params })
}

// Platform
export function getPlatform() {
  return api.get<PlatformInfo>('/platform')
}

// Users (admin)
export function getUsers() {
  return api.get<User[]>('/users')
}

export function updateUser(id: number, role: string) {
  return api.put('/users', { id, role })
}

export function deleteUser(id: number) {
  return api.delete('/users', { params: { id } })
}

// Certificate Management
export function getCertificates(params?: { page?: number; page_size?: number }) {
  return api.get('/certificates', { params })
}

export function getCertificateDetail(id: number) {
  return api.get('/certificates/detail', { params: { id } })
}

export function updateCertificate(id: number, data: any) {
  return api.put('/certificates', { id, ...data })
}

export function deleteCertificate(id: number) {
  return api.delete('/certificates', { params: { id } })
}

export default api
