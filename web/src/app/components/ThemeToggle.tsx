import { useEffect, useRef, useState } from "react";
import { Check, ChevronDown, Monitor, Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { Button } from "./ui/button";

const themeOptions = [
  { value: "light", label: "浅色", icon: Sun },
  { value: "dark", label: "深色", icon: Moon },
  { value: "system", label: "跟随系统", icon: Monitor },
] as const;

type ThemeValue = (typeof themeOptions)[number]["value"];

export function ThemeToggle() {
  const [mounted, setMounted] = useState(false);
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const { theme, resolvedTheme, setTheme } = useTheme();
  const activeTheme = (theme ?? "system") as ThemeValue;
  const activeOption = themeOptions.find((option) => option.value === activeTheme) ?? themeOptions[2];
  const isDark = resolvedTheme === "dark";

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!open) {
      return;
    }

    const handlePointerDown = (event: PointerEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setOpen(false);
      }
    };

    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  const selectTheme = (value: ThemeValue) => {
    setTheme(value);
    setOpen(false);
  };

  return (
    <div ref={menuRef} className="relative">
      <Button
        type="button"
        variant="outline"
        size="sm"
        aria-label="选择主题"
        aria-haspopup="menu"
        aria-expanded={open}
        title="选择浅色、深色或跟随系统"
        onClick={() => setOpen((current) => !current)}
      >
        {mounted && isDark ? <Moon data-icon="inline-start" /> : <Sun data-icon="inline-start" />}
        <span className="hidden lg:inline">{activeOption.label}</span>
        <ChevronDown data-icon="inline-end" className="opacity-70" />
      </Button>

      {open && (
        <div
          role="menu"
          aria-label="主题"
          className="absolute right-0 top-full z-[100] mt-2 w-44 rounded-md border bg-popover p-1 text-popover-foreground shadow-lg"
        >
          {themeOptions.map((option) => {
            const Icon = option.icon;
            const isActive = activeTheme === option.value;

            return (
              <button
                key={option.value}
                type="button"
                role="menuitemradio"
                aria-checked={isActive}
                className="flex w-full cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm outline-none transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
                onClick={() => selectTheme(option.value)}
              >
                <Icon />
                <span>{option.label}</span>
                <Check className={isActive ? "ml-auto opacity-100" : "ml-auto opacity-0"} />
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
