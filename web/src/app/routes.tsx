import { lazy } from 'react'
import { createBrowserRouter } from 'react-router'

import { LearningLayout, PublicLayout } from './components/Layout'

const Landing = lazy(() => import('./pages/Landing').then((module) => ({ default: module.Landing })))
const Dashboard = lazy(() => import('./pages/Dashboard').then((module) => ({ default: module.Dashboard })))
const CapabilityActivity = lazy(() => import('./pages/CapabilityActivity').then((module) => ({ default: module.CapabilityActivity })))
const NotFound = lazy(() => import('./pages/NotFound').then((module) => ({ default: module.NotFound })))

export const router = createBrowserRouter([
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
    ],
  },
  { path: '/learning/activities/:activityId', Component: CapabilityActivity },
  { path: '*', Component: NotFound },
])
