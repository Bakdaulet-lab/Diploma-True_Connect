import { useAuthStore } from '@/store/auth-store';

type MessageHandler = (data: Record<string, unknown>) => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private pingTimer: ReturnType<typeof setInterval> | null = null;
  private listeners: Map<string, Set<MessageHandler>> = new Map();
  private url: string;

  constructor() {
    const base = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080').replace(/^http/, 'ws');
    this.url = `${base}/v1/ws`;
  }

  connect() {
    const token = useAuthStore.getState().accessToken;
    if (!token || this.ws?.readyState === WebSocket.OPEN) return;

    this.ws = new WebSocket(`${this.url}?token=${token}`);

    this.ws.onopen = () => {
      this.startPing();
      this.emit('connected', {});
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.emit(data.type || 'message', data);
      } catch {
        // ignore non-JSON messages
      }
    };

    this.ws.onclose = () => {
      this.cleanup();
      this.scheduleReconnect();
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  disconnect() {
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.reconnectTimer = null;
    this.cleanup();
    this.ws?.close();
    this.ws = null;
  }

  sendMessage(matchId: string, content: string) {
    this.send({ type: 'message', match_id: matchId, content });
  }

  sendTyping(matchId: string) {
    this.send({ type: 'typing', match_id: matchId });
  }

  on(event: string, handler: MessageHandler) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(handler);
    return () => this.listeners.get(event)?.delete(handler);
  }

  private send(data: Record<string, unknown>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  private emit(event: string, data: Record<string, unknown>) {
    this.listeners.get(event)?.forEach((handler) => handler(data));
  }

  private startPing() {
    this.pingTimer = setInterval(() => {
      this.send({ type: 'ping' });
    }, 30000);
  }

  private cleanup() {
    if (this.pingTimer) clearInterval(this.pingTimer);
    this.pingTimer = null;
  }

  private scheduleReconnect() {
    this.reconnectTimer = setTimeout(() => this.connect(), 5000);
  }
}

export const wsService = new WebSocketService();
