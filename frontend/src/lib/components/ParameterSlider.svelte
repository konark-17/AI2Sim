<script lang="ts">
  import type { ParameterSchema } from '$lib/stores/simulation';
  import { paramValues } from '$lib/stores/simulation';

  interface Props {
    param: ParameterSchema;
    disabled?: boolean;
  }

  let { param, disabled = false }: Props = $props();

  // Derive the current value reactively
  let value = $derived($paramValues[param.name] ?? param.default);

  function handleChange(e: Event) {
    const newVal = parseFloat((e.target as HTMLInputElement).value);
    paramValues.update((v) => ({ ...v, [param.name]: newVal }));
  }

  // Compute slider fill percentage for CSS variable
  let pct = $derived(
    ((value - param.min) / (param.max - param.min)) * 100
  );

  // Format the displayed value to reasonable precision
  let displayValue = $derived(
    param.step < 0.01
      ? value.toFixed(4)
      : param.step < 0.1
        ? value.toFixed(3)
        : param.step < 1
          ? value.toFixed(2)
          : value.toFixed(0)
  );
</script>

<div class="slider-wrapper" class:disabled>
  <div class="slider-header">
    <label class="slider-label" for={`slider-${param.name}`}>
      {param.label}
    </label>
    <span class="slider-value">
      <span class="value-number">{displayValue}</span>
      {#if param.unit}
        <span class="value-unit">{param.unit}</span>
      {/if}
    </span>
  </div>

  <div class="slider-row">
    <span class="range-label">{param.min}</span>
    <input
      id={`slider-${param.name}`}
      type="range"
      class="slider-track"
      min={param.min}
      max={param.max}
      step={param.step}
      {value}
      {disabled}
      style:--slider-pct="{pct}%"
      oninput={handleChange}
    />
    <span class="range-label">{param.max}</span>
  </div>
</div>

<style>
  .slider-wrapper {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px 14px;
    background: var(--bg-glass);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    transition: border-color 200ms var(--ease-out),
                background 200ms var(--ease-out);
  }

  .slider-wrapper:hover:not(.disabled) {
    border-color: rgba(138, 99, 255, 0.25);
    background: var(--bg-glass-hover);
  }

  .slider-wrapper.disabled {
    opacity: 0.45;
    pointer-events: none;
  }

  .slider-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .slider-label {
    font-size: 0.82rem;
    font-weight: 500;
    color: var(--text-secondary);
    cursor: pointer;
    user-select: none;
    letter-spacing: 0.02em;
    text-transform: uppercase;
  }

  .slider-value {
    display: flex;
    align-items: baseline;
    gap: 3px;
    font-family: var(--font-mono);
  }

  .value-number {
    font-size: 0.88rem;
    font-weight: 600;
    color: var(--accent-primary);
  }

  .value-unit {
    font-size: 0.72rem;
    color: var(--text-muted);
  }

  .slider-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .range-label {
    font-family: var(--font-mono);
    font-size: 0.7rem;
    color: var(--text-muted);
    min-width: 28px;
    text-align: center;
    user-select: none;
  }
</style>
