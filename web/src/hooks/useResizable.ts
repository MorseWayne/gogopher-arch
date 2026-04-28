import { useCallback, useEffect, useRef, useState } from 'react';

interface UseResizableOptions {
  initialWidth: number;
  minWidth: number;
  maxWidth: number;
  storageKey?: string;
}

interface UseResizableReturn {
  width: number;
  startResize: (direction: 'left' | 'right') => (e: React.MouseEvent) => void;
}

export function useResizable({
  initialWidth,
  minWidth,
  maxWidth,
  storageKey,
}: UseResizableOptions): UseResizableReturn {
  const [width, setWidth] = useState(() => {
    if (storageKey) {
      try {
        const saved = localStorage.getItem(storageKey);
        if (saved) {
          const parsed = parseInt(saved, 10);
          if (!isNaN(parsed)) {
            return Math.max(minWidth, Math.min(maxWidth, parsed));
          }
        }
      } catch {
        // ignore localStorage errors
      }
    }
    return initialWidth;
  });

  const resizing = useRef<{
    startX: number;
    startWidth: number;
    direction: 'left' | 'right';
  } | null>(null);

  const startResize = useCallback(
    (direction: 'left' | 'right') => (e: React.MouseEvent) => {
      resizing.current = {
        startX: e.clientX,
        startWidth: width,
        direction,
      };
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
    },
    [width],
  );

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!resizing.current) return;
      const { startX, startWidth, direction } = resizing.current;
      const delta =
        direction === 'left' ? e.clientX - startX : startX - e.clientX;
      const newWidth = Math.max(
        minWidth,
        Math.min(maxWidth, startWidth + delta),
      );
      setWidth(newWidth);
    };

    const handleMouseUp = () => {
      if (resizing.current) {
        resizing.current = null;
        document.body.style.cursor = '';
        document.body.style.userSelect = '';
        if (storageKey) {
          try {
            localStorage.setItem(storageKey, String(width));
          } catch {
            // ignore localStorage errors
          }
        }
      }
    };

    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);
    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [minWidth, maxWidth, storageKey, width]);

  return { width, startResize };
}
