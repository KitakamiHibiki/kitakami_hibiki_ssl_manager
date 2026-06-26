import { reactive } from 'vue'

interface AuthState {
  token: string
  username: string
  role: string
}

const state = reactive<AuthState>({
  token: localStorage.getItem('token') || '',
  username: localStorage.getItem('username') || '',
  role: localStorage.getItem('role') || '',
})

export function useAuth() {
  const isLoggedIn = () => !!state.token

  const login = (token: string, username: string, role: string) => {
    state.token = token
    state.username = username
    state.role = role
    localStorage.setItem('token', token)
    localStorage.setItem('username', username)
    localStorage.setItem('role', role)
  }

  const logout = () => {
    state.token = ''
    state.username = ''
    state.role = ''
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
  }

  return { state, isLoggedIn, login, logout }
}
