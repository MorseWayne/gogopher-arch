import { createBrowserRouter } from "react-router";
import { Landing } from "./pages/Landing";
import { Dashboard } from "./pages/Dashboard";
import { MissionDetail } from "./pages/MissionDetail";
import { GoBasicsCourse } from "./pages/GoBasicsCourse";
import { GoBasicsChapter } from "./pages/GoBasicsChapter";
import { Layout } from "./components/Layout";

export const router = createBrowserRouter([
  {
    path: "/",
    Component: Layout,
    children: [
      { index: true, Component: Landing },
      { path: "dashboard", Component: Dashboard },
      { path: "missions/:slug", Component: MissionDetail },
      { path: "courses/go-basics", Component: GoBasicsCourse },
      { path: "courses/go-basics/:chapterSlug", Component: GoBasicsChapter },
    ],
  },
]);
