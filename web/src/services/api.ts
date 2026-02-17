const API_BASE = '/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    const msg = (err as { error?: string }).error ?? res.statusText;
    if (res.status === 404 && path.includes('system/summary')) {
      throw new Error(
        'Endpoint /api/system/summary não encontrado. Recompila e reinicia o backend (./dockscope) e garante que está a correr em :8080.'
      );
    }
    throw new Error(msg);
  }
  return res.json() as Promise<T>;
}

export type ContainerAction = 'start' | 'stop' | 'restart' | 'pause' | 'unpause';

export interface ContainerMemoryEntry {
  id: string;
  name: string;
  memory_usage: number;
  memory_percent?: number;
}

export interface ContainerMetricsEntry {
  id: string;
  name: string;
  cpu_percentage: number;
  memory_usage: number;
  memory_limit: number;
  memory_percent?: number;
}

export interface SystemSummary {
  containers_total: number;
  containers_running: number;
  containers_stopped: number;
  cpu_percent_total: number;
  memory_usage_bytes: number;
  memory_limit_bytes: number;
  images_count: number;
  volumes_count: number;
  top_containers_by_memory: ContainerMemoryEntry[];
  container_metrics: ContainerMetricsEntry[];
}

export const api = {
  getContainers: (all = false) =>
    request<import('../types/docker').Container[]>(
      `/containers${all ? '?all=true' : ''}`
    ),
  getImages: () => request<import('../types/docker').Image[]>(`/images`),
  getVolumes: () =>
    request<import('../types/docker').Volume[]>(`/volumes`),
  health: () => request<{ status: string }>('/health'),
  getSystemSummary: () =>
    request<SystemSummary>('/system/summary'),
  containerAction: (containerId: string, action: ContainerAction) =>
    request<{ ok: boolean }>(`/containers/${containerId}/action`, {
      method: 'POST',
      body: JSON.stringify({ action }),
    }),
};

export function getStatsWebSocketUrl(containerId: string): string {
  const base = window.location.origin.replace(/^http/, 'ws');
  return `${base}${API_BASE}/stats/${containerId}`;
}

const LOGS_WS_DEV_ORIGIN = 'ws://localhost:8080';

export function getLogsWebSocketUrl(containerId: string): string {
  if (import.meta.env.DEV && typeof window !== 'undefined') {
    return `${LOGS_WS_DEV_ORIGIN}${API_BASE}/logs/${containerId}`;
  }
  const base = window.location.origin.replace(/^http/, 'ws');
  return `${base}${API_BASE}/logs/${containerId}`;
}
