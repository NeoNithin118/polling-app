const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
export const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080'

async function request(path, { method = 'GET', body, token } = {}) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`

  const res = await fetch(`${API_URL}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(data.error || `Request failed (${res.status})`)
  }
  return data
}

export const api = {
  signup: (name, email, password) =>
    request('/api/auth/signup', { method: 'POST', body: { name, email, password } }),

  login: (email, password) =>
    request('/api/auth/login', { method: 'POST', body: { email, password } }),

  createPoll: (question, options, token) =>
    request('/api/polls', { method: 'POST', body: { question, options }, token }),

  myPolls: (token) => request('/api/polls', { token }),

  getPoll: (id) => request(`/api/polls/${id}`),

  vote: (id, optionId, token) =>
    request(`/api/polls/${id}/vote`, {
      method: 'POST',
      body: { optionId },
      token
    }),

  closePoll: (id, token) =>
    request(`/api/polls/${id}/close`, { method: 'PATCH', token }),
}
