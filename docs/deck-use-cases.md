---
marp: true
title: Peel use cases
---

# Four jobs for peel

Architecture review.
Agent doc fixes.
Onboarding decks.
Spec + code side by side.

Full write-up: [[use-cases]]

---

## 1 · Architecture review

Models live in dedicated `*.c4` files rendered in Peel's interactive viewer.
Walk levels, drill into boxes, and annotate labels for agents to patch.

Open model: [[c4-peel]]

```mermaid
flowchart LR
  Context["1 · Context"] --> Container["2 · Container"] --> Component["3 · Component"] --> Code["4 · Code"]
```

---

## 2 · Agent fix loop

Select a stale sentence.
Chip **Fix this**. Save.
Agent edits + acks. Drawer → **Acked**.

Playbook: [[agent-loop]]

---

## 3 · Onboarding as a deck

Present with ← / →.
`Ctrl-P` jumps into a deep spec without leaving peel.

Tour: [[deck-peel-tour]]

---

## 4 · Spec beside the build

Keep [[feedback-contract]] or [[c4-format]] open.
Annotate ambiguities.
Agent updates Go + markdown in one pass.

---

## Status is the drawer

**Open** → work queue  
**Acked** → agent done  
**Dismissed** → human discarded  

Marks only for open notes on the current file.

---

## Not for

Wiki hosting. Ticket tracking. Multi-tenant CMS.

Peel is the thin loop: quote → note → edit → ack.

---

## Run the dogfood

```sh
./peel --dir ./docs --port 18124
```

Leave a note here. That is the product.
