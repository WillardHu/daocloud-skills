# daocloud-skills

A generated CLI and AI skill package for DaoCloud Enterprise (DCE). It wraps the DCE REST API into a `dcectl` command-line tool and bundles a companion skill for AI agents.

## Overview

- **`dcectl`** — a CLI generated from DCE OpenAPI specs (ghippo, kpanda). Supports API discovery, search, and execution with built-in auth management.
- **`skills/dcectl`** — an AI agent skill that teaches agents how to use `dcectl` safely.

## Prerequisites

- Go 1.25+
- Git (for spec sync)

## Quick Start

```bash
# Sync OpenAPI specs and regenerate code
make bootstrap

# Build dcectl
make build

# Install dcectl and symlink skill for local development
make dev
```

The `make dev` target installs `dcectl` to `/usr/local/bin` and symlinks `skills/dcectl` into `~/.agents/skills/dcectl` for live use in an AI agent runtime.

## Usage

```bash
# Log in to a DCE instance
dcectl auth login --hostname https://<dce-host>

# Browse available commands
dcectl commands --json

# Search for a command by intent
dcectl search "list clusters" --json

# Inspect a command before executing
dcectl commands show kpanda clusters list --json

# Execute a command
dcectl kpanda clusters list -o json
```

## Development

| Target | Description |
|---|---|
| `make bootstrap` | Sync specs + regenerate all code |
| `make specsync` | Pull latest OpenAPI specs from upstream |
| `make codegen` | Regenerate Go code and skill references from specs |
| `make sync-one SOURCE=<name>` | Sync and regenerate a single source (e.g. `ghippo`) |
| `make build` | Build `bin/dcectl` |
| `make dev` | Build, install, and symlink skill for local debugging |
| `make dev-clean` | Remove installed binary and skill symlink |
| `make clean` | Remove `.cache` and `bin` |

## Project Structure

```
.
├── cli.yaml                  # CLI name and auth config
├── specs/sources.yaml        # OpenAPI source definitions (pinned commits)
├── internal/
│   ├── generated/            # Generated Go command modules
│   └── overlay/              # Per-source field overrides for codegen
├── skills/dcectl/            # AI agent skill (SKILL.md + references)
├── cmd/dcectl/main.go        # CLI entrypoint
└── doc.go                    # Embeds cli.yaml for use by main
```

## Adding a New API Source

1. Add an entry to `specs/sources.yaml` with the repo URL, pinned commit, and OpenAPI file path.
2. Add an overlay file to `internal/overlay/<source>.yaml` if field customisation is needed.
3. Run `make bootstrap` to sync and regenerate.

## Updating a Pinned Spec

Update the `pinned_tag` for the relevant source in `specs/sources.yaml`, then run `make sync-one SOURCE=<name>` or `make bootstrap`.
