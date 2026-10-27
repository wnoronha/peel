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
| Why peel exists | [[use-cases]] · [[deck-use-cases]] |
| Architecture as a model | [[c4-peel]] · [[deck-c4]] |
| Agent review loop | [[agent-loop]] |
| Hands-on first note | [[read]] |

## Modes

- **Docs** — rendered markdown, outline, backlinks, Goto Anything (`Ctrl/Cmd-P` or `K`) · [[read]]
- **Decks** — `marp: true` files split on `---` into focusable slides · [[view-decks]]
- **C4** — `*.c4` levels; click = comment, double-click = drill · [[view-c4]]
- **Feedback** — select / C4 click → note → `.peel/feedback.jsonl` · [[annotate]]

## Specs (deep lane)

[[feedback-contract]] · [[api]] · [[c4-format]] · [[architecture]] · [[serve]] · [[decisions]]

## Showcase tip

Walk a visitor through [[deck-peel-tour]], then [[c4-peel]] (click a Rel, double-click a box), then leave one open note and show Feedback (open / acked / dismissed) plus a toast when the agent acks. That is the product in three minutes.
