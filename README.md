<div align="center">

# ✨ AI2Sim

**Turn natural language into interactive simulations — instantly.**

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?style=for-the-badge&logo=svelte&logoColor=white)](https://svelte.dev/)
[![Vite](https://img.shields.io/badge/Vite-8-646CFF?style=for-the-badge&logo=vite&logoColor=white)](https://vite.dev/)
[![Gemini](https://img.shields.io/badge/Gemini-2.0_Flash-4285F4?style=for-the-badge&logo=google&logoColor=white)](https://ai.google.dev/)

*Describe a simulation in plain English. Watch it come to life on a canvas. Tweak it in real-time with auto-generated controls.*

---

</div>

## 🎯 What is AI2Sim?

AI2Sim is a full-stack web application that bridges the gap between **imagination and simulation**. Type a prompt like *"A solar system with orbiting planets"* or *"A flock of birds with swarming behavior"*, and AI2Sim will:

1. 🧠 Send your prompt to **Google Gemini** (LLM)
2. ⚡ Generate real, executable **JavaScript Canvas 2D** simulation code
3. 🎨 Render it live on an **HTML5 Canvas** with smooth 60fps animation
4. 🎛️ Auto-generate **parameter sliders** so you can tweak physics, visuals, and behavior in real-time

No coding required. Just describe, generate, and explore.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Browser (Svelte 5)                │
│  ┌──────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │ PromptBar│  │  Simulation  │  │  Parameter    │  │
│  │          │──│  Viewport    │──│  Panel        │  │
│  │          │  │  (Canvas 2D) │  │  (Auto-gen)   │  │
│  └────┬─────┘  └──────────────┘  └───────────────┘  │
│       │            ▲                                │
│       │ POST /api/generate                          │
│       ▼            │                                │
│  ┌─────────────────┴───────────────────────────┐    │
│  │         Vite Dev Proxy (/api → :8080)       │    │
│  └─────────────────┬───────────────────────────┘    │
└────────────────────┼────────────────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────────────┐
│              Go Backend (:8080)                    │
│  ┌────────────┐  ┌───────────┐  ┌──────────────┐  │
│  │ Middleware  │  │  Handler  │  │  LLM Service │  │
│  │ CORS+Logs  │──│ /api/*    │──│  (Gemini SDK)│  │
│  └────────────┘  └───────────┘  └──────┬───────┘  │
└─────────────────────────────────────────┼──────────┘
                                          │
                                          ▼
                                ┌──────────────────┐
                                │  Google Gemini   │
                                │  2.0 Flash API   │
                                └──────────────────┘
```

---

## 🚀 Quick Start

### Prerequisites

- [Go 1.24+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Gemini API Key](https://aistudio.google.com/apikey) (free)

### 1. Clone the repo

```bash
git clone https://github.com/konark-17/AI2Sim.git
cd AI2Sim
```

### 2. Start the Go backend

```bash
# Set your Gemini API key
# Linux/macOS:
export GEMINI_API_KEY="your-api-key-here"
# Windows (PowerShell):
$env:GEMINI_API_KEY = "your-api-key-here"

cd backend
go run .
```

> The backend starts on `http://localhost:8080`

### 3. Start the frontend (new terminal)

```bash
cd frontend
npm install
npm run dev
```

> The frontend starts on `http://localhost:5173`

### 4. Open and explore!

Navigate to **http://localhost:5173** and try prompts like:

| Prompt | What you get |
|--------|-------------|
| *"A particle system with gravity and wind"* | Particles falling with adjustable gravity, wind, and count |
| *"A bouncing ball simulation"* | Ball physics with elasticity, speed, and size controls |
| *"Conway's Game of Life"* | Cellular automaton with grid size and speed sliders |
| *"A starfield flying through space"* | 3D star effect with speed, density, and color controls |
| *"A wave interference pattern"* | Overlapping sine waves with frequency and amplitude |

---

## 📂 Project Structure

```
AI2Sim/
├── backend/                  # Go HTTP server
│   ├── main.go               # Entry point — server setup & routing
│   ├── handler.go            # HTTP handlers (POST /api/generate, GET /api/health)
│   ├── llm.go                # Gemini SDK integration & system prompt
│   ├── middleware.go          # CORS & request logging
│   ├── types.go              # Shared request/response types
│   ├── go.mod
│   └── go.sum
│
├── frontend/                 # Svelte 5 + Vite SPA
│   ├── src/
│   │   ├── App.svelte                  # Root layout + API orchestration
│   │   ├── app.css                     # Global design tokens & theme
│   │   ├── main.js                     # Svelte mount
│   │   └── lib/
│   │       ├── SimulationViewport.svelte  # Canvas + dynamic code executor
│   │       ├── PromptBar.svelte           # Input bar with loading states
│   │       └── ParameterPanel.svelte      # Auto-generated slider panel
│   ├── vite.config.js         # Vite config with API proxy
│   ├── index.html
│   └── package.json
└── README.md
```

---

## 🔌 API Reference

### `POST /api/generate`

Generate a simulation from a natural language prompt.

**Request:**
```json
{
  "prompt": "A particle system with gravity"
}
```

**Response:**
```json
{
  "code": "ctx.clearRect(0, 0, canvas.width, canvas.height); ...",
  "parameterDefs": [
    { "name": "gravity", "label": "Gravity", "min": 0.1, "max": 10, "step": 0.1 },
    { "name": "particleCount", "label": "Particle Count", "min": 10, "max": 500, "step": 1 }
  ],
  "defaults": {
    "gravity": 1.0,
    "particleCount": 100
  }
}
```

### `GET /api/health`

Health check endpoint.

**Response:**
```json
{ "status": "ok", "service": "ai2sim-backend" }
```

---

## 🛠️ Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Frontend** | Svelte 5 | Reactive UI components |
| **Bundler** | Vite 8 | Dev server + HMR + API proxy |
| **Backend** | Go 1.24 | HTTP server (stdlib `net/http`) |
| **AI/LLM** | Google Gemini 2.0 Flash | Code generation from prompts |
| **Rendering** | HTML5 Canvas 2D | Real-time simulation display |

---

## 🔮 Future Ideas

- 🗄️ **Simulation history** — Save and revisit past generations
- 🔗 **Shareable links** — Share simulations via URL
- 🧪 **WebGL support** — 3D simulations with Three.js
- 📱 **Mobile responsive** — Touch controls for parameters
- 🎨 **Prompt gallery** — Pre-built community prompts
