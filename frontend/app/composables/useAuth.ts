export interface User {
  id: string
  email: string
  full_name: string
  phone: string
  role: 'renter' | 'owner' | 'admin'
}

export function useAuth() {
  const user = useState<User | null>('auth-user', () => null)
  const api = useApi()

  async function fetchMe() {
    try {
      user.value = await api<User>('/api/auth/me')
    } catch {
      user.value = null
    }
  }

  async function login(email: string, password: string) {
    user.value = await api<User>('/api/auth/login', {
      method: 'POST',
      body: { email, password }
    })
  }

  async function register(body: {
    email: string
    password: string
    full_name: string
    phone: string
    role: 'renter' | 'owner'
  }) {
    user.value = await api<User>('/api/auth/register', {
      method: 'POST',
      body
    })
  }

  async function logout() {
    await api('/api/auth/logout', { method: 'POST' })
    user.value = null
    navigateTo('/')
  }

  return { user, fetchMe, login, register, logout }
}
