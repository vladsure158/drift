import { useState, useCallback } from "react";
import { useInput } from "ink";

interface UseKeyboardOptions {
  listLength: number;
  onSelect?: (index: number) => void;
  onQuit?: () => void;
}

export function useKeyboard({ listLength, onSelect, onQuit }: UseKeyboardOptions) {
  const [selectedIndex, setSelectedIndex] = useState(0);

  useInput((input, key) => {
    if (input === "q" || (key.ctrl && input === "c")) {
      onQuit?.();
      return;
    }

    if (key.upArrow || input === "k") {
      setSelectedIndex((prev) => (prev > 0 ? prev - 1 : listLength - 1));
    }

    if (key.downArrow || input === "j") {
      setSelectedIndex((prev) => (prev < listLength - 1 ? prev + 1 : 0));
    }

    if (key.return) {
      onSelect?.(selectedIndex);
    }
  });

  return { selectedIndex, setSelectedIndex };
}
