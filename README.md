# Airborne (`ab`)

MVP developer CLI over the UNICAGD intelligence kernel.

## Build

```sh
make build
```

## Usage

```sh
# Verify kernel integrity (Ed25519 + SHA-256)
AB_KERNEL_DIR="$PWD/kernel" ./ab verify

# Show version (no model or kernel needed)
./ab version

# Send a query to the model
AB_KERNEL_DIR="$PWD/kernel" OPENAI_API_KEY=... ./ab "what does this repository do?"

# Review a git diff from stdin
git diff | AB_KERNEL_DIR="$PWD/kernel" OPENAI_API_KEY=... ./ab review

# Analyze source files (JSON output)
AB_KERNEL_DIR="$PWD/kernel" ./ab analyze src/ --json | jq .

# Inspect the execution plan for a query (no model connection)
AB_KERNEL_DIR="$PWD/kernel" ./ab inspect plan "find architectural risks" --json

# Show the last run audit record
AB_STATE_DIR="$PWD/.state" ./ab inspect last
```

## Configuration (environment)

| Variable          | Description                  | Default                  |
|-------------------|------------------------------|--------------------------|
| `OPENAI_API_KEY`  | OpenAI API key               | (none; plan shown only)  |
| `AB_MODEL`        | Model to use                 | `gpt-4.1-mini`           |
| `AB_OPENAI_BASE_URL` | OpenAI-compatible API URL | `https://api.openai.com/v1` |
| `AB_KERNEL_DIR`   | Path to kernel directory     | (required for verify/query) |
| `AB_STATE_DIR`    | State directory for audit    | `~/.local/state/ab`      |

## Exit codes

| Code | Meaning                  |
|------|--------------------------|
| 0    | success                  |
| 1    | execution failure        |
| 2    | invalid invocation       |
| 3    | verification failure     |
| 4    | provider failure         |
| 5    | policy/context failure   |

## Invariants

1. **Integrity before intelligence.** Kernel artifacts are verified (manifest SHA-256, artifact SHA-256/size, crypto-manifest digest, Ed25519 signature, key ID) before any model request is constructed. Fail closed with exit 3.
2. **Kernel boundaries.** Only task-relevant kernel sections (governance, architecture, ontology) are selected. The full kernel is never concatenated into the model request.
3. **Bounded, hostile-by-default context.** stdin, git diff, and bounded text files are collected; binary and generated/vendor trees are skipped.
4. **READ/ANALYZE only.** No model-controlled shell, filesystem mutation, network action, commit, or deployment capability.
5. **Auditable without secrets.** Run ID, verification state, context count/truncation, duration, status, and plan are persisted. No API keys or raw project context.

## Kernel directory structure

```
kernel/
  manifest.json            # Artifact list, key_id, crypto_manifest_digest
  crypto_manifest.json     # Canonical signed document
  UNICAGD_289X.proto       # Proto definition (package UNICAGD_288X)
  UNICAGD_289X.json        # Canonical JSON runtime spec
  UNICAGD_289X.pb          # Binary protobuf (UNICAGD_RUNTIME_PB_v2_ENCRYPTED)
  signature.json           # Ed25519 signature + manifest_sha256 + key_id
  public_key.pem           # Ed25519 public key
```
