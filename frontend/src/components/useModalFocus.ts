import { useEffect, useRef } from "react";

// Trap keyboard focus, support Escape, lock scrolling and restore the opener.
export function useModalFocus<T extends HTMLElement>(
  active: boolean,
  close: () => void,
) {
  const ref = useRef<T>(null);
  const closeRef = useRef(close);
  useEffect(() => {
    closeRef.current = close;
  }, [close]);
  useEffect(() => {
    if (!active || !ref.current) return;
    const panel = ref.current;
    const opener = document.activeElement as HTMLElement | null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const focusable = () =>
      Array.from(
        panel.querySelectorAll<HTMLElement>(
          'button:not([disabled]), a[href], input:not([disabled]), select, textarea, [tabindex="0"]',
        ),
      ).filter((el) => el.getClientRects().length > 0);
    focusable()[0]?.focus();
    const keydown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        closeRef.current();
        return;
      }
      if (event.key !== "Tab") return;
      const nodes = focusable(),
        first = nodes[0],
        last = nodes[nodes.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last?.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first?.focus();
      }
    };
    const keepFocus = (event: FocusEvent) => {
      if (event.target instanceof Node && !panel.contains(event.target))
        focusable()[0]?.focus();
    };
    const frame = requestAnimationFrame(() => focusable()[0]?.focus());
    document.addEventListener("keydown", keydown);
    document.addEventListener("focusin", keepFocus);
    return () => {
      cancelAnimationFrame(frame);
      document.removeEventListener("keydown", keydown);
      document.removeEventListener("focusin", keepFocus);
      document.body.style.overflow = previousOverflow;
      opener?.focus();
    };
  }, [active]);
  return ref;
}
