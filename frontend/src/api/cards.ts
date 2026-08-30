import { apiRequest } from "./client";
import type { Card } from "../types";

export function createCard(
  boardId: number,
  listId: number,
  title: string,
  description: string,
): Promise<Card> {
  return apiRequest<Card>(`/boards/${boardId}/lists/${listId}/cards`, {
    method: "POST",
    body: JSON.stringify({ title, description }),
  });
}

export function updateCard(
  boardId: number,
  listId: number,
  cardId: number,
  title: string,
  description: string,
): Promise<Card> {
  return apiRequest<Card>(`/boards/${boardId}/lists/${listId}/cards/${cardId}`, {
    method: "PUT",
    body: JSON.stringify({ title, description }),
  });
}

export function moveCard(
  boardId: number,
  listId: number,
  cardId: number,
  targetListId: number,
  position: number,
): Promise<Card> {
  return apiRequest<Card>(
    `/boards/${boardId}/lists/${listId}/cards/${cardId}/move`,
    {
      method: "PATCH",
      body: JSON.stringify({ target_list_id: targetListId, position }),
    },
  );
}

export function deleteCard(
  boardId: number,
  listId: number,
  cardId: number,
): Promise<void> {
  return apiRequest<void>(`/boards/${boardId}/lists/${listId}/cards/${cardId}`, {
    method: "DELETE",
  });
}
