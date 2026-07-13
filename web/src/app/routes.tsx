import { lazy } from "react";
import { createBrowserRouter } from "react-router";
import { PublicLayout, LearningLayout } from "./components/Layout";

const Landing = lazy(() => import("./pages/Landing").then((module) => ({ default: module.Landing })));
const Dashboard = lazy(() => import("./pages/Dashboard").then((module) => ({ default: module.Dashboard })));
const MissionDetail = lazy(() => import("./pages/MissionDetail").then((module) => ({ default: module.MissionDetail })));
const GoBasicsCourse = lazy(() => import("./pages/GoBasicsCourse").then((module) => ({ default: module.GoBasicsCourse })));
const GoBasicsChapter = lazy(() => import("./pages/GoBasicsChapter").then((module) => ({ default: module.GoBasicsChapter })));
const CapabilityActivity = lazy(() => import("./pages/CapabilityActivity").then((module) => ({ default: module.CapabilityActivity })));

export const router = createBrowserRouter([
  {
    path: "/",
    Component: PublicLayout,
    children: [
      { index: true, Component: Landing },
    ],
  },
  {
    Component: LearningLayout,
    children: [
      { path: "/dashboard", Component: Dashboard },
      { path: "/missions/:slug", Component: MissionDetail },
      { path: "/courses/go-basics", Component: GoBasicsCourse },
      { path: "/courses/go-basics/:chapterSlug", Component: GoBasicsChapter },
    ],
  },
  { path: "/learning/activities/:activityId", Component: CapabilityActivity },
]);
