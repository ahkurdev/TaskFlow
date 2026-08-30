import { apiRequest } from "./client";
import type { Board, BoardDetail } from "../types";

export function listBoards(): Promise<Board[]> {
  return apiRequest<Board[]>("/boards/");
}

export function getBoard(boardId: number): Promise<BoardDetail> {
  return apiRequest<BoardDetail>(`/boards/${boardId}`);
}

export function createBoard(name: string, description: string): Promise<Board> {
  return apiRequest<Board>("/boards/", {
    method: "POST",
    body: JSON.stringify({ name, description }),
  });
}

export function updateBoard(
  boardId: number,
  name: string,
  description: string,
): Promise<Board> {
  return apiRequest<Board>(`/boards/${boardId}`, {
    method: "PUT",
    body: JSON.stringify({ name, description }),
  });
}

export function deleteBoard(boardId: number): Promise<void> {
  return apiRequest<void>(`/boards/${boardId}`, { method: "DELETE" });
}
