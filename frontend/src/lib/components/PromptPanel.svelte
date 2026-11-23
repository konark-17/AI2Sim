<script lang="ts">
  import {
    currentPrompt,
    paramSchema,
    paramValues,
    simulationCode,
    simulationStatus,
    streamProgress,
    errorMessage,
    isLoading,
    resetSimulation,
    initParamValues,
    addToHistory,
  } from '$lib/stores/simulation';
  import { fetchParameters, openSimulationStream } from '$lib/api';

  // Example prompts by category
  const examples = [
    { label: 'Double Pendulum', prompt: 'Simulate a double pendulum with adjustable arm lengths, masses, and gravity. Show the chaotic trajectory trail.' },
    { label: 'Boid Flocking', prompt: 'Simulate boid flocking behavior with adjustable separation, alignment, and cohesion forces.' },
    { label: 'Reaction Diffusion', prompt: 'Simulate Gray-Scott reaction diffusion with tunable feed rate and kill rate parameters.' },
    { label: 'N-Body Gravity', prompt: 'Simulate N gravitational bodies orbiting with adjustable particle count, gravity strength, and initial speed.' },
    { label: 'Fourier Waves', prompt: 'Visualize Fourier series approximation of a square wave with adjustable harmonic count and animation speed.' },
    { label: 'Gas Molecules', prompt: 'Simulate ideal gas molecules with adjustable temperature, particle count, and molecular mass.' },
    { label: 'Mandelbrot Set', prompt: 'Render an animated Mandelbrot set with adjustable zoom and color cycle parameters.' },
    { label: 'Predator–Prey', prompt: 'Simulate predator–prey ecosystem dynamics with adjustable population sizes, speed, and reproduction rates.' },
  ];

  let promptInput = $state('');
  let activeEventSource: EventSource | null = null;
  let charCount = $derived(promptInput.length);
  let showExamples = $state(false);

  // Sync with currentPrompt if updated externally (example chips, history reload)
  $effect(() => {
    if ($currentPrompt && $currentPrompt !== promptInput && $simulationStatus === 'idle') {
      promptInput = $currentPrompt;
    }
  });

  async function submit() {
    const prompt = promptInput.trim();
    if (!prompt || $isLoading) return;

    // Cancel any active stream
    activeEventSource?.close();
    activeEventSource = null;

    resetSimulation();
    currentPrompt.set(prompt);
    simulationStatus.set('fetching-params');

    try {
      // 1. Fetch parameter schema
      const { parameters } = await fetchParameters(prompt);
      paramSchema.set(parameters);
      initParamValues(parameters);

      // 2. Start streaming simulation code
      simulationStatus.set('streaming');
      streamProgress.set(0);

      const params = $paramValues;
      const es = openSimulationStream(prompt, params);
      activeEventSource = es;

      let accumulatedCode = '';
      let tokenCount = 0;

      es.onmessage = (event) => {
        let token = '';
        try {
          const parsed = JSON.parse(event.data);
          token = (typeof parsed === 'object' && parsed !== null && parsed.token !== undefined)
            ? parsed.token
            : event.data;
        } catch {
          token = event.data;
        }
        accumulatedCode += token;
        tokenCount++;
        simulationCode.set(accumulatedCode);
        // Rough progress estimate — 4096 tokens max
        streamProgress.set(Math.min((tokenCount / 300) * 100, 95));
      };

      es.addEventListener('done', () => {
        es.close();
        activeEventSource = null;
        simulationStatus.set('ready');
        streamProgress.set(100);
        addToHistory(prompt, accumulatedCode, $paramSchema, $paramValues);
      });

      es.addEventListener('error', (event) => {
        const msgEvent = event as MessageEvent;
        let errText = 'Simulation generation failed.';
        if (msgEvent.data) {
          try {
            const parsed = JSON.parse(msgEvent.data);
            errText = parsed.error || msgEvent.data;
          } catch {
            errText = msgEvent.data;
          }
        }
        errorMessage.set(errText);
        es.close();
        activeEventSource = null;
        if ($simulationStatus === 'streaming') {
          simulationStatus.set('error');
        }
      });

      es.onerror = () => {
        es.close();
        activeEventSource = null;
        if ($simulationStatus === 'streaming') {
          simulationStatus.set('error');
          errorMessage.set('Connection to backend lost. Is the server running?');
        }
      };

    } catch (err) {
      simulationStatus.set('error');
      errorMessage.set(err instanceof Error ? err.message : 'Unknown error');
    }
  }

  function cancel() {
    activeEventSource?.close();
    activeEventSource = null;
    resetSimulation();
  }

  function useExample(ex: typeof examples[0]) {
    promptInput = ex.prompt;
    showExamples = false;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      submit();
    }
  }
