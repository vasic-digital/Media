# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`digital.vasic.media` is a standalone Go module providing media file type detection, metadata extraction, and external metadata provider interfaces. It is designed to be imported by other projects (such as Catalogizer) that need to detect, categorize, and enrich media files.

## Module Structure

| Package | Purpose |
|---|---|
| `pkg/detector` | Media type detection from filenames and file extensions |
| `pkg/analyzer` | Media metadata extraction and analysis from file paths |
| `pkg/provider` | External metadata provider interfaces (TMDB, IMDB, etc.) |
| `pkg/metadata` | Common metadata types, quality info, and helper functions |

## Commands

```bash
# Run all tests
go test ./... -count=1

# Run tests with verbose output
go test -v ./... -count=1

# Run tests for a specific package
go test -v ./pkg/detector/ -count=1
go test -v ./pkg/analyzer/ -count=1
go test -v ./pkg/provider/ -count=1
go test -v ./pkg/metadata/ -count=1

# Run a single test
go test -v -run TestDetect_TVShow_S01E02 ./pkg/detector/

# Build all packages
go build ./...

# Check for issues
go vet ./...
```

## Architecture

### Detection Pipeline

`detector.Engine` uses a rule-based system with priorities:
1. TV Show patterns (S01E02, 1x01) have highest priority (100)
2. Extension-based rules (video, audio, image, book, software) have standard priority (50)
3. Custom rules can be added with `Engine.AddRule()`

### Analyzer

`analyzer.FilenameAnalyzer` wraps the detection engine and extracts structured metadata:
- Title, year, resolution, codec from filenames
- Artist/author from "Artist - Title" patterns
- Season/episode from TV show patterns

### Provider Registry

`provider.Registry` manages external metadata providers:
- Thread-safe registration and lookup
- Concurrent search across all providers
- Graceful error handling (partial failures do not block results)

### Metadata Helpers

`metadata` package provides shared types and utilities:
- `QualityInfo` with comparison and display methods
- `Resolution` with human-readable formatting
- File size formatting, title normalization, filename sanitization

## Conventions

- **Go**: Standard Go project layout under `pkg/`
- **Testing**: Table-driven tests with `testify/assert` and `testify/require`
- **Naming**: `NewXxx` constructor functions, interface-based design
- **Errors**: Wrapped errors with `fmt.Errorf("context: %w", err)`
- **Concurrency**: `sync.RWMutex` for thread-safe registry access


## ⚠️ MANDATORY: NO SUDO OR ROOT EXECUTION

**ALL operations MUST run at local user level ONLY.**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

### Why This Matters
- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

### When You See SUDO
If any script or command suggests using `sudo`:
1. STOP immediately
2. Find a user-level alternative
3. Use rootless container runtimes
4. Modify commands to work within user permissions

**VIOLATION OF THIS CONSTRAINT IS STRICTLY PROHIBITED.**

