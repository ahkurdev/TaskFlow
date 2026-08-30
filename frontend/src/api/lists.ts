import { apiRequest } from "./client";
import type { List } from "../types";

export function createList(boardId: number, name: string): Promise<List> {
  return apiRequest<List>(`/boards/${boardId}/lists`, {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export function updateList(
  boardId: number,
  listId: number,
  name: string,
): Promise<List> {
  return apiRequest<List>(`/boards/${boardId}/lists/${listId}`, {
    method: "PUT",
    body: JSON.stringify({ name }),
  });
}

export function reorderList(
  boardId: number,
  listId: number,
  position: number,
): Promise<void> {
  return apiRequest<void>(`/boards/${boardId}/lists/${listId}/reorder`, {
    method: "PATCH",
    body: JSON.stringify({ position }),
  });
}

export function deleteList(boardId: number, listId: number): Promise<void> {
  return apiRequest<void>(`/boards/${boardId}/lists/${listId}`, {
    method: "DELETE",
  });
}
