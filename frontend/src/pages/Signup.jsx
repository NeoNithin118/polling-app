import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { api } from '../api/client.js'
import { useAuth } from '../context/AuthContext.jsx'

export default function Signup() {
  const [name, setName] = useState('')
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

    if (password.length < 8) {
      setError('Password needs to be at least 8 characters.')
      return
    }

    setLoading(true)
    try {
      const data = await api.signup(name, email, password)
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
        <h1>Create an account</h1>
        <p className="auth-subtitle">
          Create an account to manage polls and vote.
        </p>

        {error && <div className="form-error">{error}</div>}

        <label>
          Name
          <input value={name} onChange={(e) => setName(e.target.value)} required minLength={2} />
        </label>

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
            minLength={8}
          />
        </label>

        <button className="btn btn-accent btn-block" type="submit" disabled={loading}>
          {loading ? 'Creating account…' : 'Create account'}
        </button>

        <p className="auth-switch">
          Already have an account? <Link to="/login">Sign in</Link>
        </p>
      </form>
    </div>
  )
}
