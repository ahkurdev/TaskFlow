export interface User {
  id: number;
  name: string;
  email: string;
}

export interface Board {
  id: number;
  user_id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface List {
  id: number;
  board_id: number;
  name: string;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface Card {
  id: number;
  list_id: number;
  title: string;
  description: string;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface ListWithCards extends List {
  cards: Card[];
}

export interface BoardDetail extends Board {
  lists: ListWithCards[];
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ApiError {
  error: string;
}