</script>

<div class="prompt-panel">
  <!-- Top bar with logo and status -->
  <div class="panel-top">
    <div class="logo">
      <span class="logo-icon">⚛</span>
      <span class="logo-text">
        AI<span class="gradient-text">2</span>Sim
      </span>
    </div>

    <div class="status-area">
      {#if $simulationStatus === 'streaming'}
        <div class="stream-status">
          <div class="mini-spinner"></div>
          <span>Generating</span>
          <span class="pct">{Math.round($streamProgress)}%</span>
        </div>
      {:else if $simulationStatus === 'ready'}
        <span class="badge badge-success">
          <span class="dot dot-green"></span> Ready
        </span>
      {:else if $simulationStatus === 'error'}
        <span class="badge badge-error">Error</span>
      {/if}
    </div>
  </div>

  <!-- Prompt input area -->
  <div class="input-area">
    <div class="textarea-wrapper">
      <textarea
        id="prompt-input"
        class="input prompt-textarea"
        placeholder="Describe a physical system, mathematical phenomenon, or biological process to simulate…"
        bind:value={promptInput}
        onkeydown={handleKeydown}
        disabled={$isLoading}
        rows={3}
        maxlength={500}
      ></textarea>
      <div class="textarea-footer">
        <button
          class="examples-toggle"
          onclick={() => (showExamples = !showExamples)}
          type="button"
        >
          <svg viewBox="0 0 16 16" fill="currentColor" width="12">
            <path d="M8 1a7 7 0 100 14A7 7 0 008 1zm.75 4.75a.75.75 0 00-1.5 0v3.5h-2a.75.75 0 000 1.5h3.5V5.75z"/>
          </svg>
          Examples
        </button>
        <span class="char-count">{charCount}/500</span>
      </div>
    </div>

    <!-- Action buttons -->
    <div class="action-row">
      {#if $isLoading}
        <button class="btn btn-ghost" onclick={cancel} id="cancel-btn">
          <svg viewBox="0 0 16 16" fill="currentColor" width="14">
            <path d="M8 1a7 7 0 100 14A7 7 0 008 1zm3.5 9.8L5.2 4.5l1.4-1.4 6.3 6.3-1.4 1.4z"/>
          </svg>
          Cancel
        </button>
      {/if}
      <button
        class="btn btn-primary"
        onclick={submit}
        disabled={!promptInput.trim() || $isLoading}
        id="generate-btn"
        style="flex: 1;"
      >
        {#if $simulationStatus === 'fetching-params'}
          <span class="mini-spinner white"></span>
          Analyzing…
        {:else if $simulationStatus === 'streaming'}
          <span class="mini-spinner white"></span>
          Generating…
        {:else}
          <svg viewBox="0 0 16 16" fill="currentColor" width="14">
            <path d="M2 8l6-6 6 6M8 2v12"/>
          </svg>
          Generate Simulation
        {/if}
      </button>
    </div>

    <p class="hint">Press <kbd>Ctrl</kbd>+<kbd>Enter</kbd> to generate</p>
  </div>

  <!-- Example prompts dropdown -->
  {#if showExamples}
    <div class="examples-dropdown" role="listbox" aria-label="Example prompts">
      <div class="examples-header">
        <span>Example Simulations</span>
        <button onclick={() => (showExamples = false)} class="close-examples">✕</button>
      </div>
      <div class="examples-grid">
        {#each examples as ex (ex.label)}
          <button
            class="example-card"
            onclick={() => useExample(ex)}
            role="option"
            aria-selected="false"
          >
            <span class="example-card-label">{ex.label}</span>
            <span class="example-card-preview">{ex.prompt.slice(0, 60)}…</span>
          </button>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Error display -->
  {#if $simulationStatus === 'error' && $errorMessage}
    <div class="error-banner" role="alert">
      <svg viewBox="0 0 16 16" fill="currentColor" width="14">
        <path d="M8 1a7 7 0 100 14A7 7 0 008 1zm-.75 4a.75.75 0 011.5 0v3a.75.75 0 01-1.5 0V5zm.75 6.5a1 1 0 110-2 1 1 0 010 2z"/>
      </svg>
      {$errorMessage}
    </div>
  {/if}

  <!-- Progress bar (visible during streaming) -->
  {#if $simulationStatus === 'streaming'}
    <div class="progress-track">
      <div
        class="progress-fill"
        style:width="{$streamProgress}%"
      ></div>
    </div>
  {/if}
</div>

<style>
  .prompt-panel {
    display: flex;
    flex-direction: column;
    gap: var(--sp-md);
    padding: var(--sp-md);
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border);
    position: relative;
  }

  /* Logo */
  .panel-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .logo-icon {
    font-size: 1.4rem;
    filter: drop-shadow(0 0 8px rgba(138,99,255,0.6));
  }

  .logo-text {
    font-size: 1.25rem;
    font-weight: 700;
    letter-spacing: -0.03em;
    color: var(--text-primary);
  }

  /* Status */
  .stream-status {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.82rem;
    color: var(--text-secondary);
  }

  .pct {
    font-family: var(--font-mono);
    font-size: 0.78rem;
    color: var(--accent-primary);
  }

  .badge-error {
    background: rgba(248,113,113,0.12);
    color: var(--accent-error);
    border: 1px solid rgba(248,113,113,0.2);
  }

  /* Input area */
  .input-area {
    display: flex;
    flex-direction: column;
    gap: var(--sp-sm);
  }

  .textarea-wrapper {
    position: relative;
  }

  .prompt-textarea {
    min-height: 90px;
    max-height: 200px;
    line-height: 1.6;
    padding-bottom: 36px;
    font-size: 0.92rem;
  }

  .textarea-footer {
    position: absolute;
    bottom: 10px;
    left: 12px;
    right: 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    pointer-events: none;
  }

  .examples-toggle {
    pointer-events: all;
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 0.75rem;
    font-family: var(--font-sans);
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 4px;
    transition: color var(--dur-fast);
    padding: 0;
  }
  .examples-toggle:hover { color: var(--accent-primary); }

  .char-count {
    font-family: var(--font-mono);
    font-size: 0.72rem;
    color: var(--text-muted);
  }

  .action-row {
    display: flex;
    gap: var(--sp-sm);
  }

  .hint {
    font-size: 0.72rem;
    color: var(--text-muted);
    text-align: center;
  }

  kbd {
    display: inline-block;
    padding: 1px 5px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 3px;
    font-family: var(--font-mono);
    font-size: 0.7rem;
    color: var(--text-secondary);
  }

  /* Examples dropdown */
  .examples-dropdown {
    position: absolute;
    top: 100%;
    left: var(--sp-md);
    right: var(--sp-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    z-index: 50;
    overflow: hidden;
    animation: fadeIn 0.2s var(--ease-out);
  }

  .examples-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--sp-sm) var(--sp-md);
    border-bottom: 1px solid var(--border);
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-secondary);
  }

  .close-examples {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 0.85rem;
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    transition: color var(--dur-fast), background var(--dur-fast);
  }
  .close-examples:hover { color: var(--text-primary); background: var(--bg-glass); }

  .examples-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1px;
    background: var(--border);
    max-height: 280px;
    overflow-y: auto;
  }

  .example-card {
    display: flex;
    flex-direction: column;
    gap: 3px;
    padding: var(--sp-sm) var(--sp-md);
    background: var(--bg-elevated);
    border: none;
    text-align: left;
    cursor: pointer;
    transition: background var(--dur-fast);
  }
  .example-card:hover { background: var(--bg-glass-hover); }

  .example-card-label {
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .example-card-preview {
    font-size: 0.73rem;
    color: var(--text-muted);
    line-height: 1.4;
  }

  /* Error banner */
  .error-banner {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-sm) var(--sp-md);
    background: rgba(248,113,113,0.08);
    border: 1px solid rgba(248,113,113,0.2);
    border-radius: var(--radius-md);
    color: var(--accent-error);
    font-size: 0.85rem;
    animation: fadeIn 0.3s var(--ease-out);
  }

  /* Progress bar */
  .progress-track {
    height: 2px;
    background: rgba(255,255,255,0.06);
    border-radius: var(--radius-full);
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: var(--gradient-brand);
    border-radius: var(--radius-full);
    transition: width 0.3s var(--ease-out);
  }

  /* Spinners */
  .mini-spinner {
    display: inline-block;
    width: 14px;
    height: 14px;
    border: 2px solid rgba(138,99,255,0.3);
    border-top-color: var(--accent-primary);
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
    flex-shrink: 0;
  }
  .mini-spinner.white {
    border-color: rgba(255,255,255,0.3);
    border-top-color: white;
  }
</style>
