import { browser } from '$app/environment';

const DEFAULT_API_URL = browser
  ? (import.meta.env.VITE_API_URL ?? 'http://localhost:8080')
  : 'http://localhost:8080';

/** Returns the base API URL from env or default */
export function getApiUrl(): string {
  return DEFAULT_API_URL;
}

/** Fetch inferred parameter schema for a given prompt */
export async function fetchParameters(prompt: string) {
  const res = await fetch(`${getApiUrl()}/api/v1/parameters`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ prompt }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(err.error ?? `HTTP ${res.status}`);
  }
  return res.json() as Promise<{ parameters: import('./stores/simulation').ParameterSchema[] }>;
}

/** Opens an SSE stream for simulation code generation */
export function openSimulationStream(
  prompt: string,
  params: Record<string, number>
): EventSource {
  const url = new URL(`${getApiUrl()}/api/v1/simulate`);
  url.searchParams.set('prompt', prompt);
  url.searchParams.set('params', JSON.stringify(params));
  return new EventSource(url.toString());
}

export interface HealthStatus {
  ok: boolean;
  ready?: boolean;
  message?: string;
  provider?: string;
  model?: string;
}

/** Check backend health */
export async function checkHealth(): Promise<HealthStatus> {
  try {
    const res = await fetch(`${getApiUrl()}/api/v1/health`, {
      signal: AbortSignal.timeout(3000),
    });
    if (!res.ok) return { ok: false, message: `Server error HTTP ${res.status}` };
    const data = await res.json();
    return {
      ok: true,
      ready: data.ready ?? true,
      message: data.message ?? '',
      provider: data.provider ?? '',
      model: data.model ?? '',
    };
  } catch {
    return { ok: false, message: 'Backend offline — start the Go server on port 8080' };
  }
}
