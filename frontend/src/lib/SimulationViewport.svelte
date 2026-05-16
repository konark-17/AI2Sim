<script>
  import { onMount } from 'svelte';

  export let parameters = {};
  export let simulationCode = ""; // In the future, this would be generated code
  
  let canvas;
  let ctx;
  let animationFrameId;
  let particles = [];

  // Mock simulation: a particle system
  function initSimulation() {
    if (!canvas) return;
    ctx = canvas.getContext('2d');
    resize();
    createParticles();
    animate();
  }

  function resize() {
    if (canvas) {
      canvas.width = canvas.parentElement.clientWidth;
      canvas.height = canvas.parentElement.clientHeight;
    }
  }

  function createParticles() {
    particles = [];
    const count = parameters.count || 50;
    for (let i = 0; i < count; i++) {
      particles.push({
        x: Math.random() * canvas.width,
        y: Math.random() * canvas.height,
        vx: (Math.random() - 0.5) * 2,
        vy: (Math.random() - 0.5) * 2,
        radius: Math.random() * 3 + 1,
        color: `hsl(${Math.random() * 360}, 70%, 60%)`
      });
    }
  }

  function animate() {
    if (!ctx) return;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    
    const speedMultiplier = parameters.speed || 1;
    const sizeMultiplier = parameters.size || 1;

    // React to count changes
    if (particles.length !== (parameters.count || 50)) {
        createParticles();
    }

    particles.forEach(p => {
      p.x += p.vx * speedMultiplier;
      p.y += p.vy * speedMultiplier;

      // Bounce off walls
      if (p.x < 0 || p.x > canvas.width) p.vx *= -1;
      if (p.y < 0 || p.y > canvas.height) p.vy *= -1;

      ctx.beginPath();
      ctx.arc(p.x, p.y, p.radius * sizeMultiplier, 0, Math.PI * 2);
      ctx.fillStyle = p.color;
      ctx.fill();
    });

    animationFrameId = requestAnimationFrame(animate);
  }

  onMount(() => {
    initSimulation();
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
  {:else}
    <canvas bind:this={canvas}></canvas>
  {/if}
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
