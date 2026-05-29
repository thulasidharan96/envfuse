# envfuse

`envfuse` is a Go CLI package/application for encrypting and decrypting `.env` files (and other binary/text assets) using [`filippo.io/age`](https://filippo.io/age) X25519 keys.

It is designed for Git-native secret distribution:
- commit only encrypted artifacts (`*.age`)
- keep private identity keys local
- decrypt only on trusted developer/runner machines

## Why this approach

- **No centralized secret manager required** for basic team workflows.
- **Asymmetric crypto model**: encryption uses public recipient keys; decryption requires a private identity key.
- **Simple CI/dev automation**: encrypted files can safely live in repository history while private keys stay outside Git.

## How it works

### High-level flow

1. Team members generate local age identity keys (`AGE-SECRET-KEY-...`) and share only their matching recipient public keys (`age1...`).
2. A source file (for example `app.env`) is encrypted with one or more recipient public keys.
3. `envfuse` writes encrypted output (`.age`) with strict `0600` file permissions.
4. Any holder of a matching private identity key can decrypt the `.age` file back to plaintext.
5. Decrypted output is written with `0600` permissions; sensitive buffers are zeroed where practical.

### Internal architecture

- **CLI layer (`cmd/`)**: Cobra commands, flags, argument validation.
- **Crypto layer (`pkg/crypto`)**: thin wrappers over `age.Encrypt` and `age.Decrypt`.
- **Root config resolution (`cmd/root.go`)**:
  - default base path = `os.UserConfigDir()/envfuse`
  - override supported with `--config-dir`

Platform default config directory:
- **Linux/macOS**: `~/.config/envfuse`
- **Windows**: `%AppData%\\Roaming\\envfuse`

## Runtime prerequisites

- Go 1.22+ (for building from source)
- age-compatible X25519 key material

### Required key files

#### 1) Recipient list file (for `encrypt --keys`)

Text file containing one recipient public key per line:

```text
# comments are allowed
age1q...
age1x...
```

Rules:
- blank lines are ignored
- lines starting with `#` are ignored
- at least one valid recipient key is required

#### 2) Identity file (for `decrypt --identity` or default path)

Text file containing a single private identity key:

```text
AGE-SECRET-KEY-1...
```

Rules:
- content is trimmed
- empty content fails decryption

## Command behavior and defaults

### Encrypt command

```bash
envfuse encrypt <target-file> --keys <recipients-file> [--out <encrypted-output>]
```

What happens:
1. Reads `<target-file>`.
2. Parses recipient keys from `--keys` file.
3. Encrypts bytes for all recipients.
4. Writes output with mode `0600`.

Defaults:
- output path defaults to `<target-file>.age` when `--out` is omitted.

### Decrypt command

```bash
envfuse decrypt <encrypted-file> [--identity <identity-file>] [--out <decrypted-output>] [--config-dir <dir>]
```

What happens:
1. Resolves config directory (`--config-dir` or platform default).
2. Resolves identity path:
   - `--identity` if provided
   - otherwise `<config-dir>/identity.key`
3. Reads encrypted target file.
4. Decrypts using the identity key.
5. Writes plaintext with mode `0600`.

Defaults:
- identity path: `<config-dir>/identity.key`
- output path: `<config-dir>/decrypted.env`

## Local development

### Setup

```bash
go mod tidy
go test ./...
```

### Build

```bash
go build ./...
```

## Release as a cross-platform CLI tool

This repository is already configured to release `envfuse` as a CLI binary for Linux, macOS, and Windows via GoReleaser (`.goreleaser.yaml`).

Release matrix:
- **GOOS**: `linux`, `darwin` (macOS), `windows`
- **GOARCH**: `amd64`, `arm64`
- **binary name**: `envfuse`

Archive formats:
- `tar.gz` for Linux/macOS
- `zip` for Windows

Before building release artifacts, GoReleaser runs:
- `go mod tidy`
- `go test ./...`

CI/CD delivery:
- GitHub Actions workflow: `.github/workflows/release.yml`
- automatic release on pushed tags matching `v*` (for example `v1.0.0`)
- manual release via `workflow_dispatch`
- artifacts are uploaded to GitHub Releases using GoReleaser

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
