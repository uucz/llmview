import { writable, derived } from 'svelte/store';

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

export interface SessionStats {
  total_cost: number;
  total_tokens: number;
  request_count: number;
}

export const calls = writable<Map<string, APICall>>(new Map());
export const session = writable<SessionStats>({ total_cost: 0, total_tokens: 0, request_count: 0 });
export const connected = writable(false);

// Derived: calls sorted by time, newest first
export const sortedCalls = derived(calls, ($calls) => {
  return Array.from($calls.values()).sort((a, b) => b.started_at - a.started_at);
});

let ws: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout>;

export function connectWS() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const url = `${protocol}//${window.location.host}/ws`;

  ws = new WebSocket(url);

  ws.onopen = () => {
    connected.set(true);
    // Load existing calls from REST API
    loadHistory();
  };

  ws.onclose = () => {
    connected.set(false);
    reconnectTimer = setTimeout(connectWS, 2000);
  };

  ws.onerror = () => {
    ws?.close();
  };

  ws.onmessage = (e) => {
    const event = JSON.parse(e.data);
    handleEvent(event);
  };
}

function handleEvent(event: { type: string; data: any }) {
  switch (event.type) {
    case 'call_start': {
      const d = event.data;
      calls.update((m) => {
        m.set(d.id, {
          id: d.id,
          provider: d.provider,
          model: d.model,
          endpoint: d.endpoint,
          streaming: d.streaming,
          status_code: 0,
          started_at: d.started_at,
          duration_ms: 0,
          input_tokens: 0,
          output_tokens: 0,
          cost: 0,
          streamText: '',
          completed: false,
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
      break;
    }

    case 'session_update': {
      session.set(event.data);
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
      session.set({
        total_cost: s.total_cost,
        total_tokens: s.total_tokens,
        request_count: s.request_count,
      });
    }
    if (callsResp.ok) {
      const list = await callsResp.json();
      if (list && list.length > 0) {
        calls.update((m) => {
          for (const c of list) {
            if (!m.has(c.id)) {
              m.set(c.id, {
                ...c,
                started_at: new Date(c.started_at).getTime(),
                streamText: '',
                completed: true,
              });
            }
          }
          return new Map(m);
        });
      }
    }
  } catch {
    // Silently fail — will get data via WebSocket
  }
}

export function disconnectWS() {
  clearTimeout(reconnectTimer);
  ws?.close();
}
