import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client.js'
import { useAuth } from '../context/AuthContext.jsx'

export default function CreatePoll() {
  const [question, setQuestion] = useState('')
  const [options, setOptions] = useState(['', ''])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { token } = useAuth()
  const navigate = useNavigate()

  function updateOption(index, value) {
    setOptions((prev) => prev.map((o, i) => (i === index ? value : o)))
  }

  function addOption() {
    if (options.length >= 10) return
    setOptions((prev) => [...prev, ''])
  }

  function removeOption(index) {
    if (options.length <= 2) return
    setOptions((prev) => prev.filter((_, i) => i !== index))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')

    const cleaned = options.map((o) => o.trim()).filter(Boolean)
    if (question.trim().length < 3) {
      setError('Give your poll a question with at least 3 characters.')
      return
    }
    if (cleaned.length < 2) {
      setError('Add at least 2 non-empty options.')
      return
    }

    setLoading(true)
    try {
      const poll = await api.createPoll(question.trim(), cleaned, token)
      navigate(`/poll/${poll.id}`)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page page-narrow">
      <h1>New poll</h1>

      <form className="poll-form" onSubmit={handleSubmit}>
        {error && <div className="form-error">{error}</div>}

        <label>
          Question
          <input
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            placeholder="What should we build next?"
            maxLength={300}
            required
          />
        </label>

        <div className="options-editor">
          <span className="options-editor-label">Options</span>
          {options.map((opt, i) => (
            <div className="option-input-row" key={i}>
              <input
                value={opt}
                onChange={(e) => updateOption(i, e.target.value)}
                placeholder={`Option ${i + 1}`}
                maxLength={150}
              />
              {options.length > 2 && (
                <button
                  type="button"
                  className="btn btn-ghost btn-small"
                  onClick={() => removeOption(i)}
                  aria-label={`Remove option ${i + 1}`}
                >
                  Remove
                </button>
              )}
            </div>
          ))}
          {options.length < 10 && (
            <button type="button" className="btn btn-ghost btn-small" onClick={addOption}>
              + Add option
            </button>
          )}
        </div>

        <button className="btn btn-accent btn-block" type="submit" disabled={loading}>
          {loading ? 'Publishing…' : 'Publish poll'}
        </button>
      </form>
    </div>
  )
}
