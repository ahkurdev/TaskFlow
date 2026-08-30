import { useEffect, useState, type FormEvent } from "react";
import { Link, useParams } from "react-router-dom";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  closestCorners,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { SortableContext, arrayMove, horizontalListSortingStrategy } from "@dnd-kit/sortable";
import * as boardsApi from "../api/boards";
import * as listsApi from "../api/lists";
import * as cardsApi from "../api/cards";
import { KanbanCard } from "../components/KanbanCard";
import { KanbanList } from "../components/KanbanList";
import { useBoardSocket } from "../hooks/useBoardSocket";
import type { BoardDetail, Card, ListWithCards } from "../types";

interface DragData {
  type: "column" | "card" | "list";
  listId: number;
  cardId?: number;
}

export function BoardDetailPage() {
  const { boardId } = useParams<{ boardId: string }>();
  const id = Number(boardId);

  const [board, setBoard] = useState<BoardDetail | null>(null);
  const [lists, setLists] = useState<ListWithCards[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isAddingList, setIsAddingList] = useState(false);
  const [newListName, setNewListName] = useState("");

  const [activeCard, setActiveCard] = useState<Card | null>(null);
  const [activeColumn, setActiveColumn] = useState<ListWithCards | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
  );

  useEffect(() => {
    void refetch({ silent: false });
  }, [id]);

  useBoardSocket(id, () => void refetch({ silent: true }));

  async function refetch({ silent }: { silent: boolean }) {
    if (!silent) setIsLoading(true);
    try {
      const detail = await boardsApi.getBoard(id);
      setBoard(detail);
      setLists(detail.lists);
      setError(null);
    } catch {
      setError("Failed to load board");
    } finally {
      if (!silent) setIsLoading(false);
    }
  }

  async function handleAddList(e: FormEvent) {
    e.preventDefault();
    if (!newListName.trim()) return;
    try {
      const list = await listsApi.createList(id, newListName.trim());
      setLists((prev) => [...prev, { ...list, cards: [] }]);
      setNewListName("");
      setIsAddingList(false);
    } catch {
      setError("Failed to create list");
    }
  }

  async function handleAddCard(listId: number, title: string) {
    try {
      const card = await cardsApi.createCard(id, listId, title, "");
      setLists((prev) =>
        prev.map((l) =>
          l.id === listId ? { ...l, cards: [...l.cards, card] } : l,
        ),
      );
    } catch {
      setError("Failed to create card");
    }
  }

  async function handleDeleteCard(listId: number, cardId: number) {
    setLists((prev) =>
      prev.map((l) =>
        l.id === listId
          ? { ...l, cards: l.cards.filter((c) => c.id !== cardId) }
          : l,
      ),
    );
    try {
      await cardsApi.deleteCard(id, listId, cardId);
    } catch {
      setError("Failed to delete card");
      void refetch({ silent: true });
    }
  }

  async function handleDeleteList(listId: number) {
    setLists((prev) => prev.filter((l) => l.id !== listId));
    try {
      await listsApi.deleteList(id, listId);
    } catch {
      setError("Failed to delete list");
      void refetch({ silent: true });
    }
  }

  function handleDragStart(event: DragStartEvent) {
    const data = event.active.data.current as DragData | undefined;
    if (!data) return;

    if (data.type === "card") {
      const list = lists.find((l) => l.id === data.listId);
      setActiveCard(list?.cards.find((c) => c.id === data.cardId) ?? null);
    } else if (data.type === "column") {
      setActiveColumn(lists.find((l) => l.id === data.listId) ?? null);
    }
  }

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    setActiveCard(null);
    setActiveColumn(null);
    if (!over) return;

    const activeData = active.data.current as DragData | undefined;
    const overData = over.data.current as DragData | undefined;
    if (!activeData) return;

    if (activeData.type === "column") {
      if (active.id === over.id) return;
      const oldIndex = lists.findIndex((l) => `column-${l.id}` === active.id);
      const newIndex = lists.findIndex((l) => `column-${l.id}` === over.id);
      if (oldIndex === -1 || newIndex === -1) return;

      const movedListId = lists[oldIndex].id;
      setLists((prev) => arrayMove(prev, oldIndex, newIndex));

      listsApi.reorderList(id, movedListId, newIndex).catch(() => {
        setError("Failed to reorder list");
        void refetch({ silent: true });
      });
      return;
    }

    if (activeData.type === "card" && overData) {
      const cardId = activeData.cardId!;
      const sourceListId = activeData.listId;
      const targetListId = overData.listId;

      let targetIndex = 0;
      const targetList = lists.find((l) => l.id === targetListId);
      if (!targetList) return;

      if (overData.type === "card") {
        const idx = targetList.cards.findIndex((c) => c.id === overData.cardId);
        targetIndex = idx === -1 ? targetList.cards.length : idx;
      } else {
        targetIndex = targetList.cards.length;
      }

      if (sourceListId === targetListId) {
        const sourceIndex = targetList.cards.findIndex((c) => c.id === cardId);
        if (sourceIndex === targetIndex) return;
      }

      setLists((prev) => {
        const source = prev.find((l) => l.id === sourceListId);
        const movingCard = source?.cards.find((c) => c.id === cardId);
        if (!movingCard) return prev;

        return prev.map((l) => {
          if (l.id === sourceListId && l.id === targetListId) {
            const withoutMoved = l.cards.filter((c) => c.id !== cardId);
            const clampedIndex = Math.min(targetIndex, withoutMoved.length);
            withoutMoved.splice(clampedIndex, 0, movingCard);
            return { ...l, cards: withoutMoved };
          }
          if (l.id === sourceListId) {
            return { ...l, cards: l.cards.filter((c) => c.id !== cardId) };
          }
          if (l.id === targetListId) {
            const cards = [...l.cards];
            const clampedIndex = Math.min(targetIndex, cards.length);
            cards.splice(clampedIndex, 0, movingCard);
            return { ...l, cards };
          }
          return l;
        });
      });

      cardsApi
        .moveCard(id, sourceListId, cardId, targetListId, targetIndex)
        .catch(() => {
          setError("Failed to move card");
          void refetch({ silent: true });
        });
    }
  }

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center text-slate-400">
        Loading board...
      </div>
    );
  }

  if (!board) {
    return (
      <div className="flex h-screen flex-col items-center justify-center gap-2 text-slate-400">
        <p>Board not found.</p>
        <Link to="/boards" className="text-indigo-400 hover:underline">
          Back to boards
        </Link>
      </div>
    );
  }

  const columnIds = lists.map((l) => `column-${l.id}`);

  return (
    <div className="flex h-screen flex-col">
      <header className="flex items-center justify-between border-b border-slate-800 px-6 py-4">
        <div className="flex items-center gap-3">
          <Link to="/boards" className="text-slate-400 hover:text-white">
            &larr;
          </Link>
          <h1 className="text-lg font-semibold text-white">{board.name}</h1>
        </div>
        {error && <span className="text-sm text-red-400">{error}</span>}
      </header>

      <main className="flex-1 overflow-x-auto px-6 py-6">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCorners}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
        >
          <div className="flex h-full items-start gap-4">
            <SortableContext items={columnIds} strategy={horizontalListSortingStrategy}>
              {lists.map((list) => (
                <KanbanList
                  key={list.id}
                  list={list}
                  onAddCard={handleAddCard}
                  onDeleteCard={handleDeleteCard}
                  onDeleteList={handleDeleteList}
                />
              ))}
            </SortableContext>

            <div className="w-72 shrink-0">
              {isAddingList ? (
                <form
                  onSubmit={handleAddList}
                  className="space-y-2 rounded-xl border border-slate-800 bg-slate-900 p-3"
                >
                  <input
                    autoFocus
                    value={newListName}
                    onChange={(e) => setNewListName(e.target.value)}
                    placeholder="List name"
                    className="w-full rounded-lg border border-slate-700 bg-slate-800 px-2 py-1 text-sm text-white outline-none focus:border-indigo-500"
                  />
                  <div className="flex gap-2">
                    <button
                      type="submit"
                      className="rounded-lg bg-indigo-600 px-3 py-1 text-xs font-medium text-white hover:bg-indigo-500"
                    >
                      Add list
                    </button>
                    <button
                      type="button"
                      onClick={() => setIsAddingList(false)}
                      className="px-3 py-1 text-xs text-slate-400 hover:text-white"
                    >
                      Cancel
                    </button>
                  </div>
                </form>
              ) : (
                <button
                  onClick={() => setIsAddingList(true)}
                  className="w-full rounded-xl border border-dashed border-slate-700 px-3 py-2 text-left text-sm text-slate-400 hover:border-slate-600 hover:text-white"
                >
                  + Add list
                </button>
              )}
            </div>
          </div>

          <DragOverlay>
            {activeCard && (
              <KanbanCard card={activeCard} listId={activeCard.list_id} onDelete={() => {}} />
            )}
            {activeColumn && (
              <div className="w-72 rounded-xl border border-indigo-600 bg-slate-900 p-3 text-sm font-semibold text-white shadow-lg">
                {activeColumn.name}
              </div>
            )}
          </DragOverlay>
        </DndContext>
      </main>
    </div>
  );
}
