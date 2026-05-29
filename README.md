# envfuse

`envfuse` is a Go CLI for Git-native secret and asset synchronization using [`filippo.io/age`](https://filippo.io/age) X25519 encryption.

It now supports **single-command Workspace Synchronization** with a repository blueprint (`envfuse.yaml`).

## Core capabilities

- Multi-recipient encryption (`envfuse encrypt`)
- Private-key decryption (`envfuse decrypt`)
- Native keypair generation (`envfuse keys generate`)
- Runtime manifest diagnostics (`envfuse manifest`)
- Workspace push/pull synchronization (`envfuse sync`)

## Workspace blueprint (`envfuse.yaml`)

Place `envfuse.yaml` at the root of your repository:

```yaml
team:
  - name: alice
    recipient: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  - name: bob
    recipient: age1yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy

manifest:
  - .env.local
  - certs/localhost.crt
  - certs/localhost.key
```

### Schema

- `team`: list of developer entries with:
  - `name`: developer label
  - `recipient`: public X25519 recipient key (`age1...`)
- `manifest`: list of repository-relative file or directory targets to synchronize

## Workspace synchronization

```bash
envfuse sync [--config <file>] [--plain]
```

Default config path: `./envfuse.yaml`

### Push directive (encryption)

When a tracked plaintext target is newer than its store bundle:

1. `envfuse sync` reads plaintext from the workspace.
2. Encrypts for all `team` recipients.
3. Writes an atomic encrypted bundle under:
   - `.envfuse/store/<relative-path>.enc`

### Pull directive (decryption)

When a store bundle is newer than local plaintext (or plaintext is missing):

1. `envfuse sync` loads local identity key from:
   - `$ENVFUSE_CONFIG_DIR/identity.key` (if set), otherwise
   - platform user config directory (`.../envfuse/identity.key`)
2. Decrypts matching `.envfuse/store/*.enc` bundle.
3. Atomically restores plaintext to its workspace-relative target path.

### Terminal status output

Each sync prints a scannable status line per file:

- `→` encrypted to store (push)
- `✓` decrypted to workspace (pull)
- `•` up to date
- `❌` error

Use `--plain` for ASCII-only markers (`[PUSH]`, `[PULL]`, `[OK]`, `[ERROR]`).

Example:

```text
→ .env.local (encrypted to workspace store)
✓ certs/localhost.crt (decrypted from workspace store)
• certs/localhost.key (up to date)
sync summary: pushed=1 pulled=1 unchanged=1 failed=0
```

## Key management

Generate local identity and recipient outputs:

```bash
envfuse keys generate [--identity-out <identity.key>] [--recipients-out <recipients.txt>] [--force]
```

Defaults:

- identity key: `<config-dir>/identity.key`
- recipients list: `<config-dir>/recipients.txt`

## Direct encrypt/decrypt commands

Encrypt for one or more recipients:

```bash
envfuse encrypt <target-file> --keys <recipients-file> [--out <encrypted-output>]
```

Decrypt with a local identity key:

```bash
envfuse decrypt <encrypted-file> [--identity <identity-file>] [--out <decrypted-output>] [--config-dir <dir>]
```

Generate runtime manifest JSON:

```bash
envfuse manifest [--out <manifest.json>] [--config-dir <dir>]
```

## Cross-platform notes

- All paths are composed with `filepath.Join` for Windows/Linux/macOS compatibility.
- Identity keys and parsed config scalars are trimmed for surrounding whitespace and carriage returns (`\r`).
- Atomic writes are used for encrypted and decrypted outputs.

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build ./...
```
