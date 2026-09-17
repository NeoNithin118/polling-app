import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { api } from '../api/client.js'
import { useAuth } from '../context/AuthContext.jsx'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const returnTo = searchParams.get('returnTo')

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await api.login(email, password)
      login(data.token, data.user)
      navigate(returnTo || '/dashboard')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <form className="auth-card" onSubmit={handleSubmit}>
        <h1>Sign in</h1>
        <p className="auth-subtitle">Manage the polls you've created.</p>

        {error && <div className="form-error">{error}</div>}

        <label>
          Email
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </label>

        <label>
          Password
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>

        <button className="btn btn-accent btn-block" type="submit" disabled={loading}>
          {loading ? 'Signing in…' : 'Sign in'}
        </button>

        <p className="auth-switch">
          New here? <Link to="/signup">Create an account</Link>
        </p>
      </form>
    </div>
  )
}
