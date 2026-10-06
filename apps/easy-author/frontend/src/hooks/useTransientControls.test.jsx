import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useTransientControls } from "./useTransientControls";

afterEach(() => vi.useRealTimers());

describe("useTransientControls", () => {
  it("reveals on demand and hides after exactly sixty seconds", () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useTransientControls({ timeoutMs: 60000, blocked: false }));
    act(() => result.current.reveal());
    expect(result.current.visible).toBe(true);
    act(() => vi.advanceTimersByTime(59999));
    expect(result.current.visible).toBe(true);
    act(() => vi.advanceTimersByTime(1));
    expect(result.current.visible).toBe(false);
  });

  it("pauses while hovered or focused and while blocked", () => {
    vi.useFakeTimers();
    let blocked = false;
    const { result, rerender } = renderHook(() => useTransientControls({ timeoutMs: 60000, blocked }));
    act(() => { result.current.reveal(); result.current.onPointerEnter(); });
    act(() => vi.advanceTimersByTime(60000));
    expect(result.current.visible).toBe(true);
    act(() => { result.current.onPointerLeave(); result.current.onFocusCapture(); });
    act(() => vi.advanceTimersByTime(60000));
    expect(result.current.visible).toBe(true);
    blocked = true;
    rerender();
    act(() => result.current.onBlurCapture());
    act(() => vi.advanceTimersByTime(60000));
    expect(result.current.visible).toBe(true);
  });

  it("resets the deadline on interaction", () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useTransientControls({ timeoutMs: 60000, blocked: false }));
    act(() => result.current.reveal());
    act(() => vi.advanceTimersByTime(50000));
    act(() => result.current.resetTimer());
    act(() => vi.advanceTimersByTime(50000));
    expect(result.current.visible).toBe(true);
  });
});
