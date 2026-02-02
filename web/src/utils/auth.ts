export interface User {
  id: number
  username: string
  is_admin: boolean
  created_at: string
  updated_at: string
}

export const getToken = (): string | null => {
  return localStorage.getItem('token')
}

export const setToken = (token: string): void => {
  localStorage.setItem('token', token)
}

export const removeToken = (): void => {
  localStorage.removeItem('token')
}

export const getUser = (): User | null => {
  const userStr = localStorage.getItem('user')
  if (!userStr) return null
  try {
    return JSON.parse(userStr)
  } catch {
    return null
  }
}

export const setUser = (user: User): void => {
  localStorage.setItem('user', JSON.stringify(user))
}

export const removeUser = (): void => {
  localStorage.removeItem('user')
}

export const isAuthenticated = (): boolean => {
  return !!getToken()
}

export const isAdmin = (): boolean => {
  const user = getUser()
  return user?.is_admin || false
}

export const logout = (): void => {
  removeToken()
  removeUser()
  window.location.href = '/login'
}
