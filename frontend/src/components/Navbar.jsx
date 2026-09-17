import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext.jsx'
import PulseDot from './PulseDot.jsx'

export default function Navbar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <header className="navbar">
      <Link to="/dashboard" className="navbar-brand">
        <PulseDot />
        <span>Signal</span>
      </Link>

      <nav className="navbar-links">
        {user ? (
          <>
            <Link to="/create" className="btn btn-accent">
              New poll
            </Link>
            <span className="navbar-user">{user.name}</span>
            <button
              className="btn btn-ghost"
              onClick={() => {
                logout()
                navigate('/login')
              }}
            >
              Sign out
            </button>
          </>
        ) : (
          <>
            <Link to="/login" className="btn btn-ghost">
              Sign in
            </Link>
            <Link to="/signup" className="btn btn-accent">
              Create account
            </Link>
          </>
        )}
      </nav>
    </header>
  )
}
