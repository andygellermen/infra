import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import CommentThread from "./CommentThread";

describe("CommentThread", () => {
  it("renders messages in API order and sends a reply", () => {
    const onReply = vi.fn();
    render(<CommentThread thread={{ id: "thread-1", status: "open", messages: [
      { id: "m1", author: "Andy", body: "Frage" },
      { id: "m2", author: "Cody", body: "Antwort" },
    ] }} onReply={onReply} />);
    expect(screen.getAllByRole("article").map((item) => item.textContent)).toEqual([
      expect.stringContaining("Frage"), expect.stringContaining("Antwort"),
    ]);
    fireEvent.change(screen.getByLabelText("Antwort"), { target: { value: "Noch ein Gedanke" } });
    fireEvent.click(screen.getByRole("button", { name: "Antwort senden" }));
    expect(onReply).toHaveBeenCalledWith("Noch ein Gedanke");
  });

  it("routes status changes and explicit deletion independently", () => {
    const onStatusChange = vi.fn();
    const onDelete = vi.fn();
    render(<CommentThread thread={{ id: "thread-1", status: "resolved", messages: [] }} onStatusChange={onStatusChange} onDelete={onDelete} />);
    expect(screen.getByText("Erledigt")).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "Wieder öffnen" }));
    fireEvent.click(screen.getByRole("button", { name: "Kommentar löschen" }));
    expect(onStatusChange).toHaveBeenCalledWith("open");
    expect(onDelete).toHaveBeenCalledOnce();
  });
});
