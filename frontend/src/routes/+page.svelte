<script lang="ts">
  import '../app.css';
  import PromptPanel from '$lib/components/PromptPanel.svelte';
  import SimulationCanvas from '$lib/components/SimulationCanvas.svelte';
  import ControlSidebar from '$lib/components/ControlSidebar.svelte';
  import { checkHealth } from '$lib/api';
  import { onMount } from 'svelte';

  let backendOnline = $state<boolean | null>(null);
  let statusBannerText = $state<string>('');

  onMount(async () => {
    const health = await checkHealth();
    if (!health.ok) {
      backendOnline = false;
      statusBannerText = health.message || 'Backend offline — start the Go server on port 8080';
    } else if (health.ready === false) {
      backendOnline = false;
      statusBannerText = health.message || 'LLM not ready — check backend configuration';
    } else {
      backendOnline = true;
    }
  });
</script>

<svelte:head>
  <title>AI2Sim — AI-Powered Scientific Simulations</title>
  <meta name="description" content="Generate interactive scientific simulations from natural language prompts. Physics, math, biology, and chemistry simulations powered by local LLMs." />
  <meta property="og:title" content="AI2Sim" />
  <meta property="og:description" content="AI-powered simulation generator" />
  <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
</svelte:head>

<div class="app-shell">
  <!-- Backend status banner (only shown when offline or unconfigured) -->
  {#if backendOnline === false}
    <div class="status-banner offline">
      <svg viewBox="0 0 16 16" fill="currentColor" width="14">
        <path d="M8 1a7 7 0 100 14A7 7 0 008 1zm-.75 4a.75.75 0 011.5 0v3a.75.75 0 01-1.5 0V5zm.75 6.5a1 1 0 110-2 1 1 0 010 2z"/>
      </svg>
      {statusBannerText}
    </div>
  {/if}

  <!-- Main layout -->
  <div class="layout">
    <!-- Left: Prompt panel -->
    <aside class="left-panel">
      <PromptPanel />
    </aside>

    <!-- Center: Simulation canvas -->
    <main class="center-panel">
      <SimulationCanvas />
    </main>

    <!-- Right: Controls + sliders -->
    <ControlSidebar />
  </div>

  <!-- Ambient background glow -->
  <div class="bg-glow" aria-hidden="true"></div>
</div>

<style>
  .app-shell {
    display: flex;
    flex-direction: column;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    position: relative;
  }

  .bg-glow {
    position: fixed;
    top: -30%;
    left: 20%;
    width: 60%;
    height: 60%;
    background: radial-gradient(ellipse at center, rgba(138,99,255,0.06) 0%, transparent 70%);
    pointer-events: none;
    z-index: 0;
  }

  .status-banner {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 20px;
    font-size: 0.82rem;
    font-weight: 500;
    z-index: 10;
    position: relative;
  }

  .status-banner.offline {
    background: rgba(251,191,36,0.08);
    border-bottom: 1px solid rgba(251,191,36,0.2);
    color: var(--accent-warning);
  }

  .layout {
    display: flex;
    flex: 1;
    overflow: hidden;
    position: relative;
    z-index: 1;
  }

  /* Left panel — prompt + controls */
  .left-panel {
    width: 360px;
    min-width: 360px;
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border);
    background: var(--bg-surface);
    overflow: hidden;
  }

  /* Center panel — simulation */
  .center-panel {
    flex: 1;
    min-width: 0;
    padding: var(--sp-md);
    display: flex;
    align-items: stretch;
    background: var(--bg-base);
    overflow: hidden;
  }

  /* Responsive: stack vertically on small screens */
  @media (max-width: 900px) {
    .layout {
      flex-direction: column;
    }

    .left-panel {
      width: 100%;
      min-width: unset;
      height: auto;
      border-right: none;
      border-bottom: 1px solid var(--border);
    }
  }
</style>
