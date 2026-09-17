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

// A stable anonymous voter id, separate from account auth, so votes
// can be deduped per-browser even for people who never sign up (only
// poll creators need an account — voting stays open).
export function getVoterId() {
  let id = localStorage.getItem('signal_voter_id')
  if (!id) {
    id = crypto.randomUUID()
    localStorage.setItem('signal_voter_id', id)
  }
  return id
}
