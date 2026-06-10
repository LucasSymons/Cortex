---
name: promote-lessons
description: Review project-level lessons.md files and promote relevant entries to the top-level Cortex profile memory. Runs on sync or manually.
---

# Promote project lessons to top-level memory

You are reviewing lessons learned in project-level memory files and deciding which ones are worth promoting to the user's top-level Cortex profile memory, so they are available in all future sessions across all projects.

## When to run

- Automatically as part of sync-profile on session handoff
- Manually when the user asks to promote lessons

## What to promote

Promote an entry if it meets any of these criteria:
- A pattern or gotcha that would apply to OTHER projects (not just this one)
- A tool behaviour that is surprising and worth remembering broadly
- An architectural decision with rationale that influences future designs
- A process or policy detail that applies across the organisation

Do NOT promote:
- Entries that are highly specific to one project's data or naming
- Entries that are already in the top-level memory
- Entries containing PII, credentials, or sensitive system details

## Steps

1. Identify any project-level `lessons.md` files in the current working directory or referenced in active context.

2. Read the top-level profile `memory/lessons.md` to know what's already there.

3. For each candidate entry in project lessons.md:
   - Assess whether it meets the promotion criteria above
   - If yes, add it to the top-level `memory/lessons.md` with a source tag: `*(promoted from [project-name], [date])*`
   - If no, skip it

4. Report: how many entries were reviewed, how many promoted, a brief summary of what was promoted.

5. The sync-profile skill will commit and push the updated memory files.

## Format for promoted entries

```markdown
**(YYYY-MM-DD) [Entry title]** *(promoted from [project], [date])*
- [bullet points as in original entry]
```
