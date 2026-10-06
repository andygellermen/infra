import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import KanbanBoard, { KANBAN_PHASES } from "./KanbanBoard";

const baseItems = {
  backlog: [{ id: "one", title: "Quelle pruefen", book_id: "a", chapter_id: "c", priority: "high" }],
  todo: [], in_progress: [], review: [], done: [],
};

describe("KanbanBoard", () => {
  it("renders all five phases, totals, paging, source navigation, and multi-book ribbons", () => {
    const onLoadMore = vi.fn();
    const onOpenSource = vi.fn();
    render(<KanbanBoard phases={KANBAN_PHASES} items={baseItems} totals={{ backlog: 14 }} limit={12}
      selectedBookCount={2} bookTitles={{ a: "Nordlicht" }} colors={new Map([["a", "#006D77"]])}
      onLoadMore={onLoadMore} onOpenSource={onOpenSource} />);
    expect(screen.getAllByRole("region")).toHaveLength(5);
    expect(screen.getByRole("heading", { name: /Backlog.*14/ })).toBeInTheDocument();
    expect(screen.getByText("Nordlicht")).toHaveClass("kanban-card__book-ribbon");
    fireEvent.click(screen.getByRole("button", { name: "Weitere Backlog-Karten anzeigen" }));
    expect(onLoadMore).toHaveBeenCalledWith("backlog");
    fireEvent.click(screen.getByRole("button", { name: "Quelle von Quelle pruefen öffnen" }));
    expect(onOpenSource).toHaveBeenCalledWith(baseItems.backlog[0]);
  });

  it("hides book ribbons for one book and moves cards by pointer and keyboard", () => {
    const onMove = vi.fn();
    render(<KanbanBoard phases={KANBAN_PHASES} items={baseItems} totals={{ backlog: 1 }} limit={12}
      selectedBookCount={1} bookTitles={{ a: "Nordlicht" }} colors={new Map([["a", "#006D77"]])} onMove={onMove} />);
    expect(screen.queryByText("Nordlicht")).not.toBeInTheDocument();
    const card = screen.getByRole("article", { name: "Quelle pruefen" });
    fireEvent.dragStart(card, { dataTransfer: { setData: vi.fn(), effectAllowed: "move" } });
    fireEvent.dragOver(screen.getByRole("region", { name: /In Arbeit/ }));
    fireEvent.drop(screen.getByRole("region", { name: /In Arbeit/ }), { dataTransfer: { getData: () => "one" } });
    expect(onMove).toHaveBeenCalledWith("one", "in_progress");
    fireEvent.keyDown(card, { key: "ArrowRight", altKey: true });
    expect(onMove).toHaveBeenCalledWith("one", "todo");
  });

  it("keeps five loaded columns operable and never renders more than twenty cards per phase", () => {
    const items = Object.fromEntries(KANBAN_PHASES.map(({ id }) => [id, Array.from({ length: 21 }, (_, index) => ({
      id: `${id}-${index}`, title: `${id} Aufgabe ${index + 1}`, book_id: "a",
    }))]));
    render(<KanbanBoard phases={KANBAN_PHASES} items={items}
      totals={Object.fromEntries(KANBAN_PHASES.map(({ id }) => [id, 21]))} limit={20} />);
    expect(screen.getAllByRole("article")).toHaveLength(100);
    expect(screen.queryByText("Weitere anzeigen")).not.toBeInTheDocument();
    expect(screen.getByRole("article", { name: "done Aufgabe 20" })).toHaveAttribute("tabindex", "0");
    expect(screen.queryByRole("article", { name: "done Aufgabe 21" })).not.toBeInTheDocument();
  });
});
