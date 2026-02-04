# Usage Guide

## Alternative Installation Methods

### Using Go

```bash
go install github.com/franzwilhelm/linc@latest
```

### From Source

```bash
git clone https://github.com/franzwilhelm/linc.git
cd linc
go build -o linc .
mv linc /usr/local/bin/  # or anywhere in your PATH
```

## Navigation

| Key | Action |
|-----|--------|
| `h` / `l` | Switch between status tabs |
| `j` / `k` | Navigate issues |
| `/` | Filter issues |
| `Enter` | Select issue / confirm |
| `o` | Open issue in browser |
| `s` | Start working on issue |
| `Esc` | Go back |
| `q` | Quit |

## Workflow

1. **Select an issue** - Browse by status, filter if needed
2. **View details** - See full description, labels, assignee
3. **Start working** - Optionally add a comment, toggle branch creation
4. **Launch agent** - Issue moves to "In Progress", comment syncs, branch created, agent starts

## Configuration

Config is stored at `~/.linc/config.json`:

```json
{
  "workspaces": [
    {
      "id": "workspace-uuid",
      "name": "My Workspace",
      "apiKey": "lin_api_xxx",
      "defaultTeamId": "team-uuid"
    }
  ],
  "directories": {
    "/path/to/project": "workspace-uuid"
  },
  "provider": "claude"
}
```

## Providers

linc is AI CLI agnostic and supports multiple coding agents:

| Provider | Status | Description |
|----------|--------|-------------|
| `claude` | ✅ Ready | Claude Code |
| `echo` | ✅ Ready | Prints prompt to terminal (for piping/debugging) |
| `opencode` | 🚧 Planned | [opencode](https://github.com/sst/opencode) support coming soon |

To change provider, edit `~/.linc/config.json` and set `"provider": "claude"` (or your preferred provider).

## Adding New Providers

Providers implement a simple interface:

```go
type Provider interface {
    Name() string
    Exec(issue linear.Issue, comment string, ctx *linear.IssueContext, planMode bool) error
}
```

See `internal/provider/claude/claude.go` for an example implementation.

## Requirements

- Go 1.21+ (for building from source)
- A Linear account with API access
- An AI coding agent (Claude Code, etc.) installed and in PATH
