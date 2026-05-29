# envfuse

`envfuse` is a zero-dependency, high-velocity CLI for securely syncing development configuration, `.env` targets, and binary assets in a Git-native workflow.

It uses [`filippo.io/age`](https://filippo.io/age) asymmetric encryption with X25519 recipients to avoid centralized secret stores while keeping local workflows simple.

## Architecture

- **Command layer**: `github.com/spf13/cobra` (`cmd/`) for consistent CLI UX and error handling.
- **Crypto layer**: `pkg/crypto` wraps `age` for multi-recipient encryption and single-identity decryption.
- **Entry point**: `main.go` delegates execution to the root command.
- **Config path model**: uses `os.UserConfigDir()` + `filepath.Join(..., "envfuse")`, resulting in:
  - Windows: `%AppData%\\Roaming\\envfuse`
  - macOS/Linux: `~/.config/envfuse`

## Security model

- No custom cryptography; only `filippo.io/age` wrappers are used.
- Secret-bearing files are written with strict `0600` permissions.
- In-memory secret byte buffers are zeroed after use where practical.

## Local development

### Prerequisites

- Go 1.22+

### Setup

```bash
go mod tidy
go test ./...
```

### Build

```bash
go build ./...
```

## Usage

### Encrypt

```bash
envfuse encrypt ./app.env --keys ./recipients.txt --out ./app.env.age
```

`recipients.txt` expects one `age1...` public recipient key per line.

### Decrypt

```bash
envfuse decrypt ./app.env.age --identity ~/.config/envfuse/identity.key --out ./app.env
```

If flags are omitted:

- `--identity` defaults to `<user-config-dir>/envfuse/identity.key`
- `--out` defaults to `<user-config-dir>/envfuse/decrypted.env`

Use `--config-dir` to override the base envfuse config directory.

## Releases

Releases are configured via `.goreleaser.yaml` for:

- GOOS: `linux`, `darwin`, `windows`
- GOARCH: `amd64`, `arm64`

## Contributing

1. Fork and create a feature branch.
2. Keep changes small and covered with tests.
3. Run `go test ./...` before opening a PR.
4. Use **Conventional Commits**:
   - `feat: ...`
   - `fix: ...`
   - `docs: ...`
   - `test: ...`
   - `chore: ...`

## License

MIT (or your repository default license terms).
