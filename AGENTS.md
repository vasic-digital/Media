# AGENTS.md

Instructions for AI agents working on this codebase.

## General Rules

- Follow Go conventions: `gofmt`, `go vet`, idiomatic error handling.
- All public types and functions must have doc comments.
- Use table-driven tests with `testify/assert` and `testify/require`.
- Do not add external dependencies without explicit approval. The only allowed test dependency is `github.com/stretchr/testify`.
- Do not create files outside the established package structure (`pkg/detector`, `pkg/analyzer`, `pkg/provider`, `pkg/metadata`).
- Run `go build ./...` and `go test ./... -count=1` before considering any change complete.

## Package Guidelines

### pkg/detector
- Detection rules must have a `Priority` field. Higher values take precedence.
- TV show detection (S01E02 patterns) must have higher priority than generic video extension matching.
- All extension matching must be case-insensitive.
- New media types should be added as constants to the `MediaType` type.

### pkg/analyzer
- The `Analyzer` interface must be implemented by any new analyzer.
- `FilenameAnalyzer` delegates detection to `detector.Engine` and enriches the result.
- Analyzers must not perform I/O. Filename-based analysis only.

### pkg/provider
- The `Provider` interface is the contract for all external metadata sources.
- `Registry` must remain thread-safe (uses `sync.RWMutex`).
- `Registry.Search` must query providers concurrently and handle partial failures gracefully.

### pkg/metadata
- Shared types used across packages belong here.
- Helper functions must be pure (no side effects, no I/O).

## Testing

- Every public function must have at least one test.
- Use subtests (`t.Run`) for table-driven tests.
- Mock implementations of interfaces should be defined in `_test.go` files, not in production code.
- Tests must pass with `go test ./... -count=1` (no test caching).


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

