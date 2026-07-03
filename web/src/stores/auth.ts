import { reactive } from 'vue'

interface AuthState {
  token: string
  username: string
  role: string
  email: string
}

const state = reactive<AuthState>({
  token: localStorage.getItem('token') || '',
  username: localStorage.getItem('username') || '',
  role: localStorage.getItem('role') || '',
  email: localStorage.getItem('email') || '',
})

export function useAuth() {
  const isLoggedIn = () => !!state.token

  const login = (token: string, username: string, role: string, email: string) => {
    state.token = token
    state.username = username
    state.role = role
    state.email = email
    localStorage.setItem('token', token)
    localStorage.setItem('username', username)
    localStorage.setItem('role', role)
    localStorage.setItem('email', email)
  }

  const logout = () => {
    state.token = ''
    state.username = ''
    state.role = ''
    state.email = ''
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
    localStorage.removeItem('email')
  }

  return { state, isLoggedIn, login, logout }
}
