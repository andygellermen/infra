import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ContextRail, { groupContextsByAnchor } from "./ContextRail";

const contexts = [
  { id: "comment-1", anchor_id: "anchor-1", context_type: "comment", status: "open", anchor: { id: "anchor-1", start_offset: 12 } },
  { id: "link-1", anchor_id: "anchor-1", context_type: "link", status: "open", anchor: { id: "anchor-1", start_offset: 12 } },
  { id: "task-1", anchor_id: "anchor-2", context_type: "work_item", status: "resolved", anchor: { id: "anchor-2", start_offset: 40 } },
];

describe("ContextRail", () => {
  it("groups several context types at one anchor without losing stable order", () => {
    const groups = groupContextsByAnchor(contexts);
    expect(groups).toHaveLength(2);
    expect(groups[0].anchorId).toBe("anchor-1");
    expect(groups[0].contexts.map((item) => item.context_type)).toEqual(["comment", "link"]);
  });

  it("exposes grouped icons, counts, state, and keyboard activation", () => {
    const onActivate = vi.fn();
    render(<ContextRail contexts={contexts} onActivate={onActivate} />);
    const grouped = screen.getByRole("button", { name: "2 Kontexte: Kommentar, Verknüpfung" });
    expect(grouped).toHaveAttribute("title", "Kommentar · Verknüpfung");
    fireEvent.keyDown(grouped, { key: "Enter" });
    expect(onActivate).toHaveBeenCalledWith(expect.objectContaining({ anchorId: "anchor-1" }));
    expect(screen.getByRole("button", { name: "1 Kontext: Aufgabe (erledigt)" })).toHaveClass("is-resolved");
  });
});
