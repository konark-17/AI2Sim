// Svelte stores for AI2Sim simulation state

import { writable, derived } from 'svelte/store';

// ─── Types ────────────────────────────────────────────────────────────────────

export interface ParameterSchema {
  name: string;
  label: string;
  min: number;
  max: number;
  default: number;
  step: number;
  unit: string;
}

export type SimulationStatus =
  | 'idle'
  | 'fetching-params'
  | 'streaming'
  | 'ready'
  | 'error';

export interface SimulationEntry {
  id: string;
  prompt: string;
  code: string;
  timestamp: Date;
  paramSchema: ParameterSchema[];
  paramValues: Record<string, number>;
}

// ─── Stores ───────────────────────────────────────────────────────────────────

/** The user's current natural-language prompt */
export const currentPrompt = writable<string>('');

/** Inferred parameter schema from the backend */
export const paramSchema = writable<ParameterSchema[]>([]);

/** Current slider values, keyed by param name */
export const paramValues = writable<Record<string, number>>({});

/** The accumulated LLM-generated simulation JS code (grows during streaming) */
export const simulationCode = writable<string>('');

/** Current status of the simulation pipeline */
export const simulationStatus = writable<SimulationStatus>('idle');

/** Most recent error message */
export const errorMessage = writable<string>('');

/** Simulation history (last N runs) */
export const simulationHistory = writable<SimulationEntry[]>([]);

/** Whether the simulation panel is playing or paused */
export const isPlaying = writable<boolean>(true);

// ─── Derived ──────────────────────────────────────────────────────────────────

/** True while the pipeline is active (fetching or streaming) */
export const isLoading = derived(
  simulationStatus,
  ($s) => $s === 'fetching-params' || $s === 'streaming'
);

/** True when the simulation code is ready to render */
export const hasSimulation = derived(
  simulationStatus,
  ($s) => $s === 'ready' || $s === 'streaming'
);

/** Stream progress percentage (0-100) */
export const streamProgress = writable<number>(0);

// ─── Actions ──────────────────────────────────────────────────────────────────

/** Reset all simulation state to idle */
export function resetSimulation(): void {
  simulationCode.set('');
  simulationStatus.set('idle');
  errorMessage.set('');
  streamProgress.set(0);
  isPlaying.set(true);
}

/** Initialize param values from a schema (set all to defaults) */
export function initParamValues(schema: ParameterSchema[]): void {
  const vals: Record<string, number> = {};
  for (const p of schema) {
    vals[p.name] = p.default;
  }
  paramValues.set(vals);
}

/** Add a completed simulation to history */
export function addToHistory(
  prompt: string,
  code: string,
  schema: ParameterSchema[],
  values: Record<string, number>
): void {
  const entry: SimulationEntry = {
    id: crypto.randomUUID(),
    prompt,
    code,
    timestamp: new Date(),
    paramSchema: schema,
    paramValues: values,
  };

  simulationHistory.update((h) => {
    const updated = [entry, ...h];
    return updated.slice(0, 10); // Keep last 10
  });
}
