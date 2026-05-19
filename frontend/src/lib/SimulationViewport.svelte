<script>
  import { onMount, afterUpdate } from 'svelte';

  export let parameters = {};
  export let simulationCode = "";
  
  let canvas;
  let ctx;
  let animationFrameId;
  let simulationFn = null;
  let lastTime = 0;

  // Build a simulation function from the received code string
  function buildSimulationFn(code) {
    if (!code) return null;
    
    try {
      // The code is a function body that receives (ctx, canvas, parameters, deltaTime)
      return new Function('ctx', 'canvas', 'parameters', 'deltaTime', code);
    } catch (e) {
      console.error("Failed to compile simulation code:", e);
      return null;
    }
  }

  // Re-compile when simulationCode changes
  $: if (simulationCode) {
    // Reset persistent state for the new simulation
    window.simState = null;
    simulationFn = buildSimulationFn(simulationCode);
    
    if (simulationFn && canvas) {
      lastTime = performance.now();
      // Restart animation loop
      cancelAnimationFrame(animationFrameId);
      animate(performance.now());
    }
  }

  function resize() {
    if (canvas) {
      canvas.width = canvas.parentElement.clientWidth;
      canvas.height = canvas.parentElement.clientHeight;
    }
  }

  function animate(timestamp) {
    if (!ctx || !simulationFn) return;
    
    const deltaTime = timestamp - lastTime;
    lastTime = timestamp;

    try {
      simulationFn(ctx, canvas, parameters, deltaTime);
    } catch (e) {
      console.error("Simulation runtime error:", e);
      // Don't kill the loop — the user might adjust parameters to fix it
    }

    animationFrameId = requestAnimationFrame(animate);
  }

  onMount(() => {
    if (canvas) {
      ctx = canvas.getContext('2d');
      resize();
    }
    window.addEventListener('resize', resize);
    return () => {
      window.removeEventListener('resize', resize);
      cancelAnimationFrame(animationFrameId);
    };
  });
</script>

<div class="viewport-container">
  {#if !simulationCode}
    <div class="placeholder">
      <div class="icon">✨</div>
      <p>Enter a prompt below to generate a simulation</p>
    </div>
  {/if}
  <!-- Canvas is always present but hidden when no simulation -->
  <canvas bind:this={canvas} class:hidden={!simulationCode}></canvas>
</div>

<style>
  .viewport-container {
    width: 100%;
    height: 100%;
    background-color: var(--bg-primary);
    position: relative;
    overflow: hidden;
    border-radius: 12px;
    box-shadow: inset 0 0 20px rgba(0,0,0,0.5), var(--shadow-lg), 0 0 10px rgba(99, 102, 241, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  canvas {
    display: block;
    width: 100%;
    height: 100%;
  }

  canvas.hidden {
    display: none;
  }

  .placeholder {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    text-align: center;
    color: var(--text-secondary);
  }

  .icon {
    font-size: 3rem;
    margin-bottom: 1rem;
    animation: float 3s ease-in-out infinite;
  }

  @keyframes float {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-10px); }
  }
</style>
