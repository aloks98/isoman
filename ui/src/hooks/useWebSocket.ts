import { useEffect, useRef, useCallback } from 'react';
import type { WSProgressMessage } from '../types/iso';

/**
 * WebSocket URL - defaults to same origin in production
 * Can be overridden with PUBLIC_WS_URL environment variable
 */
const getWebSocketURL = () => {
  const wsUrl = import.meta.env.PUBLIC_WS_URL;
  if (wsUrl) return wsUrl;

  // In development, connect directly to backend on port 8080
  if (import.meta.env.DEV) {
    return 'ws://localhost:8080/ws';
  }

  // In production, use same origin
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}/ws`;
};

interface UseWebSocketOptions {
  onMessage?: (message: WSProgressMessage) => void;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

/**
 * Custom hook for managing WebSocket connection to backend
 * Handles automatic reconnection and message parsing
 */
export function useWebSocket(options: UseWebSocketOptions = {}) {
  const {
    onMessage,
    reconnectInterval = 3000,
    maxReconnectAttempts = 5,
  } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const isMountedRef = useRef(true);

  const connect = useCallback(() => {
    if (!isMountedRef.current) return;

    try {
      const wsUrl = getWebSocketURL();
      console.log('[WebSocket] Attempting to connect to:', wsUrl);
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log('[WebSocket] Connected successfully');
        reconnectAttemptsRef.current = 0;
      };

      ws.onmessage = (event) => {
        try {
          const message: WSProgressMessage = JSON.parse(event.data);
          onMessage?.(message);
        } catch (error) {
          console.error('[WebSocket] Failed to parse message:', error);
        }
      };

      ws.onerror = (error) => {
        console.error('[WebSocket] Error:', error);
      };

      ws.onclose = () => {
        console.log('[WebSocket] Disconnected');
        wsRef.current = null;

        // Attempt to reconnect if we haven't exceeded max attempts
        if (
          isMountedRef.current &&
          reconnectAttemptsRef.current < maxReconnectAttempts
        ) {
          reconnectAttemptsRef.current += 1;
          console.log(
            `[WebSocket] Reconnecting... (attempt ${reconnectAttemptsRef.current}/${maxReconnectAttempts})`
          );

          reconnectTimeoutRef.current = window.setTimeout(() => {
            connect();
          }, reconnectInterval);
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.error('[WebSocket] Failed to connect:', error);
    }
  }, [onMessage, reconnectInterval, maxReconnectAttempts]);

  useEffect(() => {
    isMountedRef.current = true;
    connect();

    return () => {
      isMountedRef.current = false;

      // Clear reconnect timeout
      if (reconnectTimeoutRef.current !== null) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }

      // Close WebSocket connection
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [connect]);

  return {
    ws: wsRef.current,
    isConnected: wsRef.current?.readyState === WebSocket.OPEN,
  };
}
