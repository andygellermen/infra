import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import WorkViewPicker from "./WorkViewPicker";

describe("WorkViewPicker", () => {
  it("offers the three approved work views as labeled cards", () => {
    render(<WorkViewPicker open currentView="clean" onSelect={() => {}} />);
    expect(screen.getByRole("dialog", { name: "Arbeitsansicht wählen" })).toBeVisible();
    expect(screen.getByRole("button", { name: /Clean & Free/ })).toBeVisible();
    expect(screen.getByRole("button", { name: /Intense/ })).toBeVisible();
    expect(screen.getByRole("button", { name: /Review/ })).toBeVisible();
  });

  it("marks a first-book choice for persistence", () => {
    const onSelect = vi.fn();
    render(<WorkViewPicker open currentView="clean" persistDefault onSelect={onSelect} />);
    fireEvent.click(screen.getByRole("button", { name: /Intense/ }));
    expect(onSelect).toHaveBeenCalledWith("intense", { persistDefault: true });
  });

  it("keeps a regular switch session-only", () => {
    const onSelect = vi.fn();
    render(<WorkViewPicker open currentView="clean" onSelect={onSelect} />);
    fireEvent.click(screen.getByRole("button", { name: /Review/ }));
    expect(onSelect).toHaveBeenCalledWith("review", { persistDefault: false });
  });
});
