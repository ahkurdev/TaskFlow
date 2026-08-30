import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import * as boardsApi from "../api/boards";
import { useAuth } from "../context/AuthContext";
import type { Board } from "../types";

export function BoardsPage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [boards, setBoards] = useState<Board[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    boardsApi
      .listBoards()
      .then(setBoards)
      .finally(() => setIsLoading(false));
  }, []);

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      const board = await boardsApi.createBoard(name, description);
      setBoards((prev) => [board, ...prev]);
      setName("");
      setDescription("");
      setIsCreating(false);
      navigate(`/boards/${board.id}`);
    } catch {
      setError("Failed to create board");
    }
  }

  return (
    <div className="min-h-screen">
      <header className="flex items-center justify-between border-b border-slate-800 px-6 py-4">
        <h1 className="text-xl font-semibold text-white">TaskFlow</h1>
        <div className="flex items-center gap-4 text-sm text-slate-400">
          <span>{user?.name}</span>
          <button onClick={logout} className="text-slate-400 hover:text-white">
            Sign out
          </button>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-lg font-medium text-white">Your boards</h2>
          <button
            onClick={() => setIsCreating(true)}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500"
          >
            + New board
          </button>
        </div>

        {isCreating && (
          <form
            onSubmit={handleCreate}
            className="mb-6 space-y-3 rounded-xl border border-slate-800 bg-slate-900 p-4"
          >
            <input
              autoFocus
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Board name"
              className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-white outline-none focus:border-indigo-500"
            />
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Description (optional)"
              className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-white outline-none focus:border-indigo-500"
            />
            {error && <p className="text-sm text-red-400">{error}</p>}
            <div className="flex gap-2">
              <button
                type="submit"
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500"
              >
                Create
              </button>
              <button
                type="button"
                onClick={() => setIsCreating(false)}
                className="rounded-lg px-4 py-2 text-sm text-slate-400 hover:text-white"
              >
                Cancel
              </button>
            </div>
          </form>
        )}

        {isLoading ? (
          <p className="text-slate-400">Loading boards...</p>
        ) : boards.length === 0 ? (
          <p className="text-slate-400">
            No boards yet. Create one to get started.
          </p>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
            {boards.map((board) => (
              <Link
                key={board.id}
                to={`/boards/${board.id}`}
                className="rounded-xl border border-slate-800 bg-slate-900 p-4 transition hover:border-indigo-600"
              >
                <h3 className="font-medium text-white">{board.name}</h3>
                {board.description && (
                  <p className="mt-1 line-clamp-2 text-sm text-slate-400">
                    {board.description}
                  </p>
                )}
              </Link>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
