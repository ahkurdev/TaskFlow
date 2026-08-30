import { useEffect, useRef } from "react";
import { getToken } from "../api/client";

interface BoardEvent {
  type: string;
  board_id: number;
}

/**
 * Subscribes to live updates for a board over WebSocket. On any
 * "board_updated" event for this board, calls onUpdate (typically a
 * refetch) so other tabs/devices editing the same board stay in sync.
 */
export function useBoardSocket(boardId: number, onUpdate: () => void) {
  const onUpdateRef = useRef(onUpdate);
  onUpdateRef.current = onUpdate;

  useEffect(() => {
    const token = getToken();
    if (!token || !boardId) return;

    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    const socket = new WebSocket(
      `${protocol}://${window.location.host}/ws/boards/${boardId}?token=${encodeURIComponent(token)}`,
    );

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as BoardEvent;
        if (data.type === "board_updated" && data.board_id === boardId) {
          onUpdateRef.current();
        }
      } catch {
        // ignore malformed messages
      }
    };

    return () => socket.close();
  }, [boardId]);
}
