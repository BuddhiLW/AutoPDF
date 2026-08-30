# AutoPDF skills for Claude Code

Five skills that teach a coding agent how to build on AutoPDF. They load on
demand, so a project that only generates PDFs never pays for the preview
documentation.

| Skill | Covers |
| --- | --- |
| `autopdf` | Install, CLI, YAML config, choosing an integration path |
| `autopdf-embed` | Go library: `api.Engine`, testing without LaTeX, logging, v1→v2 |
| `autopdf-templates` | `delim[[ ]]` syntax, nested values, loops, escaping |
| `autopdf-preview` | Component documents, preview sessions, SSE/WebSocket/HTTP2 |
| `autopdf-plato` | Slides from Markdown/Org via plato, Beamer, adding a render target |

## Install

As a plugin, which keeps them updated with the repository:

```
/plugin marketplace add BuddhiLW/AutoPDF
/plugin install autopdf@autopdf
```

Or copy them into a single project:

```bash
git clone --depth 1 https://github.com/BuddhiLW/AutoPDF /tmp/autopdf
cp -r /tmp/autopdf/skills/autopdf* .claude/skills/
```

`~/.claude/skills/` instead of `.claude/skills/` makes them available in every
project on the machine.

## Verifying

Every Go snippet in these skills is compiled against the real `pkg/api` surface
before release, so the signatures shown are the ones that exist. When the API
changes, the snippets are re-checked rather than reviewed by eye.

## Contributing

Skills live beside the code they document so they can be updated in the same
commit as an API change. Keep each one to what an agent needs in order to act:
the shape of the call, the failure it will hit, and the thing that is surprising.
Reference material that an agent can look up belongs in `docs/`.
