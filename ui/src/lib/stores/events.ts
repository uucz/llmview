import { writable, derived, get } from 'svelte/store';

export interface APICall {
  id: string;
  provider: string;
  model: string;
  endpoint: string;
  streaming: boolean;
  status_code: number;
  started_at: number; // unix ms
  duration_ms: number;
  input_tokens: number;
  output_tokens: number;
  cost: number;
  error?: string;
  // Client-side state
  streamText: string;
  completed: boolean;
}

export interface APICallDetail {
  id: string;
  provider: string;
  model: string;
  endpoint: string;
  method: string;
  request_body: string;
  response_body: string;
  status_code: number;
  started_at: string;
  duration_ms: number;
  input_tokens: number;
  output_tokens: number;
  cost: number;
  streaming: boolean;
  error?: string;
}

export interface SessionStats {
  total_cost: number;
  total_tokens: number;
  request_count: number;
}

export interface Filters {
  providers: Set<string>;
  status: 'all' | 'success' | 'error';
  query: string;
}

export interface AppConfig {
  budget: number;
}

export const calls = writable<Map<string, APICall>>(new Map());
export const session = writable<SessionStats>({ total_cost: 0, total_tokens: 0, request_count: 0 });
export const connected = writable(false);
export const callDetails = writable<Map<string, APICallDetail>>(new Map());
export const filters = writable<Filters>({ providers: new Set(), status: 'all', query: '' });
export const appConfig = writable<AppConfig>({ budget: 0 });
export const budgetExceeded = writable(false);

// Unfiltered sorted calls (for empty-state check)
export const sortedCalls = derived(calls, ($calls) => {
  return Array.from($calls.values()).sort((a, b) => b.started_at - a.started_at);
});

// Available providers for filter chips
export const availableProviders = derived(calls, ($calls) => {
  const providers = new Set<string>();
  for (const call of $calls.values()) providers.add(call.provider);
  return Array.from(providers).sort();
});

// Filtered + sorted calls
export const filteredCalls = derived([calls, filters], ([$calls, $filters]) => {
  let list = Array.from($calls.values());

  if ($filters.providers.size > 0) {
    list = list.filter(c => $filters.providers.has(c.provider));
  }
  if ($filters.status === 'success') {
    list = list.filter(c => !c.completed || (c.status_code >= 200 && c.status_code < 400));
  } else if ($filters.status === 'error') {
    list = list.filter(c => c.completed && c.status_code >= 400);
  }
  if ($filters.query) {
    const q = $filters.query.toLowerCase();
    list = list.filter(c =>
      c.model.toLowerCase().includes(q) ||
      c.endpoint.toLowerCase().includes(q) ||
      c.provider.toLowerCase().includes(q) ||
      (c.error && c.error.toLowerCase().includes(q))
    );
  }

  return list.sort((a, b) => b.started_at - a.started_at);
});

let ws: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout>;

export function connectWS() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const url = `${protocol}//${window.location.host}/ws`;
  ws = new WebSocket(url);

  ws.onopen = () => {
    connected.set(true);
    loadHistory();
    loadConfig();
  };
  ws.onclose = () => {
    connected.set(false);
    reconnectTimer = setTimeout(connectWS, 2000);
  };
  ws.onerror = () => { ws?.close(); };
  ws.onmessage = (e) => { handleEvent(JSON.parse(e.data)); };
}

function handleEvent(event: { type: string; data: any }) {
  switch (event.type) {
    case 'call_start': {
      const d = event.data;
      calls.update((m) => {
        m.set(d.id, {
          id: d.id, provider: d.provider, model: d.model,
          endpoint: d.endpoint, streaming: d.streaming,
          status_code: 0, started_at: d.started_at, duration_ms: 0,
          input_tokens: 0, output_tokens: 0, cost: 0,
          streamText: '', completed: false,
        });
        return new Map(m);
      });
      break;
    }
    case 'call_chunk': {
      const d = event.data;
      calls.update((m) => {
        const call = m.get(d.id);
        if (call) {
          call.streamText += d.delta;
          m.set(d.id, { ...call });
        }
        return new Map(m);
      });
      break;
    }
    case 'call_end': {
      const d = event.data;
      calls.update((m) => {
        const call = m.get(d.id);
        if (call) {
          call.status_code = d.status_code;
          call.duration_ms = d.duration_ms;
          call.input_tokens = d.input_tokens;
          call.output_tokens = d.output_tokens;
          call.cost = d.cost;
          call.error = d.error;
          call.completed = true;
          m.set(d.id, { ...call });
        }
        return new Map(m);
      });
      // Invalidate cached detail
      callDetails.update(m => { m.delete(d.id); return new Map(m); });
      break;
    }
    case 'session_update': {
      session.set(event.data);
      break;
    }
    case 'budget_exceeded': {
      budgetExceeded.set(true);
      break;
    }
    case 'insight': {
      // Real-time insight from backend
      const insight = event.data as Insight;
      insights.update(list => [insight, ...list]);
      break;
    }
  }
}

