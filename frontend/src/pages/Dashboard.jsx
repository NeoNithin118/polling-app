import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client.js'
import { useAuth } from '../context/AuthContext.jsx'

export default function Dashboard() {
  const { token } = useAuth()
  const [polls, setPolls] = useState(null)
  const [error, setError] = useState('')
  const [copiedId, setCopiedId] = useState(null)

  useEffect(() => {
    api
      .myPolls(token)
      .then(setPolls)
      .catch((err) => setError(err.message))
  }, [token])

  function copyLink(id) {
    const url = `${window.location.origin}/poll/${id}`
    navigator.clipboard.writeText(url)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 1500)
  }

  return (
    <div className="page">
      <div className="page-header">
        <h1>Your polls</h1>
        <Link to="/create" className="btn btn-accent">
          New poll
        </Link>
      </div>

      {error && <div className="form-error">{error}</div>}

      {polls === null && !error && <p className="muted">Loading…</p>}

      {polls?.length === 0 && (
        <div className="empty-state">
          <p>You haven't created a poll yet.</p>
          <Link to="/create" className="btn btn-accent">
            Create your first poll
          </Link>
        </div>
      )}

      <ul className="poll-list">
        {polls?.map((poll) => (
          <li key={poll.id} className="poll-list-item">
            <div>
              <Link to={`/poll/${poll.id}`} className="poll-list-question">
                {poll.question}
              </Link>
              <div className="poll-list-meta">
                <span className={poll.isActive ? 'tag tag-live' : 'tag tag-closed'}>
                  {poll.isActive ? 'Live' : 'Closed'}
                </span>
                <span>{poll.options.length} options</span>
              </div>
            </div>
            <button className="btn btn-ghost" onClick={() => copyLink(poll.id)}>
              {copiedId === poll.id ? 'Copied' : 'Copy link'}
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
