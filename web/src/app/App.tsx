import { ThemeProvider as NextThemesProvider } from "next-themes";
import { RouterProvider } from "react-router";
import { router } from "./routes";

export default function App() {
  return (
    <NextThemesProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
      <RouterProvider router={router} />
    </NextThemesProvider>
  );
}
