<script lang="ts">
  import ParameterSlider from './ParameterSlider.svelte';
  import type { SimulationEntry } from '$lib/stores/simulation';
  import {
    paramSchema,
    paramValues,
    simulationStatus,
    simulationCode,
    isPlaying,
    isLoading,
    hasSimulation,
    addToHistory,
    currentPrompt,
    simulationHistory,
  } from '$lib/stores/simulation';

  let collapsed = $state(false);
  let activeTab: 'params' | 'history' = $state('params');

  function togglePlay() {
    isPlaying.update((v) => !v);
  }

  function resetParams() {
    paramValues.update((vals) => {
      const reset = { ...vals };
      for (const p of $paramSchema) {
        reset[p.name] = p.default;
      }
      return reset;
    });
  }

  function downloadCode() {
    const code = $simulationCode;
    if (!code) return;
    const blob = new Blob([code], { type: 'text/javascript' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `ai2sim-${Date.now()}.js`;
    a.click();
    URL.revokeObjectURL(url);
  }

  function loadHistory(entry: SimulationEntry) {
    currentPrompt.set(entry.prompt);
    paramSchema.set(entry.paramSchema);
    paramValues.set({ ...entry.paramValues });
    simulationCode.set(entry.code);
    simulationStatus.set('ready');
    activeTab = 'params';
  }

  function formatTime(date: Date | string): string {
    const d = typeof date === 'string' ? new Date(date) : date;
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  // Param count
  let paramCount = $derived($paramSchema.length);
  let historyCount = $derived($simulationHistory.length);
</script>

<aside class="sidebar" class:collapsed>
  <!-- Toggle Button -->
  <button
    class="collapse-btn"
    onclick={() => (collapsed = !collapsed)}
    aria-label={collapsed ? 'Expand panel' : 'Collapse panel'}
    data-tooltip={collapsed ? 'Expand' : 'Collapse'}
  >
    <svg viewBox="0 0 16 16" fill="currentColor" class:flipped={collapsed}>
      <path d="M9.5 12.5l-4-4.5 4-4.5"/>
    </svg>
  </button>

  {#if !collapsed}
  <div class="sidebar-content">
    <!-- Header -->
    <div class="sidebar-header">
      <h2 class="sidebar-title">Controls</h2>
      {#if $hasSimulation}
        <div class="playback-buttons">
          <button
            class="btn btn-ghost icon-btn"
            onclick={togglePlay}
            data-tooltip={$isPlaying ? 'Pause' : 'Play'}
          >
            {#if $isPlaying}
              <!-- Pause icon -->
              <svg viewBox="0 0 16 16" fill="currentColor" width="14">
                <rect x="3" y="2" width="4" height="12" rx="1"/>
                <rect x="9" y="2" width="4" height="12" rx="1"/>
              </svg>
            {:else}
              <!-- Play icon -->
              <svg viewBox="0 0 16 16" fill="currentColor" width="14">
                <path d="M3 2.5l10 5.5-10 5.5V2.5z"/>
              </svg>
            {/if}
          </button>
          <button
            class="btn btn-ghost icon-btn"
            onclick={downloadCode}
            data-tooltip="Download JS"
            aria-label="Download simulation JavaScript"
          >
            <svg viewBox="0 0 16 16" fill="currentColor" width="14">
              <path d="M8 12l-4.5-4.5L4.9 6l2.6 2.6V2h1v6.6L11.1 6l1.4 1.5L8 12z"/>
              <rect x="2" y="13" width="12" height="1.5" rx="0.75"/>
            </svg>
          </button>
        </div>
      {/if}
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button
        class="tab"
        class:active={activeTab === 'params'}
        onclick={() => (activeTab = 'params')}
      >
        Parameters
        {#if paramCount > 0}
          <span class="tab-count">{paramCount}</span>
        {/if}
      </button>
      <button
        class="tab"
        class:active={activeTab === 'history'}
        onclick={() => (activeTab = 'history')}
      >
        History
        {#if historyCount > 0}
          <span class="tab-count">{historyCount}</span>
        {/if}
      </button>
    </div>

    <!-- Tab Content -->
    {#if activeTab === 'params'}
      <div class="params-panel">
        {#if $paramSchema.length === 0}
          <div class="empty-params">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32">
              <path d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"/>
            </svg>
            <p>Parameters will appear<br/>after you submit a prompt</p>
          </div>
        {:else}
          <div class="params-list">
            {#each $paramSchema as param (param.name)}
              <ParameterSlider {param} disabled={$isLoading} />
            {/each}
          </div>

          <div class="params-footer">
            <button
              class="btn btn-ghost w-full"
              onclick={resetParams}
              disabled={$isLoading}
            >
              <svg viewBox="0 0 16 16" fill="currentColor" width="14">
                <path d="M13 8A5 5 0 112 8h1.5a3.5 3.5 0 109.5 0H11l2.5-3L16 8h-3z"/>
              </svg>
              Reset to Defaults
            </button>
          </div>
        {/if}
      </div>
    {:else}
      <!-- History tab -->
      <div class="history-panel">
        {#if $simulationHistory.length === 0}
          <div class="empty-params">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32">
              <path d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p>Completed simulations<br/>will appear here</p>
          </div>
        {:else}
          <div class="history-list">
            {#each $simulationHistory as entry (entry.id)}
              <div class="history-card">
                <div class="history-card-header">
                  <span class="history-prompt">{entry.prompt}</span>
                  <span class="history-time">{formatTime(entry.timestamp)}</span>
                </div>
                <div class="history-card-meta">
                  <span class="badge badge-primary">{entry.paramSchema.length} params</span>
                  <button class="btn btn-ghost history-load-btn" onclick={() => loadHistory(entry)}>
                    Load
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </div>
  {/if}
</aside>

<style>
  .sidebar {
    position: relative;
    width: 280px;
    min-width: 280px;
    height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--bg-surface);
    border-left: 1px solid var(--border);
    transition: width var(--dur-normal) var(--ease-out),
                min-width var(--dur-normal) var(--ease-out);
    overflow: hidden;
  }

  .sidebar.collapsed {
    width: 40px;
    min-width: 40px;
  }

  .collapse-btn {
    position: absolute;
    top: 50%;
    left: -12px;
    transform: translateY(-50%);
    width: 24px;
    height: 48px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
    cursor: pointer;
    z-index: 10;
    transition: color var(--dur-fast), background var(--dur-fast);
  }

  .collapse-btn:hover { color: var(--text-primary); background: var(--bg-glass-hover); }
  .collapse-btn svg { width: 12px; transition: transform var(--dur-normal) var(--ease-out); }
  .collapse-btn svg.flipped { transform: scaleX(-1); }

  .sidebar-content {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    animation: fadeIn 0.2s var(--ease-out);
  }

  .sidebar-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--sp-md) var(--sp-md) var(--sp-sm);
    border-bottom: 1px solid var(--border);
  }

  .sidebar-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .playback-buttons {
    display: flex;
    gap: 4px;
  }

  .icon-btn {
    padding: 6px 8px;
    min-width: unset;
  }

  /* Tabs */
  .tabs {
    display: flex;
    border-bottom: 1px solid var(--border);
    padding: 0 var(--sp-md);
    gap: var(--sp-md);
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 10px 0;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-muted);
    font-size: 0.85rem;
    font-weight: 500;
    font-family: var(--font-sans);
    cursor: pointer;
    transition: color var(--dur-fast), border-color var(--dur-fast);
    margin-bottom: -1px;
  }

  .tab.active {
    color: var(--accent-primary);
    border-bottom-color: var(--accent-primary);
  }

  .tab-count {
    background: rgba(138,99,255,0.15);
    color: var(--accent-primary);
    border-radius: var(--radius-full);
    font-size: 0.7rem;
    padding: 1px 6px;
    font-weight: 700;
  }

  /* Params panel */
  .params-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .empty-params {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--sp-md);
    padding: var(--sp-xl);
    color: var(--text-muted);
    text-align: center;
    flex: 1;
  }

  .empty-params p {
    font-size: 0.85rem;
    color: var(--text-muted);
    line-height: 1.6;
  }

  .params-list {
    flex: 1;
    overflow-y: auto;
    padding: var(--sp-md);
    display: flex;
    flex-direction: column;
    gap: var(--sp-sm);
  }

  .params-footer {
    padding: var(--sp-md);
    border-top: 1px solid var(--border);
  }

  .history-panel {
    flex: 1;
    overflow-y: auto;
    padding: var(--sp-md);
  }

  .history-list {
    display: flex;
    flex-direction: column;
    gap: var(--sp-sm);
  }

  .history-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px;
    background: var(--bg-glass);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    transition: background var(--dur-fast), border-color var(--dur-fast);
  }

  .history-card:hover {
    background: var(--bg-glass-hover);
    border-color: rgba(138, 99, 255, 0.25);
  }

  .history-card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 8px;
  }

  .history-prompt {
    font-size: 0.82rem;
    font-weight: 500;
    color: var(--text-primary);
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .history-time {
    font-size: 0.7rem;
    color: var(--text-muted);
    font-family: var(--font-mono);
    white-space: nowrap;
  }

  .history-card-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .history-load-btn {
    padding: 4px 10px;
    font-size: 0.75rem;
    border-radius: var(--radius-sm);
    color: var(--accent-primary);
    border-color: rgba(138, 99, 255, 0.3);
  }

  .history-load-btn:hover {
    background: rgba(138, 99, 255, 0.15);
    color: var(--text-primary);
  }
</style>
