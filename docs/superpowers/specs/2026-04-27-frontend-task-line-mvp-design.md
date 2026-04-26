# Frontend Task Line MVP Design

- Date: 2026-04-27
- Status: approved for implementation planning
- Scope: frontend-only Day 0 to Day 5 intern task line

## Summary

GoGopher Arch now has the right product positioning and an intern workbench, but the workbench still behaves like a single hard-coded sample. The next step is to make the first-week learning loop real by adding a selectable Day 0 to Day 5 task line in the frontend.

This MVP keeps tasks as static frontend data. It does not add a database, task API, authentication, AI feedback, or a backend grading engine. The goal is to prove the product shape first: a learner can pick a task, read the task card and short lesson, edit Go code, run the sandbox, see task-specific feedback, and read a short review.

## Goals

- Provide a complete first-week task line from Day 0 through Day 5.
- Keep each task self-contained with background, acceptance criteria, short lesson, mentor hints, default code, checks, and review notes.
- Refactor the current hard-coded Day 1 workbench into data-driven task rendering.
- Keep the current `/api/v1/execute` sandbox request and response contract unchanged.
- Keep task checks in the frontend for this iteration, with a clean boundary for moving checks server-side later.

## Non-Goals

- No new Go task API.
- No database or saved user progress.
- No login, accounts, or multi-user state.
- No LLM, RAG, Agent, or AI mentor integration.
- No full browser end-to-end testing framework.
- No redesign of the whole visual system.

## User Experience

The first screen remains an intern workbench. The visible change is that the learner can choose tasks across the first week instead of seeing only Day 1.

The workbench shows:

- A compact task navigation list for Day 0 to Day 5.
- A task card with background, objective, and acceptance criteria.
- A short lesson for the current task.
- Monaco Editor preloaded with the task's starter code.
- A run button that sends the current code to the existing Gateway.
- Task feedback derived from sandbox status, stdout, stderr, and task-specific checks.
- Mentor hints and a review section for the current task.

When the learner switches tasks, the editor loads that task's starter code and clears the previous output. This avoids confusing stale feedback from another task.

## Task Set

### Day 0: First Sandbox Run

Purpose: help beginners confirm they can run Go code and read stdout.

Starter code prints a simple onboarding message. The check passes when the sandbox succeeds and stdout contains the expected phrase.

### Day 1: Fix Nil Map Write

Purpose: teach map initialization and Go zero values.

Starter code contains a nil map write. The check passes when the code exits successfully and stdout includes the expected user score.

### Day 2: Complete JSON Handler Output

Purpose: practice structs, JSON tags, and HTTP-handler-style response shaping without adding a backend exercise harness.

Starter code asks the learner to implement a function that marshals a response struct. The check passes when stdout contains JSON with the expected `id` and `name` fields.

### Day 3: Add Validation And Error Return

Purpose: practice input validation and explicit error handling.

Starter code accepts an empty request name. The learner updates validation so invalid input returns a clear error. The check passes when stdout contains a validation error and valid input still succeeds.

### Day 4: Write Table-Driven Checks

Purpose: introduce table-driven thinking inside a runnable Go program.

Because the sandbox currently runs `go run`, not `go test`, the MVP models table-driven tests as a small loop over cases that prints `PASS` for each case. The check passes when all cases print success.

### Day 5: Respect Context Timeout

Purpose: introduce context cancellation and timeout-aware code.

Starter code simulates a slow operation. The learner updates the operation to respect `ctx.Done()`. The check passes when stdout shows the timeout path instead of waiting for the slow path to finish.

## Frontend Architecture

### `web/src/tasks.ts`

Create a focused task catalog module.

Responsibilities:

- Define `InternTask`, `TaskCheck`, and related types.
- Export `internshipTasks`.
- Provide enough structured data for rendering and feedback.

The task data should include:

- `id`
- `day`
- `title`
- `track`
- `summary`
- `background`
- `objective`
- `starterCode`
- `criteria`
- `lesson`
- `mentorHints`
- `review`
- `checks`

Checks are declarative and intentionally small. Supported MVP check types:

- stdout includes a string
- stdout matches a regular expression
- stderr does not include a string
- sandbox exit succeeds

This keeps `App.tsx` from knowing task-specific strings.

### `web/src/App.tsx`

Refactor the current component to render data-driven tasks.

Responsibilities:

- Track selected task id.
- Reset code and output when the selected task changes.
- Submit code to `/api/v1/execute` using the current contract.
- Convert sandbox output and the current task's checks into feedback rows.
- Render task navigation, task card, lesson, editor, feedback, hints, console, and review.

`App.tsx` should not contain the Day 0 to Day 5 task content directly. It may contain generic feedback rendering and sandbox error handling.

### `web/src/App.css`

Extend the existing workbench CSS.

Responsibilities:

- Add task navigation styles.
- Add selected and passed/failed states for task items.
- Add a compact review section.
- Preserve responsive behavior for desktop and mobile.

No major visual redesign is needed. The interface should stay quiet, practical, and workbench-like.

## Data Flow

1. User selects a task.
2. `App` loads `task.starterCode` into Monaco and clears output.
3. User edits code.
4. User clicks run.
5. Frontend posts `{ id, code, language: "go", timeout }` to the existing Gateway.
6. Gateway forwards to the sandbox engine.
7. Sandbox returns stdout, stderr, status, duration, and exit code.
8. Frontend evaluates the current task's checks.
9. Workbench shows connection status, run status, task check status, console output, hints, and review.

## Error Handling

- If the Gateway request fails, the feedback panel should show a connection failure and keep task checks idle.
- If the sandbox returns an error status or non-zero exit code, the run check fails and task-specific checks fail unless their definition explicitly only needs stderr.
- If a task check fails, the learner sees the check label and a short hint, not an implementation answer.
- Switching tasks clears old errors and output.

## Testing And Verification

Manual verification for this MVP:

- `cd web && npm run build`
- `go test ./...`
- Search for Day 0 through Day 5 task labels in `web/src`.
- Search for new task-line UI labels such as `任务列表`, `任务后复盘`, and `导师提示`.
- Confirm old advanced-positioning terms still do not appear in production-visible README/docs/frontend files.

The implementation should keep code structured so future tests can target check evaluation separately if a test runner is added.

## Future Migration Path

This static frontend task catalog is a deliberate stepping stone.

Later, the project can move the same shape to:

- A backend `/api/v1/tasks` endpoint.
- A persisted progress model.
- A backend task-check runner.
- Real `go test` execution support in the sandbox.
- AI-assisted mentor feedback using RAG or Agent workflows.

The MVP should avoid decisions that make those migrations harder. The task data shape should therefore look like a future API response, even while it lives in TypeScript.
