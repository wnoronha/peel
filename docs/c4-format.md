# C4 format spec

C4 files use the `.c4` extension and plain text. They are not markdown. Frontmatter is not read. Dogfood model: [[c4-peel]]. Viewer guide: [[view-c4]]. Focused deck: [[deck-c4]].

## Levels

One C4 diagram block makes one level. Levels split on lines that contain exactly `---`:

```
C4Context
title Search context
Person(user, "User", "issues queries")
System(search, "Search", "crawl, index, serve")
Rel(user, search, "queries")
---
C4Container
title Search containers
Container(blend, "Blender", "merges organic and ads")
```

The viewer lists levels in file order (dropdown + ←/→). Empty sections are dropped. URLs carry `?file=….c4&level=2` (1-based). Outline entries use each block’s `title`.

## Titles

The first `title <name>` line in a block names the tab, breadcrumb, and outline entry. A block without one is called `Level N`. The file title is the first `title` line in the file, else the filename without extension.

## Diagram types → kind icons

Each block is one mermaid C4 diagram, rendered by the vendored pele bundle. The server maps the opening directive to a **kind** for the UI:

| Directive | Kind shown |
|---|---|
| `C4Context` | Context |
| `C4Container` | Container |
| `C4Component` | Component |
| `C4Code` | Code |
| `C4Dynamic` | Code *(treated as the fourth classic level)* |
| `C4Deployment` | Deployment |

`%%` comment lines are allowed. Keep one `Rel` / `Rel_U` / … per line; back-edges pele cannot parse fail that block with a `diagram error` message.

### Element ids matter

Pele uses the first argument of `Person` / `System` / `Container` / `Component` / `Rel` as the element id. Feedback stores that as `c4Id` (edges often look like `user-peel`). Prefer stable ids so agents can rematch after label edits.

```
Person(user, "User", "…")          %% c4Id = user
Rel(user, peel, "CLI + browser")   %% edge c4Id ≈ user-peel; c4From=user c4To=peel
```

## Interaction (viewer)

| Gesture | Result |
|---|---|
| **Click** box or relationship | Opens comment composer; note may include `c4Id`, `c4Label`, and for edges `c4From` / `c4To`, plus `level` / `levelKind` |
| **Double-click** box | Drill one level deeper (when a deeper level exists) |
| Select label text | Same composer path as docs |

Outline jumps go straight to a named level. Zoom: `−` / Fit / `+` / Reset.

## Limits

C4 sources have no `[[wikilinks]]`. Cross-file references are plain text. Agents should match on `c4Id` / Rel endpoints when quotes drift — see [[feedback-contract]] and [[agent-loop]].

## Dogfood levels in this repo

[[c4-peel]] ships Context → Container → Web UI components → Go server components → reload dynamic → note/ack dynamic → fold Code. Use it as the template for new models.
