import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import BookKanbanFilter from "./BookKanbanFilter";

const books = [{ id: "a", title: "Nordlicht" }, { id: "b", title: "Suedwind" }];
const counts = {
  a: { backlog: 2, todo: 1, in_progress: 3, review: 1, done: 5 },
  b: { backlog: 0, todo: 4, in_progress: 0, review: 2, done: 1 },
};

describe("BookKanbanFilter", () => {
  it("shows three aggregated counters per book and supports multiple selection", () => {
    const onToggle = vi.fn();
    render(<BookKanbanFilter books={books} selectedIds={["a", "b"]} counts={counts} onToggle={onToggle} />);
    expect(screen.getByRole("button", { name: /Nordlicht aus Auswahl entfernen/ })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByLabelText("Nordlicht: offen")).toHaveTextContent("3");
    expect(screen.getByLabelText("Nordlicht: in Arbeit")).toHaveTextContent("4");
    expect(screen.getByLabelText("Nordlicht: fertig")).toHaveTextContent("5");
    fireEvent.click(screen.getByRole("button", { name: /Suedwind aus Auswahl entfernen/ }));
    expect(onToggle).toHaveBeenCalledWith("b");
  });
});
