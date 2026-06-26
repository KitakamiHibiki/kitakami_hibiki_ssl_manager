import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

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

export function getDomains() {
  return api.get<Domain[]>('/domains')
}

export function createDomain(data: { domain: string; email: string; challenge?: string }) {
  return api.post<Domain>('/domains', data)
}

export function deleteDomain(id: number) {
  return api.delete(`/domains/${id}`)
}

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

export function getPlatform() {
  return api.get<PlatformInfo>('/platform')
}

export default api
