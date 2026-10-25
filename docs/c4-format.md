# C4 format spec

C4 files use the `.c4` extension and plain text. They are not markdown. Frontmatter is not read.

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

The viewer renders levels as tabs in file order. Clicking a box drills one level deeper. Empty sections are dropped.

## Titles

The first `title <name>` line in a block names the tab, the breadcrumb, and the outline entry. A block without one is called `Level N`. The file title is the first `title` line in the file, else the filename without extension.

## Diagram types

Each block is one mermaid C4 diagram, rendered by the vendored pele bundle: `C4Context`, `C4Container`, `C4Component`, or `C4Dynamic`. `%%` comment lines are allowed. Relationship lines that pele cannot parse fail that block with a `diagram error` message, so keep one relation per line and avoid back-edges between the same pair.

## Limits

C4 sources have no `[[wikilinks]]`. References to other files are plain text. Annotations still work: quotes are captured from diagram labels and `startLine` maps onto the matching source line where possible.
