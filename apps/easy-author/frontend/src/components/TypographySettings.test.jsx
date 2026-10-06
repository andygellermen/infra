import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { DEFAULT_TYPOGRAPHY } from "../lib/typography";
import TypographySettings from "./TypographySettings";

describe("TypographySettings", () => {
  it("edits the global writing standard", () => {
    const onGlobalChange = vi.fn();
    render(<TypographySettings globalDefaults={DEFAULT_TYPOGRAPHY} bookOverrides={{}} onGlobalChange={onGlobalChange} />);
    fireEvent.change(screen.getByRole("spinbutton", { name: "Fließtext global" }), { target: { value: "19" } });
    expect(onGlobalChange).toHaveBeenCalledWith(expect.objectContaining({ bodySize: 19 }));
  });

  it("shows inherited values and identifies book overrides", () => {
    render(<TypographySettings globalDefaults={DEFAULT_TYPOGRAPHY} bookOverrides={{ h1Size: 46 }} />);
    expect(screen.getByText("H1 · abweichend")).toBeVisible();
    expect(screen.getByText(`Fließtext · geerbt (${DEFAULT_TYPOGRAPHY.bodySize})`)).toBeVisible();
  });

  it("resets the book to empty overrides", () => {
    const onResetBook = vi.fn();
    render(<TypographySettings globalDefaults={DEFAULT_TYPOGRAPHY} bookOverrides={{ bodySize: 20 }} onResetBook={onResetBook} />);
    fireEvent.click(screen.getByRole("button", { name: "Auf Gesamtstandard zurücksetzen" }));
    expect(onResetBook).toHaveBeenCalledWith({});
  });

  it("returns an overridden book field to inheritance when cleared", () => {
    const onBookChange = vi.fn();
    render(
      <TypographySettings
        globalDefaults={DEFAULT_TYPOGRAPHY}
        bookOverrides={{ firstLineIndent: 24 }}
        onBookChange={onBookChange}
      />,
    );
    fireEvent.change(screen.getByRole("spinbutton", { name: "Erstzeileneinzug für dieses Buch" }), {
      target: { value: "" },
    });
    expect(onBookChange).toHaveBeenCalledWith({});
  });
});
