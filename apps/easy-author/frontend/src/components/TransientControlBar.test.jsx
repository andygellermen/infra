import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";
import TransientControlBar from "./TransientControlBar";

describe("TransientControlBar", () => {
  it("reveals all controls from the ellipsis trigger", () => {
    render(<TransientControlBar book={{ title: "Roman" }} chapter={{ title: "Kapitel 1" }} workView="clean" saveState="Gespeichert" />);
    fireEvent.click(screen.getByRole("button", { name: "Alle Bedienelemente anzeigen" }));
    expect(screen.getByRole("toolbar", { name: "Schreibsteuerung" })).toBeVisible();
    expect(screen.getByText("Roman")).toBeVisible();
    expect(screen.getByText("Kapitel 1")).toBeVisible();
  });

  it("routes work view and global actions", () => {
    const onWorkView = vi.fn();
    const onAppearance = vi.fn();
    const onSettings = vi.fn();
    const onCommand = vi.fn();
    render(<TransientControlBar workView="clean" onWorkView={onWorkView} onAppearance={onAppearance} onSettings={onSettings} onCommand={onCommand} />);
    fireEvent.click(screen.getByRole("button", { name: "Alle Bedienelemente anzeigen" }));
    fireEvent.click(screen.getByRole("button", { name: /Clean & Free/ }));
    fireEvent.click(screen.getByRole("button", { name: "Befehlspalette öffnen" }));
    fireEvent.click(screen.getByRole("button", { name: "Erscheinungsbild öffnen" }));
    fireEvent.click(screen.getByRole("button", { name: "Einstellungen öffnen" }));
    expect(onWorkView).toHaveBeenCalledWith("clean");
    expect(onCommand).toHaveBeenCalledOnce();
    expect(onAppearance).toHaveBeenCalledOnce();
    expect(onSettings).toHaveBeenCalledOnce();
  });

  it("keeps the established editor actions inside the temporary bar", () => {
    const onSave = vi.fn();
    const onEditorMode = vi.fn();
    const onWritingTools = vi.fn();
    const onHelp = vi.fn();
    const onFullscreen = vi.fn();
    render(
      <TransientControlBar
        editorMode="rich"
        onSave={onSave}
        onEditorMode={onEditorMode}
        onWritingTools={onWritingTools}
        onHelp={onHelp}
        onFullscreen={onFullscreen}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Alle Bedienelemente anzeigen" }));
    fireEvent.click(screen.getByRole("button", { name: "Kapitel speichern" }));
    fireEvent.click(screen.getByRole("button", { name: "Markdown" }));
    fireEvent.click(screen.getByRole("button", { name: "Werkzeuge" }));
    fireEvent.click(screen.getByRole("button", { name: "Hilfe" }));
    fireEvent.click(screen.getByRole("button", { name: "Vollbild" }));
    expect(onSave).toHaveBeenCalledOnce();
    expect(onEditorMode).toHaveBeenCalledWith("markdown");
    expect(onWritingTools).toHaveBeenCalledOnce();
    expect(onHelp).toHaveBeenCalledOnce();
    expect(onFullscreen).toHaveBeenCalledOnce();
  });

  it("closes the temporary bar with Escape", () => {
    render(<TransientControlBar />);
    fireEvent.click(screen.getByRole("button", { name: "Alle Bedienelemente anzeigen" }));
    fireEvent.keyDown(screen.getByRole("toolbar", { name: "Schreibsteuerung" }), { key: "Escape" });
    expect(screen.queryByRole("toolbar", { name: "Schreibsteuerung" })).not.toBeInTheDocument();
  });

  it("supports complete forward keyboard traversal through its actions", async () => {
    const user = userEvent.setup();
    render(<TransientControlBar />);
    await user.tab();
    expect(screen.getByRole("button", { name: "Alle Bedienelemente anzeigen" })).toHaveFocus();
    await user.keyboard("{Enter}");
    await user.tab();
    expect(screen.getByRole("button", { name: /Clean & Free/ })).toHaveFocus();
    const actions = ["Kapitel speichern", "Rich", "Markdown", "Werkzeuge", "Kanban öffnen", "Hilfe", "Vollbild", "Befehlspalette öffnen", "Erscheinungsbild öffnen", "Einstellungen öffnen"];
    for (const name of actions) {
      await user.tab();
      expect(screen.getByRole("button", { name })).toHaveFocus();
    }
  });
});
