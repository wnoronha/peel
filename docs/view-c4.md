# View a C4 model

C4 files end in `.c4` and hold one C4 diagram per level. The format spec is in `c4-format.md`.

## Switch levels

Click the level tabs (`Search context`, `Search containers`, and so on) or press `←` and `→`. The breadcrumb shows where you are. The URL carries `?file=c4-search.c4&level=2`, so levels are shareable and survive reload.

## Drill into a box

Click any box to go one level deeper, the same gesture as the zoom icon in IcePanel. The sidebar outline and the levels dropdown jump directly instead.

## Zoom the canvas

Use `−`, `Fit`, `+`, and `Reset` in the control bar. The canvas fits the width on open and on every level switch. If you zoom by hand, resizing the window no longer refits until you press **Fit** again.

## Annotate a diagram

Drag-select node or edge labels and save a note as usual. The feedback list searches across levels when you jump back to a note.

![C4 view](assets/c4.png)
