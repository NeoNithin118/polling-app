import { createContext, useContext, useEffect, useState } from 'react'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() => localStorage.getItem('signal_token'))
  const [user, setUser] = useState(() => {
    const raw = localStorage.getItem('signal_user')
    return raw ? JSON.parse(raw) : null
  })

  useEffect(() => {
    if (token) localStorage.setItem('signal_token', token)
    else localStorage.removeItem('signal_token')
  }, [token])

  useEffect(() => {
    if (user) localStorage.setItem('signal_user', JSON.stringify(user))
    else localStorage.removeItem('signal_user')
  }, [user])

  function login(nextToken, nextUser) {
    setToken(nextToken)
    setUser(nextUser)
  }

  function logout() {
    setToken(null)
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ token, user, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}

