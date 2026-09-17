# Signal — a live polling tool

Create a poll, share the link, watch votes land in real time. Built for the
GUVI × HCL developer internship task.

**Stack:** React (Vite) · Go (Gin) · MongoDB · Redis (live counters + pub/sub)

## How the stack is actually used

- **MongoDB** is the source of truth: users, poll definitions, and a
  permanent per-vote audit log.
- **Redis** drives everything live: vote counts are stored and incremented
  in a Redis hash (`HINCRBY`) so the hot path never hits Mongo, duplicate
  votes are blocked with a Redis set (`SADD`), and every vote is published
  to a Redis pub/sub channel that the backend's WebSocket hub fans out to
  every connected browser — that's what makes results update with no
  refresh.
- **Go/Gin** exposes a small REST API for auth/poll management plus a
  WebSocket endpoint for live results, with all input validation done
  server-side.
- **React** is the voting/creation UI, split cleanly from the backend in
  its own folder.

## Project structure

```
backend/    Go (Gin) API + WebSocket server
frontend/   React (Vite) client
```

## Running it locally

You'll need Go 1.22+, Node 20+, a MongoDB instance, and a Redis instance
(local installs or free-tier Atlas/Upstash both work).

### Backend

```bash
cd backend
cp .env.example .env      # fill in Mongo/Redis URIs and a real JWT secret
go mod tidy                # resolves and locks dependency versions
go run ./cmd/server
```

The API listens on `http://localhost:8080` (health check at `/healthz`).

### Frontend

```bash
cd frontend
cp .env.example .env      # defaults already point at localhost:8080
npm install
npm run dev
```

Open `http://localhost:5173`. Sign up, create a poll, then open the
poll's `/poll/:id` link in a second tab (or another browser) and vote —
the first tab updates instantly.

## Key decisions

- **Voting stays open, account required only to create/manage polls** —
  matches the brief ("basic auth… before someone can create or manage a
  poll"), and keeps the voter experience frictionless.
- **Anonymous voters are deduped with a client-generated ID + IP**, stored
  in a Redis set per poll. It's not bulletproof against a determined
  bad actor (nothing anonymous is), but it stops casual double-voting
  without forcing every voter to create an account.
- **Redis is the live source, Mongo is the durable source.** On poll
  creation the option counters are seeded in Redis at zero; every vote
  updates Redis synchronously and Mongo is appended to as an audit log.
  If Redis ever loses data, the vote log in Mongo is enough to rebuild it.
- **One Redis pub/sub channel per poll, one Gin WebSocket hub per server
  instance.** This scales horizontally: any backend instance can serve
  any client because every instance listens to the same Redis channels.
- **JWT auth with bcrypt-hashed passwords**, deliberately simple per the
  brief's "not expecting anything elaborate."

## Deployment

- **Frontend** → Vercel or Netlify (`npm run build`, output in `dist/`).
- **Backend** → Render or Railway (both support long-lived WebSocket
  connections; point them at the `backend/Dockerfile` or run
  `go run ./cmd/server` directly).
- **MongoDB** → MongoDB Atlas free tier.
- **Redis** → Upstash Redis (supports pub/sub on the free tier).

Set `ALLOWED_ORIGIN` on the backend to the deployed frontend URL, and
`VITE_API_URL` / `VITE_WS_URL` on the frontend to the deployed backend URL
(`https://…` and `wss://…` respectively).

## What's next / extra features to consider

- Poll expiry (`closesAt`) with automatic closing
- Live viewer count per poll (the WebSocket hub already tracks connection
  counts per poll — `Hub.ViewerCount`)
- CSV export of the Mongo vote log
- QR code for the share link
