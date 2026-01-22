import { useEffect, useRef, useState } from 'react';
import { WebSocketMessage } from '../types';

interface UseWebSocketOptions {
  workspaceId: string;
  userId: string;
  onMessage?: (message: WebSocketMessage) => void;
}

export function useWebSocket({ workspaceId, userId, onMessage }: UseWebSocketOptions) {
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!workspaceId || !userId) return;

    const wsUrl = `ws://localhost:8080/ws?workspace_id=${workspaceId}&user_id=${userId}`;
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setIsConnected(true);
      console.log('WebSocket connected');
    };

    ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        onMessage?.(message);
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
      setIsConnected(false);
      console.log('WebSocket disconnected');
      // Attempt to reconnect after 3 seconds
      setTimeout(() => {
        if (workspaceId && userId) {
          wsRef.current = new WebSocket(wsUrl);
        }
      }, 3000);
    };

    wsRef.current = ws;

    return () => {
      ws.close();
    };
  }, [workspaceId, userId, onMessage]);

  return { isConnected };
}
