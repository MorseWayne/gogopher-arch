import { lazy } from 'react'
import { createBrowserRouter } from 'react-router'

import { LearningLayout, PublicLayout } from './components/Layout'
import { RouteError } from './pages/RouteError'

const Landing = lazy(() => import('./pages/Landing').then((module) => ({ default: module.Landing })))
const Dashboard = lazy(() => import('./pages/Dashboard').then((module) => ({ default: module.Dashboard })))
const Roadmap = lazy(() => import('./pages/Roadmap').then((module) => ({ default: module.Roadmap })))
const CapabilityActivity = lazy(() => import('./pages/CapabilityActivity').then((module) => ({ default: module.CapabilityActivity })))
const GoBasicsCourse = lazy(() => import('./pages/GoBasicsCourse').then((module) => ({ default: module.GoBasicsCourse })))
const GoBasicsChapter = lazy(() => import('./pages/GoBasicsChapter').then((module) => ({ default: module.GoBasicsChapter })))
const NotFound = lazy(() => import('./pages/NotFound').then((module) => ({ default: module.NotFound })))

export const router = createBrowserRouter([
  {
    errorElement: <RouteError />,
    children: [
      {
        path: '/',
        Component: PublicLayout,
        children: [
          { index: true, Component: Landing },
        ],
      },
      {
        Component: LearningLayout,
        children: [
          { path: '/dashboard', Component: Dashboard },
          { path: '/roadmap', Component: Roadmap },
          { path: '/courses/go-basics', Component: GoBasicsCourse },
          { path: '/courses/go-basics/:chapterSlug', Component: GoBasicsChapter },
        ],
      },
      { path: '/learning/activities/:activityId', Component: CapabilityActivity },
      { path: '*', Component: NotFound },
    ],
  },
])
