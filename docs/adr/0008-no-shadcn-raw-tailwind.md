# 0008 — No shadcn/ui; render with raw Tailwind utilities

We are not adopting `shadcn/ui` for component primitives. Frontend components are written as plain React + raw Tailwind utility classes on the JSX (`bg-slate-800`, `border-slate-700`, `rounded-md`, `p-4`, `space-y-2`, etc.). This supersedes the "Components: shadcn/ui (Radix primitives + Tailwind)" row of ADR-0007 for the v1 frontend; the rest of ADR-0007 (Tailwind v4, CVA-free Button, lucide-react for icons, `useState` for state) is unchanged.

## Context

ADR-0007 picked shadcn/ui on the assumption that v1 would need dialogs, dropdowns, and other Radix primitives. In practice the v1 surface is two screens (Dashboard placeholder, Server Profiles CRUD), and shadcn was only used to style a form + a card on one of them. Every shadcn semantic token (`bg-primary`, `bg-card`, `text-muted-foreground`, ...) had no CSS-variable definition in the project, so the form worked only because every consumer site overrode the tokens with raw `bg-slate-*` classes. The Radix dialog and dropdown packages were installed but never imported; `react-hook-form` and `zod` likewise. The shadcn setup was paying its complexity budget without buying anything.

## Decision

Render directly with raw Tailwind utilities. Keep `lucide-react` for icons (one import, tree-shaken). Keep `clsx` for class-name composition in the one place that needs it. Drop `components.json`, the `components/ui/` directory, the `cn` helper, the unused Radix packages, and the unused form packages.

## Considered Options

- **shadcn semantic tokens, properly defined** — emit `--background`, `--card`, `--border`, ... in `style.css`. Would make the shadcn components *work* as designed, but still pays the cost of 4 hand-maintained components for two screens and an import path that future contributors have to learn.
- **A different component library (Radix Themes, Mantine, MUI)** — heavyweight, designed for app shells with many primitives we don't have. Lock-in on a styling API that fights Tailwind utilities.
- **CSS variables in `:root` for our slate palette** — adds an indirection for a palette of six colors used in two files. Pure over-engineering at this size.

## Consequences

- The styling pattern is "Tailwind utilities all the way down" — no CSS variables, no component library, no design tokens beyond Tailwind's built-in palette.
- `src/frontend/src/components/` exists but stays small; new shared components are added only when a second screen needs them (YAGNI applies in both directions).
- A future contributor who wants to reach for shadcn should re-open this ADR rather than reinstall it. The justification ("we need dialogs / forms / a11y primitives") should be concrete before the dependency tree grows back.
