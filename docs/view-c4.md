# View a C4 model

C4 files end in `.c4` and hold one diagram per level (split on `---`). Spec: [[c4-format]]. Dogfood model: [[c4-peel]].

New to C4? Start with the slide tour [[deck-c4]] (why it matters with agents, then click vs drill).

## Open a model

Files drawer → **C4** tab, or Goto Anything (`Ctrl/Cmd-P`) and type `c4`. The meta line shows path · c4 · level count. URL: `?file=c4-peel.c4&level=2`.

![C4 context](assets/c4.png)

## Switch levels

| Control | Action |
|---|---|
| Level dropdown | Jump by name (kind icon beside each entry) |
| `←` / `→` | Previous / next level |
| `Home` / `End` | First / last |
| Outline | Jump to a titled level |

Breadcrumb shows the current level title. Kind icons: Context, Container, Component, Code (`C4Code` **or** `C4Dynamic`), Deployment.

![C4 containers](assets/c4-containers.png)

## Comment or drill

| Gesture | Result |
|---|---|
| **Click** box or Rel | Composer opens; note may carry `c4Id`, `c4Label`, `c4From`/`c4To`, `level`, `levelKind` |
| **Double-click** box | Drill one level deeper when a deeper level exists |
| Select label text | Same path as doc selection |

Single-click waits briefly so a double-click does not also open the composer. Canvas title hint: “Click to comment · double-click to drill”.

![Gestures (deck)](assets/deck-c4-gestures.png)

## Zoom

`−`, **Fit**, `+`, **Reset**. Canvas fits width on open and on level switch. Manual zoom disables auto-refit until **Fit**.

## Agent rematch

After you comment a Rel, the agent should prefer `c4From`/`c4To` (or the edge `c4Id`) over the quote string if labels change. Playbook: [[agent-loop]].
