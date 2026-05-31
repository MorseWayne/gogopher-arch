---
name: go-course-chapter-redesign
description: Redesign a Go course chapter using GoGopher Arch's tutorial-grade course design standard.
---

# Go Course Chapter Redesign

Use this skill when the user asks to improve, enrich, rewrite, standardize, or continue redesigning a Go course chapter.

## Core principles

- Keep course content in-app and MDX-first. Do not replace tutorials with external link collections.
- Do not invent course content from scratch. Ground the chapter in local `gopl-zh.github.com` source where relevant, then summarize, reorganize, modernize, and adapt it to GoGopher Arch.
- Use official Go docs to verify modern toolchain behavior and newer features.
- Use Go by Example for short runnable example style, not as a copy source.
- Use Learn Go with Tests, Exercism, Gophercises, Effective Go, and style guides as exercise/design inspiration, not as pasted content.
- Turn external material into GoGopher Arch's own learning path: scenario, fundamentals, examples, pitfalls, engineering practice, recap, exercises.

## Preferred chapter flow

Do not front-load dense concept maps. Use this order by default:

1. Concrete scenario introduction.
2. Problem the learner will solve.
3. Source notes and how sources are integrated.
4. Step-by-step fundamentals.
5. Minimal runnable examples.
6. Engineering-oriented examples.
7. Pitfalls and anti-examples.
8. Practice bridges tied to exercises.
9. Concept recap / concept map after the learner has context.
10. Engineering perspective and review checklist.

## Fundamentals checklist

For every core concept, cover these before moving too far into engineering abstraction:

- Definition form: syntax and naming shape.
- Initialization or usage form: zero value, literal, constructor, make/new, command, or function signature.
- Semantic model: value vs reference behavior, memory model, lifecycle, command execution model, or toolchain behavior.
- Common mistakes: what beginners usually misunderstand.
- Minimal example: short and runnable.
- Engineering example: how this appears in backend work.

## MDX content components

Use existing MDX components when they improve clarity:

- `SourceNote`: identify a source and explain how it was used.
- `CompareNote`: compare teaching approaches or clarify trade-offs.
- `ExamplePair`: show minimal vs engineering version, bad vs good, or before vs after.
- `PitfallCard`: highlight symptom and fix.
- `DeepDive`: keep optional advanced material out of the main flow.
- `PracticeBridge`: connect a concept to a concrete exercise.

Avoid adding new components unless the existing ones are insufficient.

## Exercise design

Each redesigned chapter should move toward layered exercises:

- `warmup`: verify the basic concept can run.
- `core`: solve a realistic small problem.
- `challenge`: debug, refactor, test, review, or handle edge cases.

Exercise prompts should connect to the chapter scenario. Hints should be progressive. Solution outlines should check reasoning, not just reveal code.

If CodeMirror snippets are relevant, ensure exercise `concepts` contain useful keywords such as `map`, `sort`, `strings.Builder`, `table-driven tests`, `regression test`, or similar.

## Source workflow

1. Read the current chapter MDX.
2. Read corresponding local `gopl-zh.github.com` chapter files.
3. Check official Go docs only when behavior may be modern-version-sensitive.
4. Identify what should be basic explanation, what should be engineering practice, and what belongs in a deep dive.
5. Rewrite in MDX using the preferred flow.
6. Keep source attribution clear through `SourceNote` and metadata where already present.

## Validation

For code/content changes in this project:

- Run `npm run build --prefix web` after MDX or frontend changes.
- Run `git diff --check`.
- Note Vite chunk warnings honestly; they are non-blocking unless the build fails.
- Update `.claude/WORKFLOW.md` for Level 2/3 tracked course work.

## Anti-patterns

- Starting an article with a dense concept map before examples.
- Copying source material without reorganization or attribution.
- Writing only engineering scenarios while skipping definitions and basic syntax.
- Listing external recommendations instead of integrating them into the lesson.
- Adding exercises disconnected from the text.
- Treating coverage, benchmark numbers, or output matching as goals without explaining learner value.
