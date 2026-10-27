# View a C4 model

C4 files end in `.c4` and hold one diagram per level (split on `---`). Spec: [[c4-format]]. Dogfood model: [[c4-peel]].

## Switch levels

Use the level dropdown (kind icons for context / container / component / dynamic / code) or press `←` / `→`. The breadcrumb shows where you are. The URL carries `?file=….c4&level=2`, so levels are shareable.

## Drill into a box

Click any box to go one level deeper. The outline jumps directly to a named level.

## Zoom

`−`, **Fit**, `+`, **Reset**. Canvas fits width on open and on level switch. Manual zoom disables auto-refit until **Fit**.

## Annotate

Select node or edge labels → note as usual. Anchors may include `level` and `levelKind` for agents ([[agent-loop]]).

![C4 view](assets/c4.png)
