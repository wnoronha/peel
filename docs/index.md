# Peel

One Go binary that turns a folder of markdown into a readable site — with slide decks, C4 models, and an annotation loop agents can actually consume.

Open this folder with peel and you are dogfooding the product.

```sh
./peel --dir ./docs --port 18124
```

## Start here

| You want | Open |
|---|---|
| Five-minute tour | [[deck-peel-tour]] |
| Why peel exists | [[use-cases]] |
| Architecture as a model | [[c4-peel]] |
| Agent review loop | [[agent-loop]] |
| Hands-on first note | [[read]] |

## Modes

- **Docs** — rendered markdown, outline, backlinks, Goto Anything (`Ctrl/Cmd-P` or `K`)
- **Decks** — `marp: true` files split on `---` into focusable slides · [[view-decks]]
- **C4** — `*.c4` levels with kind icons, zoom, drill-down · [[view-c4]]
- **Annotate** — select text → note → `.peel/feedback.jsonl` · [[annotate]]

## Specs (deep lane)

[[feedback-contract]] · [[api]] · [[c4-format]] · [[architecture]] · [[serve]]

## Showcase tip

Walk a visitor through [[deck-peel-tour]], then [[c4-peel]], then leave one open note on [[use-cases]] and show the Feedback drawer (open / acked / dismissed). That is the product in three minutes.
