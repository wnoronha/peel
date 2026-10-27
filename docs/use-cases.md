# Use cases

Peel pays off when docs are the product of record and agents do the grunt edits. Four jobs it is built for.

## 1. Architecture review in the browser

**Who:** Staff eng, tech lead, new hire  
**Files:** `*.c4` plus narrative `*.md`  
**Flow:** Open [[c4-peel]] (gestures: [[deck-c4]]), walk Context → Container → Component → Dynamic/Code. **Click** a box or Rel to comment (“this store is Postgres, not Redis”). **Double-click** a box to drill when a deeper level exists.

Agents receive `c4Id` / `c4From` / `c4To` plus `level` / `levelKind`, patch the `.c4` source, and ack. You get a toast and the diagram reloads.

## 2. Agent-assisted doc fix loop

**Who:** Writer + any coding agent (Cursor, Claude, Codex, …)  
**Files:** any markdown tree  
**Flow:** While reading [[architecture]], select a stale sentence → chip **Fix this** → Save. Run:

```sh
skills/peel/peel_api.py pair --timeout 600
```

The agent prints the note JSON, edits the file, `ack`s. Feedback → **Acked**; toast confirms.

See [[agent-loop]] for event shapes and catch-up rules.

## 3. Onboarding as a deck

**Who:** Team lead onboarding someone  
**Files:** `marp: true` decks  
**Flow:** Present [[deck-peel-tour]] or [[deck-use-cases]] full-screen. Arrows move slides. When someone asks “where do notes go?”, jump to [[feedback-contract]] via Goto Anything (`Ctrl/Cmd-P`) without leaving peel.

## 4. Spec + implementation side by side

**Who:** Pair building a format  
**Files:** [[c4-format]], [[feedback-contract]], [[api]]  
**Flow:** Keep the contract open, annotate ambiguities, let the agent update both the Go handlers and the markdown in one pass. Backlinks show which guides reference the contract.

## What peel is not for

- Long-form publishing with auth and multi-tenant hosting  
- Replacing Linear/Jira for project tracking  
- Live multiplayer whiteboarding (not in this binary)

## Try it here

1. Open [[c4-peel]] and **click** a relationship to leave a note.  
2. Open Feedback → confirm it is **Open** (peach mark / toast if remote).  
3. Have an agent ack via `peel_api.py` → **Acked** + toast.  
4. Dismiss a junk note → **Dismissed**.

That filter row is the collaboration status surface.
