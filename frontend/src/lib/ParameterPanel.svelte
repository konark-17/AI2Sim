<script>
  export let parameters = {};
  export let parameterDefs = []; // [{ name: 'speed', label: 'Speed', min: 0.1, max: 5, step: 0.1 }]
</script>

<div class="panel-container">
  <div class="panel-header">
    <h3>Simulation Parameters</h3>
  </div>
  
  <div class="parameters-list">
    {#if parameterDefs.length === 0}
      <p class="empty-state">No parameters available for current simulation.</p>
    {:else}
      {#each parameterDefs as def}
        <div class="parameter-item">
          <div class="param-header">
            <label for={def.name}>{def.label}</label>
            <span class="value">{parameters[def.name]}</span>
          </div>
          <input 
            type="range" 
            id={def.name}
            min={def.min} 
            max={def.max} 
            step={def.step || 1}
            bind:value={parameters[def.name]}
          />
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .panel-container {
    width: 300px;
    height: 100%;
    background-color: var(--bg-secondary);
    border-left: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    box-shadow: -5px 0 25px rgba(0, 0, 0, 0.5);
    z-index: 10;
  }

  .panel-header {
    padding: 20px;
    border-bottom: 1px solid var(--border-color);
  }

  h3 {
    margin: 0;
    font-size: 1.1rem;
    color: var(--text-primary);
    font-weight: 600;
  }

  .parameters-list {
    padding: 20px;
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .empty-state {
    color: var(--text-secondary);
    font-size: 0.9rem;
    text-align: center;
    margin-top: 20px;
  }

  .parameter-item {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .param-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  label {
    font-size: 0.9rem;
    color: var(--text-secondary);
    font-weight: 500;
  }

  .value {
    font-size: 0.85rem;
    color: var(--accent-primary);
    background: rgba(99, 102, 241, 0.1);
    padding: 2px 6px;
    border-radius: 4px;
    font-family: monospace;
  }

  /* Custom Range Slider Styling */
  input[type=range] {
    -webkit-appearance: none;
    width: 100%;
    background: transparent;
  }

  input[type=range]::-webkit-slider-thumb {
    -webkit-appearance: none;
    height: 16px;
    width: 16px;
    border-radius: 50%;
    background: var(--accent-primary);
    cursor: pointer;
    margin-top: -6px;
    box-shadow: 0 0 10px rgba(99, 102, 241, 0.5);
    transition: transform 0.1s;
  }

  input[type=range]::-webkit-slider-thumb:hover {
    transform: scale(1.2);
  }

  input[type=range]::-webkit-slider-runnable-track {
    width: 100%;
    height: 4px;
    cursor: pointer;
    background: var(--bg-tertiary);
    border-radius: 2px;
  }
</style>
