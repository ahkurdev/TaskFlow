import { useState, type FormEvent } from "react";
import { useDroppable } from "@dnd-kit/core";
import {
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { KanbanCard } from "./KanbanCard";
import type { ListWithCards } from "../types";

interface Props {
  list: ListWithCards;
  onAddCard: (listId: number, title: string) => void;
  onDeleteCard: (listId: number, cardId: number) => void;
  onDeleteList: (listId: number) => void;
}

export function KanbanList({
  list,
  onAddCard,
  onDeleteCard,
  onDeleteList,
}: Props) {
  const [isAdding, setIsAdding] = useState(false);
  const [title, setTitle] = useState("");

  // Draggable "handle" for reordering this list among its siblings.
  const {
    attributes,
    listeners,
    setNodeRef: setColumnRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: `column-${list.id}`,
    data: { type: "column", listId: list.id },
  });

  // Separate droppable zone for the card area, so dropping into an empty
  // list still resolves to a valid drop target.
  const { setNodeRef: setDropZoneRef } = useDroppable({
    id: `list-${list.id}`,
    data: { type: "list", listId: list.id },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;
    onAddCard(list.id, title.trim());
    setTitle("");
    setIsAdding(false);
  }

  const cardIds = list.cards.map((c) => `card-${c.id}`);

  return (
    <div
      ref={setColumnRef}
      style={style}
      className="flex w-72 shrink-0 flex-col rounded-xl border border-slate-800 bg-slate-900"
    >
      <div
        {...attributes}
        {...listeners}
        className="flex cursor-grab items-center justify-between px-3 py-2 active:cursor-grabbing"
      >
        <h3 className="text-sm font-semibold text-white">{list.name}</h3>
        <button
          onClick={() => onDeleteList(list.id)}
          className="text-xs text-slate-500 hover:text-red-400"
        >
          Delete
        </button>
      </div>

      <div
        ref={setDropZoneRef}
        className="min-h-[40px] flex-1 space-y-2 px-2 pb-2"
      >
        <SortableContext items={cardIds} strategy={verticalListSortingStrategy}>
          {list.cards.map((card) => (
            <KanbanCard
              key={card.id}
              card={card}
              listId={list.id}
              onDelete={() => onDeleteCard(list.id, card.id)}
            />
          ))}
        </SortableContext>
      </div>

      <div className="p-2">
        {isAdding ? (
          <form onSubmit={handleSubmit} className="space-y-2">
            <input
              autoFocus
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Card title"
              className="w-full rounded-lg border border-slate-700 bg-slate-800 px-2 py-1 text-sm text-white outline-none focus:border-indigo-500"
            />
            <div className="flex gap-2">
              <button
                type="submit"
                className="rounded-lg bg-indigo-600 px-3 py-1 text-xs font-medium text-white hover:bg-indigo-500"
              >
                Add
              </button>
              <button
                type="button"
                onClick={() => setIsAdding(false)}
                className="px-3 py-1 text-xs text-slate-400 hover:text-white"
              >
                Cancel
              </button>
            </div>
          </form>
        ) : (
          <button
            onClick={() => setIsAdding(true)}
            className="w-full rounded-lg px-2 py-1 text-left text-sm text-slate-400 hover:bg-slate-800 hover:text-white"
          >
            + Add card
          </button>
        )}
      </div>
    </div>
  );
}
