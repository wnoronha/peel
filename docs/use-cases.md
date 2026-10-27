# Use cases

Peel pays off when docs are the product of record and agents do the grunt edits. Four jobs it is built for.

## 1. Architecture review in the browser

**Who:** Staff eng, tech lead, new hire  
**Files:** `*.c4` plus narrative `*.md`  
**Flow:** Open [[c4-peel]], walk Context → Container → Component → Code. Click a box to drill. Leave notes on labels that are wrong (“this store is Postgres, not Redis”).

Agents read those notes with `level` / `levelKind` anchors and patch the `.c4` source. You refresh; the diagram updates.

## 2. Agent-assisted doc fix loop

**Who:** Writer + Cursor / Claude / Codex  
**Files:** any markdown tree  
**Flow:** While reading [[architecture]], select a stale sentence → chip **Fix this** → save. The agent tails `.peel/feedback.jsonl`, edits the file, `POST`s an `ack`. You switch the drawer to **Acked** and confirm.

See [[agent-loop]] for the exact steps and event shapes.

## 3. Onboarding as a deck

**Who:** Team lead onboarding someone  
**Files:** `marp: true` decks  
**Flow:** Present [[deck-peel-tour]] or [[deck-use-cases]] full-screen. Arrows move slides. When someone asks “where do notes go?”, jump to [[feedback-contract]] via Goto Anything without leaving peel.

## 4. Spec + implementation side by side

**Who:** Pair building a format  
**Files:** [[c4-format]], [[feedback-contract]], [[api]]  
**Flow:** Keep the contract open, annotate ambiguities, let the agent update both the Go handlers and the markdown in one pass. Backlinks show which guides reference the contract.

## What peel is not for

- Long-form publishing with auth and multi-tenant hosting  
- Replacing Linear/Jira for project tracking  
- Free-form JS experiments in the browser (keep scenes declarative if you add that later)

## Try it here

1. Open [[c4-peel]] and leave a note on a relationship label.  
2. Open Feedback → confirm it is **Open**.  
3. Have an agent ack it → it moves to **Acked**.  
4. Dismiss a junk note → **Dismissed**.

That filter row is the collaboration status surface.
