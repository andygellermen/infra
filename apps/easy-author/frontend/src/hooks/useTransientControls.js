import { useCallback, useEffect, useRef, useState } from "react";

export function useTransientControls({ timeoutMs = 60000, blocked = false } = {}) {
  const [visible, setVisible] = useState(false);
  const [hovered, setHovered] = useState(false);
  const [focused, setFocused] = useState(false);
  const timerRef = useRef(null);

  const clearTimer = useCallback(() => {
    window.clearTimeout(timerRef.current);
    timerRef.current = null;
  }, []);

  const resetTimer = useCallback(() => {
    clearTimer();
    if (!visible || hovered || focused || blocked) return;
    timerRef.current = window.setTimeout(() => setVisible(false), timeoutMs);
  }, [blocked, clearTimer, focused, hovered, timeoutMs, visible]);

  const reveal = useCallback(() => setVisible(true), []);
  const dismiss = useCallback(() => {
    clearTimer();
    setVisible(false);
  }, [clearTimer]);
  const hide = useCallback(() => {
    clearTimer();
    if (!hovered && !focused && !blocked) setVisible(false);
  }, [blocked, clearTimer, focused, hovered]);

  useEffect(() => {
    resetTimer();
    return clearTimer;
  }, [clearTimer, resetTimer]);

  return {
    visible,
    reveal,
    dismiss,
    hide,
    resetTimer,
    onPointerEnter: () => { setHovered(true); clearTimer(); },
    onPointerLeave: () => setHovered(false),
    onFocusCapture: () => { setFocused(true); clearTimer(); },
    onBlurCapture: (event) => {
      if (!event?.currentTarget?.contains?.(event.relatedTarget)) setFocused(false);
    },
  };
}
