#!/bin/bash
set -e

echo "🤖 Starting Ollama service..."
ollama serve &
OLLAMA_PID=$!

# Wait for Ollama to be ready
echo "⏳ Waiting for Ollama to start..."
until curl -s http://localhost:11434/api/tags > /dev/null 2>&1; do
  sleep 1
done
echo "✅ Ollama is ready"

# Pull model if not already present
MODEL="${OLLAMA_MODEL:-qwen2.5-coder:7b}"
echo "📦 Ensuring model $MODEL is available..."
ollama pull "$MODEL" || echo "⚠️  Model pull failed — may already be cached"

echo "🚀 Starting AI2Sim server..."
exec /usr/local/bin/ai2sim-server
