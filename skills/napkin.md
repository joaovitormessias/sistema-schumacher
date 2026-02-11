# Napkin

You maintain a per-repo markdown file that tracks mistakes, corrections, and
patterns that work or don't. You read it before doing anything and update it
continuously as you work — whenever you learn something worth recording.

**This skill is always active. Every session. No trigger required.**

## Where the napkin lives (agent-agnostic)

Use a tool-agnostic path by default:

* Primary (preferred): `skills/napkin.md`

For compatibility with existing repos/tools:

* If `.claude/napkin.md` already exists, keep using it.
* Else if `.codex/napkin.md` exists, keep using it.
* Else if `.cursor/napkin.md` exists, keep using it.

If none exist, create the primary path and use it going forward.

**Rule:** always read/write exactly one napkin file per repo (pick using the
precedence above) so knowledge stays consolidated.

## Session Start: Read Your Notes

First thing, every session — read the napkin file before doing anything else.
Internalize what's there and apply it silently. Don't announce that you read it.
Just apply what you know.

If no napkin exists yet, create one at the chosen path (usually
`skills/napkin.md`):

```markdown
# Napkin

## Corrections
| Date | Source | What Went Wrong | What To Do Instead |
|------|--------|----------------|-------------------|

## User Preferences
- (accumulate here as you learn them)

## Patterns That Work
- (approaches that succeeded)

## Patterns That Don't Work
- (approaches that failed and why)

## Domain Notes
- (project/domain context that matters)
```

Adapt the sections to fit the repo's domain. Design something you can usefully
consume.

## Continuous Updates

Update the napkin as you work, not just at session start and end. Write to it
whenever you learn something worth recording:

* **You hit an error and figure out why.** Log it immediately. Don't wait.
* **The user corrects you.** Log what you did and what they wanted instead.
* **You catch your own mistake.** Log it. Your mistakes count the same as
  user corrections — maybe more, because you're the one who knows what went
  wrong internally.
* **You try something and it fails.** Log the approach and why it didn't work
  so you don't repeat it.
* **You try something and it works well.** Log the pattern.
* **You re-read the napkin mid-task** because you're about to do something
  you've gotten wrong before. Good. Do this.

The napkin is a living document. Treat it like working memory that persists
across sessions, not a journal you write in once.

## What to Log

Log anything that would change your behavior if you read it next session:

* **Your own mistakes**: wrong assumptions, bad approaches, misread code,
  failed commands, incorrect fixes you had to redo.
* **User corrections**: anything the user told you to do differently.
* **Tool/environment surprises**: things about this repo, its tooling, or its
  patterns that you didn't expect.
* **Preferences**: how the user likes things done — style, structure, process.
* **What worked**: approaches that succeeded, especially non-obvious ones.

Be specific. "Made an error" is useless. "Assumed the API returns a list but
it returns a paginated object with `.items`" is actionable.

## Napkin Maintenance

Every 5-10 sessions, or when the file exceeds ~150 lines, consolidate:

* Merge redundant entries into a single rule.
* Promote repeated corrections to User Preferences.
* Remove entries that are now captured as top-level rules.
* Archive resolved or outdated notes.
* Keep total length under 200 lines of high-signal content.

A 50-line napkin of hard-won rules beats a 500-line log of raw entries.

## Example

**Early in a session** — you misread a function signature and pass args in the
wrong order. You catch it yourself. Log it:

```markdown
| 2026-02-06 | self | Passed (name, id) to createUser but signature is (id, name) | Check function signatures before calling; this codebase doesn't follow conventional arg ordering |
```

**Mid-session** — user corrects your import style. Log it:

```markdown
| 2026-02-06 | user | Used relative imports | This repo uses absolute imports from `src/` — always |
```

**Later** — you re-read the napkin before editing another file and use
absolute imports without being told. That's the loop working.

## Session Notes (2026-02-10)

### Corrections
| Date | Source | What Went Wrong | What To Do Instead |
|------|--------|-----------------|--------------------|
| 2026-02-10 | self | Used `ON CONFLICT (trip_id, booking_passenger_id)` with only partial unique index; migration failed (`42P10`) | Create a concrete unique constraint matching conflict columns when using `ON CONFLICT` |
| 2026-02-10 | self | Attempted to write very large files in one shell command on Windows and hit `code 206` | Split writes into smaller chunks or use incremental patching |
| 2026-02-10 | self | Booking form inputs used `setForm({ ...form, ... })` while fare quote updates `total_amount` asynchronously, causing intermittent overwrite to zero | In Booking steps, always use functional updates `setForm((prev) => ({ ...prev, ... }))` for field edits |

### Patterns That Work
- Validate backend-wide changes with `gofmt` + `go test ./...` before touching production DB.
- Keep workflow blockers explicit (`requirements_missing`) so frontend can render pending compliance items.
- Apply Supabase migration via MCP after local checks to keep code/schema synchronized.
- For backend observability in this repo, emit event= and metric= lines via log.Printf in handlers/services (current production pattern).

### User Preferences
- 2026-02-10: When discussing major flow issues (ex.: criação de rotas), do a full code-flow analysis first and do not edit product code in that turn.
- 2026-02-10: For business-rule discussions, bring external references from the web (regulatory/market) to ground recommendations.
- 2026-02-10: In operational forms (ex.: Rotas/Paradas), ambiguous numeric fields must be explicitly labeled with business meaning.

### Domain Notes
- Current app flow has route stop entities in DB/API, but `Routes` UI still does only quick route header creation (`name`, `origin_city`, `destination_city`) without stop configuration.
- Downstream modules (trip stops, booking boarding/alighting, segment pricing, operational workflow) assume stops are configured; missing stop setup upstream creates operational bottlenecks.
- `docs-sistema/Procedimentos Operacionais Fretamento.md` reinforces that operations start from request intake with boarding point and schedule, then route/itinerary organization (including all trip points), passenger list checks, D-1 vehicle/driver assignment, and authorization workflow (DETER/ANTT) before departure.
- Fretamento workflow also requires pre-trip travel folder/document checklist and post-trip reconciliation, so route/trip setup should capture enough stop/timing detail for authorization and execution artifacts.
- Booking auto fare depends on `segment_fares`; repo currently has quote logic but no API/UI CRUD for `segment_fares`, so in many real scenarios checkout stays sem tarifa automatica until DB has active segment fare data.