async function loadHistory() {
  try {
    const [sessResp, callsResp] = await Promise.all([
      fetch('/api/session'),
      fetch('/api/calls'),
    ]);
    if (sessResp.ok) {
      const s = await sessResp.json();
      session.set({ total_cost: s.total_cost, total_tokens: s.total_tokens, request_count: s.request_count });
    }
    if (callsResp.ok) {
      const list = await callsResp.json();
      if (list && list.length > 0) {
        calls.update((m) => {
          for (const c of list) {
            if (!m.has(c.id)) {
              m.set(c.id, { ...c, started_at: new Date(c.started_at).getTime(), streamText: '', completed: true });
            }
          }
          return new Map(m);
        });
      }
    }
  } catch {
    // Will get data via WebSocket
  }
}

export function disconnectWS() {
  clearTimeout(reconnectTimer);
  ws?.close();
}

// Fetch full detail for a call (cached)
export async function fetchDetail(id: string): Promise<APICallDetail | null> {
  const cached = get(callDetails).get(id);
  if (cached) return cached;
  try {
    const resp = await fetch(`/api/calls/${id}`);
    if (!resp.ok) return null;
    const detail: APICallDetail = await resp.json();
    callDetails.update(m => { m.set(id, detail); return new Map(m); });
    return detail;
  } catch {
    return null;
  }
}

// Export filtered calls as JSON
export function exportJSON() {
  const data = get(filteredCalls);
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
  downloadBlob(blob, `llmview-${Date.now()}.json`);
}

// Export filtered calls as CSV
export function exportCSV() {
  const data = get(filteredCalls);
  const headers = ['id', 'provider', 'model', 'endpoint', 'status_code', 'started_at', 'duration_ms', 'input_tokens', 'output_tokens', 'cost', 'streaming', 'error'];
  const rows = data.map(c =>
    headers.map(h => {
      const val = (c as any)[h];
      if (h === 'started_at') return new Date(val).toISOString();
      if (typeof val === 'string' && (val.includes(',') || val.includes('"') || val.includes('\n')))
        return `"${val.replace(/"/g, '""')}"`;
      return val ?? '';
    }).join(',')
  );
  const csv = [headers.join(','), ...rows].join('\n');
  downloadBlob(new Blob([csv], { type: 'text/csv' }), `llmview-${Date.now()}.csv`);
}

async function loadConfig() {
  try {
    const resp = await fetch('/api/config');
    if (resp.ok) {
      const cfg = await resp.json();
      appConfig.set({ budget: cfg.budget || 0 });
    }
  } catch {}
}

// Insights
export interface Insight {
  type: 'loop_detected' | 'prompt_waste' | 'model_downgrade' | 'burn_rate';
  severity: 'info' | 'warning' | 'critical';
  title: string;
  description: string;
  savings?: number;
  token_savings?: number;
  call_ids?: string[];
  detected_at: number;
}

export const insights = writable<Insight[]>([]);

export async function fetchInsights(): Promise<Insight[]> {
  try {
    const resp = await fetch('/api/insights');
    if (!resp.ok) return [];
    const data: Insight[] = await resp.json();
    insights.set(data);
    return data;
  } catch {
    return [];
  }
}

// Replay a stored call with optional overrides
export async function replayCall(callId: string, overrides: Record<string, any> = {}): Promise<{ status_code?: number; error?: string }> {
  try {
    const resp = await fetch('/api/replay', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ call_id: callId, overrides }),
    });
    return await resp.json();
  } catch (e: any) {
    return { error: e.message || 'Replay failed' };
  }
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
