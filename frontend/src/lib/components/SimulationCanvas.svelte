<script lang="ts">
  import {
    simulationCode,
    simulationStatus,
    paramValues,
    isPlaying,
    currentPrompt,
  } from '$lib/stores/simulation';
  import { onDestroy } from 'svelte';

  let iframeEl: HTMLIFrameElement | undefined = $state();
  let statusText = $state('Waiting for simulation...');

  function extractRunnableCode(raw: string): string {
    let cleaned = raw.trim();
    if (cleaned.includes('```')) {
      const match = cleaned.match(/```(?:javascript|js)?\s*([\s\S]*?)(?:```|$)/i);
      if (match && match[1]) {
        cleaned = match[1].trim();
      }
    }
    // If there's preamble text before "function createSimulation", slice from there
    const fnIndex = cleaned.indexOf('function createSimulation');
    if (fnIndex !== -1) {
      cleaned = cleaned.slice(fnIndex);
      // Balanced brace search to strip any trailing explanations after createSimulation
      let braceDepth = 0;
      let started = false;
      let endIdx = -1;
      for (let i = 0; i < cleaned.length; i++) {
        if (cleaned[i] === '{') {
          braceDepth++;
          started = true;
        } else if (cleaned[i] === '}') {
          braceDepth--;
          if (started && braceDepth === 0) {
            endIdx = i + 1;
            break;
          }
        }
      }
      if (endIdx !== -1) {
        cleaned = cleaned.slice(0, endIdx);
      }
    }
    return cleaned;
  }

  // Build the sandboxed iframe HTML that wraps the LLM code.
  // The code is embedded as JSON inside a <script type="application/json"> tag,
  // which avoids any escaping issues with backticks or special characters.
  function buildIframeHTML(rawCode: string, params: Record<string, number>): string {
    const code = extractRunnableCode(rawCode);
    const payload = JSON.stringify({ code, params });
    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    html, body { width: 100%; height: 100%; overflow: hidden; background: #07070f; }
    canvas { display: block; width: 100%; height: 100%; }
    #error-overlay {
      display: none;
      position: fixed; inset: 0;
      background: rgba(7,7,15,0.92);
      color: #f87171;
      font-family: 'JetBrains Mono', monospace;
      font-size: 13px;
      padding: 24px;
      white-space: pre-wrap;
      word-break: break-all;
      z-index: 100;
    }
  </style>
</head>
<body>
  <canvas id="sim-canvas"></canvas>
  <div id="error-overlay"></div>
  <script id="sim-payload" type="application/json">${payload}<\/script>
  <script>
    (function() {
      var payload = JSON.parse(document.getElementById('sim-payload').textContent);
      var simCode = payload.code;
      var initialParams = payload.params;

      var canvas = document.getElementById('sim-canvas');
      var errorOverlay = document.getElementById('error-overlay');
      var simController = null;
      var createSim = null;

      // Compile the LLM code once
      try {
        var compileFn = new Function(simCode + '; if (typeof createSimulation === "function") return createSimulation; if (typeof window.createSimulation === "function") return window.createSimulation; return null;');
        createSim = compileFn();
        if (typeof createSim !== 'function') {
          throw new Error('createSimulation function was not found in the generated code.');
        }
      } catch (compileErr) {
        errorOverlay.style.display = 'block';
        errorOverlay.textContent = 'Compile Error:\\n\\n' + (compileErr.stack || compileErr.message);
      }

      function resize() {
        canvas.width  = window.innerWidth;
        canvas.height = window.innerHeight;
      }

      function startSim(params) {
        if (typeof createSim !== 'function') return;
        resize();
        try {
          errorOverlay.style.display = 'none';
          simController = createSim(canvas, params);
          window.__simController = simController;
        } catch (err) {
          errorOverlay.style.display = 'block';
          errorOverlay.textContent = 'Runtime Error:\\n\\n' + (err.stack || err.message);
        }
      }

      window.addEventListener('resize', function() {
        if (simController && simController.stop) simController.stop();
        startSim(initialParams);
      });

      // Message listener for play/pause/params updates from parent
      window.addEventListener('message', function(e) {
        if (e.data.type === 'STOP' && window.__simController) {
          window.__simController.stop();
        } else if (e.data.type === 'RESTART') {
          if (window.__simController) window.__simController.stop();
          startSim(initialParams);
        } else if (e.data.type === 'UPDATE_PARAMS') {
          if (window.__simController) window.__simController.stop();
          initialParams = e.data.params;
          startSim(initialParams);
        }
      });

      startSim(initialParams);
    })();
  <\/script>
</body>
</html>`;
  }

  // Reactively update the iframe when code or params change
  let prevCode = '';
  let prevParams = '';

  $effect(() => {
    const code = $simulationCode;
    const params = $paramValues;
    const status = $simulationStatus;

    if (!iframeEl) return;

    const paramsStr = JSON.stringify(params);
    const codeChanged = code !== prevCode;
    const paramsChanged = paramsStr !== prevParams;

    if (status === 'ready' && codeChanged) {
      iframeEl.srcdoc = buildIframeHTML(code, params);
      prevCode = code;
      prevParams = paramsStr;
      statusText = 'Simulation running';
    } else if (status === 'ready' && paramsChanged) {
      iframeEl.contentWindow?.postMessage({ type: 'UPDATE_PARAMS', params }, '*');
      prevParams = paramsStr;
    }
  });

  // Handle play/pause
  $effect(() => {
    const playing = $isPlaying;
    if (!iframeEl || !iframeEl.contentWindow) return;
    iframeEl.contentWindow.postMessage(
      { type: playing ? 'RESTART' : 'STOP' },
      '*'
    );
  });

  // Status text derived from simulation status
  $effect(() => {
    switch ($simulationStatus) {
      case 'idle': statusText = 'Enter a prompt to generate a simulation'; break;
      case 'fetching-params': statusText = 'Analyzing your prompt...'; break;
      case 'streaming': statusText = 'Generating simulation code with AI...'; break;
      case 'error': statusText = 'Simulation failed'; break;
    }
  });

  function setExamplePrompt(prompt: string) {
    currentPrompt.set(prompt);
  }
</script>

<div class="canvas-container">
  {#if $simulationStatus === 'idle'}
    <!-- Empty state -->
    <div class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
          <circle cx="32" cy="32" r="28" stroke="url(#grad1)" stroke-width="2" stroke-dasharray="4 3" />
          <circle cx="32" cy="32" r="14" fill="url(#grad2)" opacity="0.2"/>
          <path d="M22 32 Q32 18 42 32 Q32 46 22 32Z" fill="url(#grad1)" opacity="0.6"/>
          <defs>
            <linearGradient id="grad1" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#8a63ff"/>
              <stop offset="50%" stop-color="#3ecfff"/>
              <stop offset="100%" stop-color="#ff6b9d"/>
            </linearGradient>
            <radialGradient id="grad2" cx="50%" cy="50%" r="50%">
              <stop offset="0%" stop-color="#8a63ff"/>
              <stop offset="100%" stop-color="transparent"/>
            </radialGradient>
          </defs>
        </svg>
      </div>
      <h3>No simulation running</h3>
      <p>Describe a phenomenon to simulate using natural language.<br/>
         The AI will generate an interactive visualization for you.</p>
      <div class="example-prompts">
        <span class="example-label">Try:</span>
        <button class="example-chip" onclick={() => setExamplePrompt('Simulate a double pendulum with chaotic trajectory trails')}>Double pendulum</button>
        <button class="example-chip" onclick={() => setExamplePrompt('Simulate Gray-Scott reaction diffusion with tunable feed and kill rates')}>Reaction diffusion</button>
        <button class="example-chip" onclick={() => setExamplePrompt('Simulate boid flocking behavior with separation, alignment, and cohesion')}>Boid flocking</button>
        <button class="example-chip" onclick={() => setExamplePrompt('Render an animated Mandelbrot fractal set with color cycling')}>Mandelbrot set</button>
      </div>
    </div>
  {:else}
    <!-- Simulation iframe -->
    <iframe
      bind:this={iframeEl}
      class="sim-iframe"
      sandbox="allow-scripts"
      title="AI2Sim Simulation Canvas"
      srcdoc=""
    ></iframe>

    <!-- Streaming overlay -->
    {#if $simulationStatus === 'fetching-params' || $simulationStatus === 'streaming'}
      <div class="loading-overlay">
        <div class="loading-spinner"></div>
        <span class="loading-text">{statusText}</span>
      </div>
    {/if}

    <!-- Stream indicator -->
    {#if $simulationStatus === 'streaming'}
      <div class="stream-badge">
        <span class="dot dot-purple animate-pulse"></span>
        Generating
      </div>
    {/if}
  {/if}
</div>

<style>
  .canvas-container {
    position: relative;
    width: 100%;
    height: 100%;
    border-radius: var(--radius-lg);
    overflow: hidden;
    background: #07070f;
    border: 1px solid var(--border);
  }

  .sim-iframe {
    width: 100%;
    height: 100%;
    border: none;
    display: block;
    border-radius: var(--radius-lg);
  }

  /* Empty state */
  .empty-state {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--sp-md);
    padding: var(--sp-xl);
    text-align: center;
    animation: fadeIn 0.6s var(--ease-out);
  }

  .empty-icon {
    width: 80px;
    height: 80px;
    opacity: 0.7;
    filter: drop-shadow(0 0 20px rgba(138,99,255,0.4));
    animation: spin 20s linear infinite;
  }

  .empty-state h3 {
    font-size: 1.2rem;
    color: var(--text-primary);
    margin-top: var(--sp-sm);
  }

  .empty-state p {
    font-size: 0.9rem;
    color: var(--text-muted);
    max-width: 360px;
    line-height: 1.6;
  }

  .example-prompts {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--sp-sm);
    justify-content: center;
    margin-top: var(--sp-sm);
  }

  .example-label {
    font-size: 0.8rem;
    color: var(--text-muted);
  }

  .example-chip {
    padding: 5px 12px;
    border-radius: var(--radius-full);
    border: 1px solid rgba(138,99,255,0.25);
    background: rgba(138,99,255,0.08);
    color: var(--accent-primary);
    font-size: 0.8rem;
    font-family: var(--font-sans);
    cursor: pointer;
    transition: all 150ms var(--ease-out);
  }

  .example-chip:hover {
    background: rgba(138,99,255,0.18);
    border-color: rgba(138,99,255,0.5);
    transform: translateY(-1px);
  }

  /* Loading overlay */
  .loading-overlay {
    position: absolute;
    inset: 0;
    background: rgba(7, 7, 15, 0.85);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--sp-md);
    backdrop-filter: blur(4px);
    border-radius: var(--radius-lg);
  }

  .loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(138,99,255,0.2);
    border-top-color: var(--accent-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  .loading-text {
    color: var(--text-secondary);
    font-size: 0.9rem;
    font-weight: 500;
  }

  /* Stream badge */
  .stream-badge {
    position: absolute;
    top: var(--sp-md);
    right: var(--sp-md);
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    padding: 4px 12px;
    background: rgba(7,7,15,0.8);
    border: 1px solid rgba(138,99,255,0.3);
    border-radius: var(--radius-full);
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--accent-primary);
    backdrop-filter: blur(8px);
  }
</style>
