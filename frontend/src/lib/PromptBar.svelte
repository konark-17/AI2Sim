<script>
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();
  export let isGenerating = false;
  let prompt = "";

  function handleSubmit() {
    if (!prompt.trim() || isGenerating) return;
    dispatch('submit', { prompt });
    prompt = "";
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  }
</script>

<div class="prompt-container">
  <div class="input-wrapper" class:generating={isGenerating}>
    <textarea 
      bind:value={prompt}
      on:keydown={handleKeyDown}
      placeholder={isGenerating ? "Generating simulation..." : "Describe the simulation you want to create... (e.g. 'A particle system with gravity')"}
      disabled={isGenerating}
      rows="1"
    ></textarea>
    
    <button on:click={handleSubmit} disabled={!prompt.trim() || isGenerating}>
      {#if isGenerating}
        <span class="loader"></span>
      {:else}
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
      {/if}
    </button>
  </div>
</div>

<style>
  .prompt-container {
    position: absolute;
    bottom: 30px;
    left: 50%;
    transform: translateX(-50%);
    width: 60%;
    max-width: 800px;
    min-width: 400px;
    z-index: 20;
  }

  .input-wrapper {
    display: flex;
    align-items: center;
    background: rgba(26, 26, 36, 0.85);
    backdrop-filter: blur(12px);
    border: 1px solid var(--border-color);
    border-radius: 24px;
    padding: 8px 8px 8px 24px;
    box-shadow: var(--shadow-md), 0 0 20px rgba(99, 102, 241, 0.1);
    transition: all 0.3s ease;
  }

  .input-wrapper:focus-within {
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 1px var(--accent-primary), 0 0 30px rgba(99, 102, 241, 0.2);
  }

  .input-wrapper.generating {
    border-color: var(--accent-secondary);
    animation: pulse 2s infinite;
  }

  textarea {
    flex: 1;
    background: transparent;
    border: none;
    color: var(--text-primary);
    font-family: inherit;
    font-size: 1rem;
    resize: none;
    outline: none;
    padding-top: 8px; /* Alignment tweak */
  }

  textarea::placeholder {
    color: var(--text-secondary);
  }

  button {
    background: var(--accent-primary);
    color: white;
    border: none;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s;
    flex-shrink: 0;
  }

  button:hover:not(:disabled) {
    background: #4f46e5;
    transform: scale(1.05);
  }

  button:disabled {
    background: var(--bg-tertiary);
    color: var(--text-secondary);
    cursor: not-allowed;
  }

  /* Loader */
  .loader {
    width: 18px;
    height: 18px;
    border: 2px solid rgba(255,255,255,0.3);
    border-radius: 50%;
    border-top-color: white;
    animation: spin 1s ease-in-out infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  @keyframes pulse {
    0% { box-shadow: 0 0 0 0 rgba(236, 72, 153, 0.4); }
    70% { box-shadow: 0 0 0 10px rgba(236, 72, 153, 0); }
    100% { box-shadow: 0 0 0 0 rgba(236, 72, 153, 0); }
  }
</style>
