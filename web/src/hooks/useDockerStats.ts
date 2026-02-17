import { useEffect, useRef, useState } from 'react';
import type { ContainerMetrics } from '../types/docker';
import { getStatsWebSocketUrl } from '../services/api';

const MAX_POINTS = 60;

export function useDockerStats(containerId: string | null) {
  const [metrics, setMetrics] = useState<ContainerMetrics[]>([]);
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
    if (!containerId) {
      queueMicrotask(() => {
        setMetrics([]);
        setConnected(false);
        setError(null);
      });
      return;
    }

    queueMicrotask(() => {
      setError(null);
      setMetrics([]);
    });
    closedByCleanupRef.current = false;

    const timeoutId = setTimeout(() => {
      if (closedByCleanupRef.current) return;

      const url = getStatsWebSocketUrl(containerId);
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        if (mountedRef.current) setConnected(true);
      };

      ws.onmessage = (event) => {
        if (!mountedRef.current) return;
        try {
          const data = JSON.parse(event.data) as ContainerMetrics;
          setMetrics((prev) => {
            const next = [...prev, data];
            if (next.length > MAX_POINTS) return next.slice(-MAX_POINTS);
            return next;
          });
        } catch {
          void 0;
        }
      };

      ws.onerror = () => {
        if (mountedRef.current && !closedByCleanupRef.current) {
          setError('Erro na conexão WebSocket');
        }
      };

      ws.onclose = () => {
        wsRef.current = null;
        if (mountedRef.current && !closedByCleanupRef.current) {
          setConnected(false);
        }
      };
    }, 0);

    return () => {
      closedByCleanupRef.current = true;
      clearTimeout(timeoutId);
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [containerId]);

  return { metrics, connected, error };
}
