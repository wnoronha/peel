# View a deck

Decks are markdown with `marp: true` in the frontmatter. Each `---` section is one slide. Side panels start closed so slides get the width.

Dogfood decks: [[deck-peel-tour]], [[deck-use-cases]].

## Move between slides

- `←` / `→` in the control bar or on the keyboard
- `Home` / `End` for first and last
- Click a collapsed title strip to jump
- Outline or headings dropdown in the meta line

The counter reads `3 / 13`. Scroll-spy keeps it honest when you wheel by hand.

## Link to a slide

URL carries `?file=deck-peel-tour.md&slide=3`. Share or reload lands on that slide. Browser back leaves the deck page, not the previous slide index.

## Annotate

Select text or diagram labels on the open slide, then save as usual. Jumping from Feedback expands the right slide first. Anchors may include `slide`.

![Deck view](assets/deck.png)
