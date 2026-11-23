# ⚛ AI2Sim

> **AI-Powered Scientific Simulation Generator**
> Describe any physical, mathematical, biological, or chemical system in plain English — AI2Sim generates an interactive, animated simulation in seconds.

---

## 🚀 Features

- **Natural Language Prompts** — Describe any simulation (e.g. "double pendulum with adjustable gravity")
- **AI-Generated Canvas Code** — `qwen2.5-coder:7b` generates self-contained JavaScript that runs on HTML5 Canvas
- **Live Parameter Sliders** — Auto-inferred sliders for every variable parameter, update the simulation in real-time
- **SSE Streaming** — See the simulation start rendering as the code generates token-by-token
- **Sandboxed Execution** — LLM code runs in an isolated `<iframe>` for security
- **LRU Cache** — Identical prompts are served from cache instantly (no re-inference)
- **Dark Theme** — Premium glassmorphism UI with brand gradients

---

## 🏗 Architecture

```
┌─────────────────────────┐        ┌─────────────────────────────────┐
│   Vercel (Frontend)      │        │   Fly.io (Backend + Ollama)      │
│                         │  HTTPS │                                  │
│   SvelteKit + Svelte 5  │◄──────►│   Go + Fiber v2                  │
│   ├── Prompt Panel      │  SSE   │   ├── /api/v1/parameters         │
│   ├── Parameter Sliders │        │   ├── /api/v1/simulate (SSE)     │
│   └── Canvas Renderer   │        │   └── Concurrency limiter        │
└─────────────────────────┘        │                                  │
                                   │   Ollama (localhost:11434)       │
                                   │   └── qwen2.5-coder:7b           │
                                   └─────────────────────────────────┘
```

---

## 📁 Project Structure

```
AI2Sim/
├── frontend/                   # SvelteKit app (→ Vercel)
│   ├── src/
│   │   ├── routes/+page.svelte     # Main 3-panel layout
│   │   ├── lib/
│   │   │   ├── components/
│   │   │   │   ├── PromptPanel.svelte       # Prompt input + SSE handler
│   │   │   │   ├── SimulationCanvas.svelte  # Sandboxed iframe renderer
│   │   │   │   ├── ControlSidebar.svelte    # Sliders + playback controls
│   │   │   │   └── ParameterSlider.svelte   # Single slider component
│   │   │   ├── stores/simulation.ts         # Svelte state management
│   │   │   └── api.ts                       # Backend API client
│   │   └── app.css              # Dark theme design system
│   └── .env                    # VITE_API_URL=http://localhost:8080
│
├── backend/                    # Go server (→ Fly.io)
│   ├── cmd/server/main.go          # Entry point
│   ├── internal/
│   │   ├── api/handlers.go         # REST + SSE handlers
│   │   ├── llm/
│   │   │   ├── ollama.go           # Ollama HTTP API client
│   │   │   └── prompt_builder.go   # Prompt engineering + param schemas
│   │   └── middleware/limiter.go   # Concurrency gate middleware
│   ├── Dockerfile              # Multi-stage: Go + Ollama
│   ├── docker-entrypoint.sh    # Start Ollama, pull model, start server
│   └── fly.toml                # Fly.io deployment config
│
└── docker-compose.yml          # Local dev: Ollama + backend together
```

---

## 🛠 Local Development

### Prerequisites
- [Go 1.24+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Ollama](https://ollama.com/) installed and running
- [Docker](https://docker.com/) (optional, for docker-compose approach)

### Option A: Direct (no Docker)

**1. Pull the LLM model:**
```bash
ollama pull qwen2.5-coder:7b
```

**2. Start the Go backend:**
```bash
cd backend
go run ./cmd/server
# Server starts on http://localhost:8080
```

**3. Start the SvelteKit frontend:**
```bash
cd frontend
npm install
npm run dev
# Opens http://localhost:5173
```

### Option B: Docker Compose

```bash
# From project root — pulls the model and starts everything
docker compose up
```
Frontend dev server still runs locally:
```bash
cd frontend && npm run dev
```

---

## 🚀 Deployment

### Frontend → Vercel

```bash
cd frontend
npm install -g vercel
vercel login
vercel deploy --prod

# Set your backend URL as an env variable in Vercel dashboard:
# VITE_API_URL = https://ai2sim-backend.fly.dev
```

### Backend → Fly.io

```bash
cd backend

# Install Fly CLI: https://fly.io/docs/hands-on/install-flyctl/
fly auth login

# Create the app (first time only)
fly launch --no-deploy

# Create persistent volume for Ollama models
fly volumes create ollama_models --size 30 --region sin

# Set CORS to allow your Vercel domain
fly secrets set ALLOWED_ORIGINS="https://your-app.vercel.app"

# Deploy
fly deploy
```

> **Note:** First deployment will pull `qwen2.5-coder:7b` (~4.7 GB) — this takes ~5-10 minutes. Subsequent deploys are fast.

---

## 🔧 Configuration

### Backend Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama API endpoint |
| `OLLAMA_MODEL` | `qwen2.5-coder:7b` | Model to use for code generation |
| `ALLOWED_ORIGINS` | `http://localhost:5173` | Comma-separated CORS origins |

### Frontend Environment Variables (`.env`)

| Variable | Default | Description |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080` | Go backend base URL |

---

## 🧠 Supported Simulation Types

| Type | Examples |
|---|---|
| **Physics** | Double pendulum, N-body gravity, projectile motion, wave interference |
| **Math & Chaos** | Mandelbrot/Julia sets, Lorenz attractor, Fourier series |
| **Biology** | Boid flocking, predator-prey, cellular automata, Conway's Life |
| **Chemistry** | Gray-Scott reaction diffusion, Brownian motion, ideal gas molecules |

---

## 🔑 Key Technical Decisions

| Decision | Rationale |
|---|---|
| Sandboxed `<iframe>` for JS | LLM-generated code could be malicious; iframe `sandbox="allow-scripts"` isolates it completely |
| SSE over WebSocket | Unidirectional streaming is simpler, works behind all proxies/CDNs |
| Ollama HTTP API (not Go SDK) | SDK requires Go 1.26; direct HTTP calls work with Go 1.24 and have zero extra deps |
| LRU cache in Go | Identical prompt+params = instant replay, no wasted inference |
| Semaphore concurrency limiter | Prevents server OOM from too many parallel 7B model inferences |
| Model warm-up on startup | Eliminates 10-15 second first-request cold start |

---

## 📄 License

MIT
