import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type { Card } from "../types";

interface Props {
  card: Card;
  listId: number;
  onDelete: () => void;
}

export function KanbanCard({ card, listId, onDelete }: Props) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: `card-${card.id}`,
    data: { type: "card", listId, cardId: card.id },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className="group cursor-grab rounded-lg border border-slate-700 bg-slate-800 p-3 text-sm text-slate-100 shadow-sm active:cursor-grabbing"
    >
      <div className="flex items-start justify-between gap-2">
        <p className="whitespace-pre-wrap break-words">{card.title}</p>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          className="text-slate-500 opacity-0 transition group-hover:opacity-100 hover:text-red-400"
          aria-label="Delete card"
        >
          x
        </button>
      </div>
      {card.description && (
        <p className="mt-1 text-xs text-slate-400">{card.description}</p>
      )}
    </div>
  );
}
