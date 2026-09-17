import { useEffect, useRef, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api, WS_URL } from '../api/client.js'
import { useAuth } from '../context/AuthContext.jsx'
import OptionBar from '../components/OptionBar.jsx'
import PulseDot from '../components/PulseDot.jsx'

function votedKey(pollId) {
  return `signal_voted_${pollId}`
}

export default function PollView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { user, token } = useAuth()
  const [poll, setPoll] = useState(null)
  const [counts, setCounts] = useState({})
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [voting, setVoting] = useState(false)
  const [connected, setConnected] = useState(false)
  const [votedOption, setVotedOption] = useState(() => localStorage.getItem(votedKey(id)))
  const [copied, setCopied] = useState(false)
  const socketRef = useRef(null)

  const loadPoll = useCallback(() => {
    api
      .getPoll(id)
      .then((data) => {
        setPoll(data.poll)
        setCounts(data.counts || {})
        setTotal(data.total || 0)
      })
      .catch((err) => setError(err.message))
  }, [id])

  useEffect(() => {
    loadPoll()
  }, [loadPoll])

  // Open one WebSocket connection per poll view. Every vote broadcast
  // for this poll (from any voter, anywhere) arrives here and updates
  // state directly — no polling, no refresh.
  useEffect(() => {
    const wsScheme = WS_URL.startsWith('https') ? 'wss' : WS_URL.startsWith('http') ? 'ws' : ''
    const base = wsScheme ? WS_URL.replace(/^https?/, wsScheme) : WS_URL
    const socket = new WebSocket(`${base}/ws/polls/${id}`)
    socketRef.current = socket

    socket.onopen = () => setConnected(true)
    socket.onclose = () => setConnected(false)
    socket.onerror = () => setConnected(false)
    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data)
        if (payload.pollId === id) {
          setCounts(payload.counts || {})
          setTotal(payload.total || 0)
        }
      } catch {
        // ignore malformed frames
      }
    }

    return () => socket.close()
  }, [id])

  async function handleVote(optionId) {
    if (votedOption || voting) return
  
    // User must be logged in to vote.
    // The poll itself remains publicly viewable.
    if (!token) {
      navigate(`/login?returnTo=/poll/${id}`)
      return
    }
  
    setVoting(true)
    setError('')
  
    try {
      const data = await api.vote(id, optionId, token)
  
      setCounts(data.counts)
      setTotal(data.total)
  
      localStorage.setItem(votedKey(id), optionId)
      setVotedOption(optionId)
    } catch (err) {
      setError(err.message)
    } finally {
      setVoting(false)
    }
  }

  async function handleClose() {
    try {
      await api.closePoll(id, token)
      loadPoll()
    } catch (err) {
      setError(err.message)
    }
  }

  function copyLink() {
    navigator.clipboard.writeText(window.location.href)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  if (error && !poll) {
    return (
      <div className="page">
        <div className="form-error">{error}</div>
      </div>
    )
  }

  if (!poll) {
    return (
      <div className="page">
        <p className="muted">Loading…</p>
      </div>
    )
  }

  const isOwner = user && user.id === poll.ownerId
  const maxCount = Math.max(0, ...Object.values(counts))

  return (
    <div className="page page-narrow">
      <div className="poll-header">
        <PulseDot label={poll.isActive ? (connected ? 'Live' : 'Reconnecting…') : 'Closed'} />
        {isOwner && poll.isActive && (
          <button className="btn btn-ghost btn-small" onClick={handleClose}>
            Close poll
          </button>
        )}
      </div>

      <h1 className="poll-question">{poll.question}</h1>

      {error && <div className="form-error">{error}</div>}

      <div className="option-list">
        {poll.options.map((opt) => (
          <OptionBar
            key={opt.id}
            option={opt}
            count={counts[opt.id] || 0}
            total={total}
            onVote={!votedOption && poll.isActive ? handleVote : undefined}
            disabled={voting}
            isWinner={total > 0 && (counts[opt.id] || 0) === maxCount && maxCount > 0}
          />
        ))}
      </div>

      <p className="poll-footnote">
        {votedOption
          ? "You've voted. Results update live as others vote."
          : poll.isActive
          ? 'Tap an option to vote.'
          : 'This poll is closed.'}
        {' · '}
        {total} {total === 1 ? 'vote' : 'votes'} so far
      </p>

      <button className="btn btn-ghost" onClick={copyLink}>
        {copied ? 'Link copied' : 'Copy share link'}
      </button>
    </div>
  )
}
