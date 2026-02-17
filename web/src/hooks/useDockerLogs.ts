import { useEffect, useRef, useState } from 'react';
import { getLogsWebSocketUrl } from '../services/api';

const MAX_LENGTH = 100_000;

export function useDockerLogs(containerId: string | null, enableWhen = true) {
  const [text, setText] = useState('');
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const mountedRef = useRef(true);
  const closedByCleanupRef = useRef(false);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (!containerId || !enableWhen) {
      queueMicrotask(() => {
        setText('');
        setConnected(false);
        setError(null);
      });
      return;
    }

    queueMicrotask(() => {
      setError(null);
      setText('');
    });
    closedByCleanupRef.current = false;

    const timeoutId = setTimeout(() => {
      if (closedByCleanupRef.current) return;

      const url = getLogsWebSocketUrl(containerId);
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        if (mountedRef.current) setConnected(true);
      };

      ws.onmessage = (event) => {
        if (!mountedRef.current) return;
        const chunk = typeof event.data === 'string' ? event.data : '';
        if (chunk.startsWith('{"error":')) {
          try {
            const o = JSON.parse(chunk) as { error?: string };
            if (o.error && mountedRef.current) setError(o.error);
          } catch {
            setText((prev) => prev + chunk);
          }
          return;
        }
        setText((prev) => {
          const next = prev + chunk;
          if (next.length > MAX_LENGTH) return next.slice(-MAX_LENGTH);
          return next;
        });
      };

      ws.onerror = () => {
        if (mountedRef.current && !closedByCleanupRef.current) {
          setError('Erro na conexão WebSocket (logs)');
        }
      };

      ws.onclose = () => {
        wsRef.current = null;
        if (mountedRef.current && !closedByCleanupRef.current) {
          setConnected(false);
        }
      };
    }, 300);

    return () => {
      closedByCleanupRef.current = true;
      clearTimeout(timeoutId);
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [containerId, enableWhen]);

  return { text, connected, error };
}
